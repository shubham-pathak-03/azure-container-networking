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
	errorsToRetryOn        []*ErrorDefinition
	lineFailureDefinitions []*ErrorDefinition
	retryCount             int // TODO rename these to try counts
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

// ErrorDefinition defines an error by a regular expression and its error code.
type ErrorDefinition struct {
	matchPattern string
	re           *regexp.Regexp
}

// LineErrorHandler defines an error and how to handle it
type LineErrorHandler struct {
	Definition *ErrorDefinition
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
		errorsToRetryOn:        make([]*ErrorDefinition, 0),
		lineFailureDefinitions: make([]*ErrorDefinition, len(lineFailurePatterns)),
		retryCount:             0,
		maxRetryCount:          maxRetryCount,
		exec:                   exec,
	}
	for k, lineFailurePattern := range lineFailurePatterns {
		creator.lineFailureDefinitions[k] = NewErrorDefinition(lineFailurePattern)
	}
	return creator
}

func NewErrorDefinition(pattern string) *ErrorDefinition {
	return &ErrorDefinition{
		matchPattern: pattern,
		re:           regexp.MustCompile(pattern),
	}
}

func (creator *FileCreator) AddErrorToRetryOn(definition *ErrorDefinition) {
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
	fileString := creator.ToString()
	wasFileAltered, err := creator.runCommandOnceWithFile(fileString, cmd, args...)
	if err == nil {
		return nil
	}
	for {
		if creator.hasNoMoreRetries() {
			return fmt.Errorf("%w", err) // using fmt.Errorf to appease go lint
		}
		commandString := cmd + " " + strings.Join(args, " ")
		if wasFileAltered {
			fileString = creator.ToString()
			log.Logf("rerunning command [%s] with new file:\n%s", commandString, fileString)
		} else {
			log.Logf("rerunning command [%s] with the same file")
		}
		wasFileAltered, err = creator.runCommandOnceWithFile(fileString, cmd, args...)
	}
}

// RunCommandOnceWithFile runs the command with the file once and increments the retry count.
// It returns whether the file was altered and any error.
// For automatic retrying and proper logging, use RunCommandWithFile. This method was designed for external testing.
func (creator *FileCreator) RunCommandOnceWithFile(cmd string, args ...string) (bool, error) {
	commandString := cmd + " " + strings.Join(args, " ")
	if creator.hasNoMoreRetries() {
		return false, fmt.Errorf("reached max retry count %d for command [%s]", creator.retryCount, commandString)
	}
	fileString := creator.ToString()
	return creator.runCommandOnceWithFile(fileString, cmd, args...)
}

func (creator *FileCreator) runCommandOnceWithFile(fileString, cmd string, args ...string) (bool, error) {
	command := creator.exec.Command(cmd, args...)
	command.SetStdin(bytes.NewBufferString(fileString))

	stdErrBytes, err := command.CombinedOutput()
	if err == nil {
		return false, nil // success
	}
	creator.retryCount++

	commandString := cmd + " " + strings.Join(args, " ")
	stdErr := string(stdErrBytes)
	log.Errorf("on try number %d, failed to run command [%s] with error [%v] and stdErr [%s]. Used file:\n%s", creator.retryCount, commandString, err, stdErr, fileString)
	if creator.hasNoMoreRetries() {
		return false, fmt.Errorf("after %d tries, failed to run command [%s] with final error [%w] and stdErr [%s]", creator.retryCount, commandString, err, stdErr)
	}

	// begin the retry logic
	creator.retryCount++
	if creator.hasFileLevelError(stdErr) {
		log.Logf("detected a file-level error", commandString)
		return false, fmt.Errorf("%w", err) // using fmt.Errorf to appease go lint
	}

	// no file-level error, so handle line-level error if there is one
	lineNum := creator.getErrorLineNumber(commandString, stdErr)
	if lineNum == -1 {
		// can't detect a line number error
		return false, fmt.Errorf("%w", err) // using fmt.Errorf to appease go lint
	}
	wasFileAltered := creator.handleLineError(lineNum, commandString, stdErr)
	return wasFileAltered, fmt.Errorf("%w", err) // using fmt.Errorf to appease go lint
}

func (creator *FileCreator) hasNoMoreRetries() bool {
	return creator.retryCount >= creator.maxRetryCount
}

func (creator *FileCreator) hasFileLevelError(stdErr string) bool {
	for _, errorDefinition := range creator.errorsToRetryOn {
		if errorDefinition.isMatch(stdErr) {
			return true
		}
	}
	return false
}

func (definition *ErrorDefinition) isMatch(stdErr string) bool {
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
			log.Logf("expected error with line number, but error regex pattern %s didn't produce a number for command [%s] with stdErr [%s]", definition.matchPattern, commandString, stdErr)
			continue
		}
		if lineNum < 1 || lineNum > len(creator.lines) {
			log.Logf("expected error with line number, but error regex pattern %s produced an invalid line number %d for command [%s] with stdErr [%s]", definition.matchPattern, lineNum, commandString, stdErr)
			continue
		}
		return lineNum
	}
	return -1
}

// return whether the file was altered
func (creator *FileCreator) handleLineError(lineNum int, commandString, stdErr string) bool {
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
			return true
		case AbortSection:
			log.Errorf("aborting section associated with line %d for command [%s]", lineNumIndex, commandString)
			section, exists := creator.sections[line.sectionID]
			if !exists {
				log.Errorf("can't abort section because line references section %d which doesn't exist, so skipping the line instead", line.sectionID)
				creator.lineNumbersToOmit[lineNumIndex] = struct{}{}
			} else {
				for _, lineNum := range section.lineNums {
					creator.lineNumbersToOmit[lineNum] = struct{}{}
				}
			}
		}
		errorHandler.Callback()
		return true
	}
	log.Logf("no error handler for command [%s] with stdErr [%s]", commandString, stdErr)
	return false
}