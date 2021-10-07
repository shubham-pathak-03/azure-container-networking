package ioutil

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Azure/azure-container-networking/log"

	kexec "k8s.io/utils/exec"
)

// TODO add file creator log prefix

// FileCreator is a tool for:
// - building a buffer file
// - running a command with the file
// - handling errors in the file
type FileCreator struct {
	lines                  []*Line
	sections               map[string]*Section // key is sectionID
	lineNumbersToOmit      map[int]struct{}
	errorsToRetryOn        []*errorDefinition
	lineFailureDefinitions []*errorDefinition
	retryCount             int
	maxRetryCount          int
	exec                   kexec.Interface
}

// TODO for iptables:
// lineFailurePattern := "line (\\d+) failed"
// AND "Error occurred at line: (\\d+)"

// Line defines the content, section, and error handlers for a line
type Line struct {
	content       string
	sectionID     string
	errorHandlers []*LineErrorHandler
}

// Section is a logically connected components (not necessarily adjacent lines)
type Section struct {
	id       string
	lineNums []int
}

// errorDefinition defines an error by a regular expression and its error code.
type errorDefinition struct {
	matchPattern string
	re           *regexp.Regexp
}

// LineErrorHandler defines an error and how to handle it
type LineErrorHandler struct {
	Definition *errorDefinition
	Method     LineErrorHandlerMethod
	Reason     string
	Callback   func()
}

// LineErrorHandlerMethod defines behavior when an error occurs
type LineErrorHandlerMethod string

// possible LineErrorHandlerMethod
const (
	SkipLine     LineErrorHandlerMethod = "skip"
	AbortSection LineErrorHandlerMethod = "abort"
)

func NewFileCreator(maxRetryCount int, exec kexec.Interface, lineFailurePatterns ...string) *FileCreator {
	creator := &FileCreator{
		lines:                  make([]*Line, 0),
		sections:               make(map[string]*Section),
		lineNumbersToOmit:      make(map[int]struct{}),
		errorsToRetryOn:        make([]*errorDefinition, 0),
		lineFailureDefinitions: make([]*errorDefinition, len(lineFailurePatterns)),
		retryCount:             0,
		maxRetryCount:          maxRetryCount,
		exec:                   exec,
	}
	for k, lineFailurePattern := range lineFailurePatterns {
		creator.lineFailureDefinitions[k] = NewErrorDefinition(lineFailurePattern)
	}
	return creator
}

func NewErrorDefinition(pattern string) *errorDefinition {
	return &errorDefinition{
		matchPattern: pattern,
		re:           regexp.MustCompile(pattern),
	}
}

func (creator *FileCreator) AddErrorToRetryOn(definition *errorDefinition) {
	creator.errorsToRetryOn = append(creator.errorsToRetryOn, definition)
}

func (creator *FileCreator) AddLine(sectionID string, errorHandlers []*LineErrorHandler, items ...string) {
	section, exists := creator.sections[sectionID]
	if !exists {
		section = &Section{sectionID, make([]int, 0)}
		creator.sections[sectionID] = section
	}
	spaceSeparatedItems := strings.Join(items, " ")
	line := &Line{spaceSeparatedItems, sectionID, errorHandlers}
	creator.lines = append(creator.lines, line)
	section.lineNums = append(section.lineNums, len(creator.lines)-1)
}

// ToString combines the lines in the FileCreator and ends with a new line.
func (creator *FileCreator) ToString() string {
	result := strings.Builder{}
	for lineNum, line := range creator.lines {
		_, isOmitted := creator.lineNumbersToOmit[lineNum]
		if !isOmitted {
			result.WriteString(line.content + "\n")
		}
	}
	return result.String()
}

