package ioutil

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Azure/azure-container-networking/log"
	kexec "k8s.io/utils/exec"
	// osexec "os/exec"
)

// FileCreator is a tool for:
// - building a buffer file
// - running a command with the file
// - handling errors in the file
type FileCreator struct {
	lines              []*Line
	sections           map[int]*Section // key is sectionID
	lineNumbersToOmit  map[int]struct{}
	errorsToRetryOn    []*ErrorDefinition
	lineFailurePattern string
	lineFailureRegex   *regexp.Regexp
	retryCount         int
	maxRetryCount      int
	exec               kexec.Interface
}

// TODO for iptables:
// lineFailurePattern := "line (\\d+) failed"
// AND "Error occurred at line: (\\d+)"

// Line defines the content, section, and error handlers for a line
type Line struct {
	content       string
	sectionID     int
	errorHandlers []*LineErrorHandler
}

// Section is a logically connected components (not necessarily adjacent lines)
type Section struct {
	id       int
	lineNums []int
}

// ErrorDefinition defines an error by a regular expression and its error code.
type ErrorDefinition struct {
	matchPattern string
	re           *regexp.Regexp
}

// LineErrorHandler defines an error and how to handle it
type LineErrorHandler struct {
	definition *ErrorDefinition
	method     LineErrorHandlerMethod
	reason     string
}

// LineErrorHandlerMethod defines behavior when an error occurs
type LineErrorHandlerMethod string

// possible LineErrorHandlerMethod
const (
	SkipLine                  LineErrorHandlerMethod = "skip"
	AbortSection              LineErrorHandlerMethod = "abort"
	RetryOnceThenSkipLine     LineErrorHandlerMethod = "retry then skip"
	RetryOnceThenAbortSection LineErrorHandlerMethod = "retry then abort"
)

func NewFileCreator(lineFailurePattern string, maxRetryCount int) *FileCreator {
	return &FileCreator{
		lines:              make([]*Line, 0),
		sections:           make(map[int]*Section),
		lineNumbersToOmit:  make(map[int]struct{}),
		errorsToRetryOn:    make([]*ErrorDefinition, 0),
		lineFailurePattern: lineFailurePattern,
		lineFailureRegex:   regexp.MustCompile(lineFailurePattern),
		retryCount:         0,
		maxRetryCount:      maxRetryCount,
		exec:               nil,
	}
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

func (creator *FileCreator) CreateSection(sectionID int) {
	creator.sections[sectionID] = &Section{sectionID, make([]int, 0)}
}

func (creator *FileCreator) AddLine(sectionID int, errorHandlers []*LineErrorHandler, items ...string) {
	spaceSeparatedItems := strings.Join(items, " ")
	line := &Line{spaceSeparatedItems, sectionID, errorHandlers}
	creator.lines = append(creator.lines, line)
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
	creator.assertExecExists()

	commandString := cmd + " " + strings.Join(args, " ")
	fileString := creator.ToString()
	for {
		command := creator.exec.Command(cmd, args...)
		command.SetStdin(bytes.NewBufferString(fileString))
		stdErrBuffer := bytes.NewBuffer(nil)
		command.SetStderr(stdErrBuffer)

		creator.retryCount++
		isLastTry := creator.retryCount >= creator.maxRetryCount
		err := command.Start()
		if err != nil {
			if isLastTry {
				return fmt.Errorf("after %d tries, couldn't start command [%s] with error [%w]", creator.retryCount, commandString, err)
			}
			continue // retry
		}
		err = command.Wait()
		if err == nil {
			return nil // success
		}

		stdErr := stdErrBuffer.String()
		log.Errorf("on try number %d, failed to run command [%s] with error [%w] and stdErr [%s]. Used file:\n%s", creator.retryCount, commandString, err, stdErr, fileString)
		if isLastTry {
			return fmt.Errorf("after %d tries, failed to run command [%s] with final error [%w] and stdErr [%s]", creator.retryCount, commandString, err, stdErr)
		}

		// retry if there was a known file-level error
		for _, errorDefinition := range creator.errorsToRetryOn {
			if errorDefinition.isMatch(stdErr) {
				log.Logf("retrying command [%s] with same file", commandString) // TODO include message about the error?
			}
			continue // retry
		}

		// handle line-level error if there is one
		lineNum := creator.getErrorLineNumber(commandString, stdErr)
		if lineNum == -1 {
			// can't detect a line number error
			log.Logf("retrying command [%s] with same file", commandString)
			continue // retry
		}
		creator.handleLineErrors(lineNum, commandString, stdErr)
		fileString := creator.ToString()
		log.Logf("rerunning command [%s] with new file:\n%s", commandString, fileString)
	}
}

func (creator *FileCreator) handleLineErrors(lineNum int, commandString, stdErr string) {
	line := creator.lines[lineNum]
	for _, errorHandler := range line.errorHandlers {
		if !errorHandler.definition.isMatch(stdErr) {
			continue
		}
		switch errorHandler.method {
		case RetryOnceThenSkipLine, RetryOnceThenAbortSection:
			log.Errorf("retrying line %d one more time for command [%s] for the following reason [%s]", lineNum, commandString, errorHandler.reason) // TODO change wording and possibly have a second error message when skip/abort happens?
			errorHandler.method = AbortSection
			if errorHandler.method == RetryOnceThenSkipLine {
				errorHandler.method = SkipLine
			}
			return
		case SkipLine:
			creator.lineNumbersToOmit[lineNum] = struct{}{}
			log.Errorf("skipping line %d for command [%s] for the following reason [%s]", lineNum, commandString, errorHandler.reason)
			return
		case AbortSection:
			creator.removeSection(line.sectionID)
			log.Errorf("aborting section associated with line %d for command [%s] for the following reason [%s]", lineNum, commandString, errorHandler.reason)
			return
		}
	}
	log.Logf("no error handler for command [%s] with stdErr [%s]", commandString, stdErr)
}

func (creator *FileCreator) assertExecExists() {
	if creator.exec == nil {
		creator.exec = kexec.New()
	}
}

func (definition *ErrorDefinition) isMatch(stdErr string) bool {
	return definition.re.MatchString(stdErr)
}

// return -1 if there's a failure
func (creator *FileCreator) getErrorLineNumber(commandString, stdErr string) int {
	result := creator.lineFailureRegex.FindStringSubmatch(stdErr)
	if result == nil || len(result) < 2 {
		// check length just to be safe even though we should always have either result == nil || len(result) == 2
		log.Errorf("expected error with line number, but couldn't detect one with error regex pattern %s for command [%s] with stdErr [%s]", creator.lineFailurePattern, commandString, stdErr)
		return -1
	}
	lineNumberString := result[1]
	lineNumber, err := strconv.Atoi(lineNumberString)
	if err != nil {
		log.Errorf("error regex pattern %s didn't produce a number for command [%s] with stdErr [%s]", creator.lineFailurePattern, commandString, stdErr)
		return -1
	}
	if lineNumber < 0 || lineNumber >= len(creator.lines) {
		log.Errorf("error regex pattern %s produced an invalid line number for command [%s] with stdErr [%s]", creator.lineFailurePattern, commandString, stdErr)
		return -1
	}
	return lineNumber
}

func (creator *FileCreator) removeSection(sectionID int) {
	section := creator.sections[sectionID]
	for _, lineNum := range section.lineNums {
		creator.lineNumbersToOmit[lineNum] = struct{}{}
	}
}
