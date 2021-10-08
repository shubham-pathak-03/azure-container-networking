package ioutil

import (
	"fmt"
	"testing"

	"github.com/Azure/azure-container-networking/log"
	testutils "github.com/Azure/azure-container-networking/test/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testCommandString = "test command"
	section1ID        = "section1"
	section2ID        = "section2"
)

var fakeSuccessCommand = testutils.TestCmd{
	Cmd:      []string{testCommandString},
	Stdout:   "success",
	ExitCode: 0,
}

func TestToStringAndSections(t *testing.T) {
	creator := NewFileCreator(0, nil)
	creator.AddLine(section1ID, nil, "line1-item1", "line1-item2", "line1-item3")
	creator.AddLine(section2ID, nil, "line2-item1", "line2-item2", "line2-item3")
	creator.AddLine(section1ID, nil, "line3-item1", "line3-item2", "line3-item3")

	section1 := creator.sections[section1ID]
	require.Equal(t, section1ID, section1.id)
	require.Equal(t, []int{0, 2}, section1.lineNums)

	section2 := creator.sections[section2ID]
	require.Equal(t, section2ID, section2.id)
	require.Equal(t, []int{1}, section2.lineNums)

	fileString := creator.ToString()
	assert.Equal(
		t,
		`line1-item1 line1-item2 line1-item3
line2-item1 line2-item2 line2-item3
line3-item1 line3-item2 line3-item3
`,
		fileString,
	)
}

func TestRecoveryForFileLevelError(t *testing.T) {
	calls := []testutils.TestCmd{
		{
			Cmd:      []string{testCommandString},
			Stdout:   "file-level error",
			ExitCode: 4,
		},
		fakeSuccessCommand,
	}
	fexec := testutils.GetFakeExecWithScripts(calls)
	creator := NewFileCreator(1, fexec)
	creator.AddErrorToRetryOn(NewErrorDefinition("file-level error"))
	require.NoError(t, creator.RunCommandWithFile(testCommandString))
}

func TestRecoveryForLineError(t *testing.T) {
	calls := []testutils.TestCmd{
		{
			Cmd:      []string{testCommandString},
			Stdout:   "failure on line 7",
			ExitCode: 4,
		},
		fakeSuccessCommand,
	}
	fexec := testutils.GetFakeExecWithScripts(calls)
	creator := NewFileCreator(1, fexec, "failure on line (\\d+)")
	require.NoError(t, creator.RunCommandWithFile(testCommandString))
}

func TestTotalFailureAfterRetries(t *testing.T) {
	errorCommand := testutils.TestCmd{
		Cmd:      []string{testCommandString},
		Stdout:   "some error",
		ExitCode: 4,
	}
	calls := []testutils.TestCmd{errorCommand, errorCommand, errorCommand}
	fexec := testutils.GetFakeExecWithScripts(calls)
	creator := NewFileCreator(2, fexec)
	require.Error(t, creator.RunCommandWithFile(testCommandString))
}

func TestHandleLineErrorForAbortSection(t *testing.T) {
	creator := NewFileCreator(1, nil)
	errorHandlers := []*LineErrorHandler{
		// first error handler doesn't match (include this to make sure the real match gets reached)
		{
			Definition: NewErrorDefinition("abc"),
			Method:     AbortSection,
			Reason:     "",
			Callback:   func() {},
		},
		{
			Definition: NewErrorDefinition("match-pattern"),
			Method:     AbortSection,
			Reason:     "error requiring us to abort section",
			Callback:   func() { log.Logf("abort section callback") },
		},
	}
	creator.AddLine(section1ID, errorHandlers, "line1-item1", "line1-item2", "line1-item3")
	creator.AddLine(section2ID, nil, "line2-item1", "line2-item2", "line2-item3")
	creator.AddLine(section1ID, nil, "line3-item1", "line3-item2", "line3-item3")
	stdErr := "failure: match-pattern do something please"
	creator.handleLineError(1, testCommandString, stdErr)
	fileString := creator.ToString()
	assert.Equal(t, "line2-item1 line2-item2 line2-item3\n", fileString)
}

func TestHandleLineErrorForSkipLine(t *testing.T) {
	creator := NewFileCreator(1, nil)
	errorHandlers := []*LineErrorHandler{
		{
			Definition: NewErrorDefinition("match-pattern"),
			Method:     SkipLine,
			Reason:     "error requiring us to skip this line",
			Callback:   func() { log.Logf("skip line callback") },
		},
	}
	creator.AddLine("", nil, "line1-item1", "line1-item2", "line1-item3")
	creator.AddLine("", errorHandlers, "line2-item1", "line2-item2", "line2-item3")
	creator.AddLine("", nil, "line3-item1", "line3-item2", "line3-item3")
	stdErr := "failure: match-pattern do something please"
	creator.handleLineError(2, testCommandString, stdErr)
	fileString := creator.ToString()
	assert.Equal(
		t,
		`line1-item1 line1-item2 line1-item3
line3-item1 line3-item2 line3-item3
`,
		fileString,
	)
}

func TestHandleLineErrorNoMatch(t *testing.T) {
	creator := NewFileCreator(1, nil)
	errorHandlers := []*LineErrorHandler{
		// first error handler doesn't match (include this to make sure the real match gets reached)
		{
			Definition: NewErrorDefinition("abc"),
			Method:     AbortSection,
			Reason:     "",
			Callback:   func() {},
		},
	}
	creator.AddLine("", nil, "line1-item1", "line1-item2", "line1-item3")
	creator.AddLine("", errorHandlers, "line2-item1", "line2-item2", "line2-item3")
	creator.AddLine("", nil, "line3-item1", "line3-item2", "line3-item3")
	stdErr := "failure: match-pattern do something please"
	fileStringBefore := creator.ToString()
	creator.handleLineError(2, testCommandString, stdErr)
	fileStringAfter := creator.ToString()
	require.Equal(t, fileStringBefore, fileStringAfter)
}

func TestGetErrorLineNumber(t *testing.T) {
	type args struct {
		lineFailurePatterns []string
		stdErr              string
	}

	tests := []struct {
		name            string
		args            args
		expectedLineNum int
	}{
		{
			"pattern that doesn't match",
			args{
				[]string{"abc"},
				"xyz",
			},
			-1,
		},
		{
			"matching pattern with no group",
			args{
				[]string{"abc"},
				"abc",
			},
			-1,
		},
		{
			"matching pattern with non-numeric group",
			args{
				[]string{"(abc)"},
				"abc",
			},
			-1,
		},
		{
			"stderr gives an out-of-bounds line number",
			args{
				[]string{"line (\\d+)"},
				"line 777",
			},
			-1,
		},
		{
			"good line match",
			args{
				[]string{"line (\\d+)"},
				`there was a failure
				on line 11 where the failure happened
				fix it please`,
			},
			11,
		},
		{
			"good line match with other pattern that doesn't match",
			args{
				[]string{"abc", "line (\\d+)"},
				`there was a failure
				on line 11 where the failure happened
				fix it please`,
			},
			11,
		},
	}

	commandString := "test command"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creator := NewFileCreator(0, nil, tt.args.lineFailurePatterns...)
			for i := 0; i < 15; i++ {
				creator.AddLine("", nil, fmt.Sprintf("line%d", i))
			}
			lineNum := creator.getErrorLineNumber(commandString, tt.args.stdErr)
			require.Equal(t, tt.expectedLineNum, lineNum)
		})
	}
}