func (creator *FileCreator) RunCommandWithFile(cmd string, args ...string) error {
	commandString := cmd + " " + strings.Join(args, " ")
	fileString := creator.ToString()
	for {
		command := creator.exec.Command(cmd, args...)
		command.SetStdin(bytes.NewBufferString(fileString))
		// stdErrBuffer := bytes.NewBuffer(nil)
		// command.SetStderr(stdErrBuffer)

		stdErrBytes, err := command.CombinedOutput()
		if err == nil {
			return nil // success
		}

		stdErr := string(stdErrBytes)
		// stdErr := stdErrBuffer.String()
		log.Errorf("on try number %d, failed to run command [%s] with error [%w] and stdErr [%s]. Used file:\n%s", creator.retryCount, commandString, err, stdErr, fileString)
		wasLastTry := creator.retryCount >= creator.maxRetryCount
		if wasLastTry {
			return fmt.Errorf("after %d tries, failed to run command [%s] with final error [%w] and stdErr [%s]", creator.retryCount, commandString, err, stdErr)
		}

		// begin the retry logic
		creator.retryCount++
		if creator.hasFileLevelError(stdErr) {
			log.Logf("detected a file-level error: retrying command [%s] with same file", commandString)
			continue // retry
		}

		// no file-level error, so handle line-level error if there is one
		lineNum := creator.getErrorLineNumber(commandString, stdErr)
		if lineNum == -1 {
			// can't detect a line number error
			log.Logf("unknown error: retrying command [%s] with same file", commandString)
			continue // retry
		}
		creator.handleLineError(lineNum, commandString, stdErr)
		fileString := creator.ToString()
		log.Logf("rerunning command [%s] with new file:\n%s", commandString, fileString)
	}
}

func (creator *FileCreator) hasFileLevelError(stdErr string) bool {
	for _, errorDefinition := range creator.errorsToRetryOn {
		if errorDefinition.isMatch(stdErr) {
			return true
		}
	}
	return false
}

func (definition *errorDefinition) isMatch(stdErr string) bool {
	return definition.re.MatchString(stdErr)
}

// return -1 if there's a failure
func (creator *FileCreator) getErrorLineNumber(commandString, stdErr string) int {
	for _, definition := range creator.lineFailureDefinitions {
		result := definition.re.FindStringSubmatch(stdErr)
		if result == nil || len(result) < 2 {
			log.Logf("expected error with line number, but couldn't detect one with error regex pattern [%s] for command [%s] with stdErr [%s]", definition.matchPattern, commandString, stdErr)
			continue
		}
		lineNumString := result[1]
		lineNum, err := strconv.Atoi(lineNumString)
		if err != nil {
			log.Logf("error regex pattern %s didn't produce a number for command [%s] with stdErr [%s]", definition.matchPattern, commandString, stdErr)
			continue
		}
		if lineNum < 1 || lineNum > len(creator.lines) {
			log.Logf("error regex pattern %s produced invalid line number %d for command [%s] with stdErr [%s]", definition.matchPattern, lineNum, commandString, stdErr)
			continue
		}
		return lineNum
	}
	return -1
}

func (creator *FileCreator) handleLineError(lineNum int, commandString, stdErr string) {
	lineNumIndex := lineNum - 1
	line := creator.lines[lineNumIndex]
	for _, errorHandler := range line.errorHandlers {
		if !errorHandler.Definition.isMatch(stdErr) {
			continue
		}
		switch errorHandler.Method {
		case SkipLine:
			log.Errorf("skipping line %d for command [%s]", lineNumIndex, commandString)
			creator.lineNumbersToOmit[lineNumIndex] = struct{}{}
			errorHandler.Callback()
			return
		case AbortSection:
			log.Errorf("aborting section associated with line %d for command [%s]", lineNumIndex, commandString)
			section, exists := creator.sections[line.sectionID]
			if !exists {
				log.Errorf("line references section %d which doesn't exist", line.sectionID)
				return
			}
			for _, lineNum := range section.lineNums {
				creator.lineNumbersToOmit[lineNum] = struct{}{}
			}
			errorHandler.Callback()
			return
		}
	}
	log.Logf("no error handler for command [%s] with stdErr [%s]", commandString, stdErr)
}
