package ioutil

import (
	"bytes"
	"fmt"
	"strings"

	kexec "k8s.io/utils/exec"
	// osexec "os/exec"
)

type FileCreator struct {
	lines      []string
	retryCount int
}

func InitFileCreator() *FileCreator {
	return &FileCreator{
		lines:      make([]string, 0),
		retryCount: 0,
	}
}

func (fileCreator *FileCreator) AddLine(items ...string) {
	spaceSeparatedItems := strings.Join(items, " ")
	fileCreator.lines = append(fileCreator.lines, spaceSeparatedItems)
}

func (fileCreator *FileCreator) ToString() string {
	return strings.Join(fileCreator.lines, "\n")
}

func (fileCreator *FileCreator) GetBuffer() *bytes.Buffer {
	return bytes.NewBufferString(fileCreator.ToString())
}

func (fileCreator *FileCreator) IncRetryCount() {
	fileCreator.retryCount++
}

func (fileCreator *FileCreator) GetRetryCount() int {
	return fileCreator.retryCount
}

// TODO do retries in this function? Possibly specify a function argument for specifying which line to retry at
func (fileCreator *FileCreator) RunWithFile(command kexec.Cmd) error { // or *osexec.Cmd
	command.SetStdin(fileCreator.GetBuffer())
	// command.Stdin = fileCreator.GetBuffer() // if we use os/exec

	// EXTRA (for debugging exit status 1)
	stdErrBuffer := bytes.NewBuffer(nil)
	command.SetStderr(stdErrBuffer)
	// command.Stderr = stdErrBuffer // if we use os/exec

	// TODO would this ever be useful?
	// outBuffer := bytes.NewBuffer(nil)
	// command.Stdout = outBuffer

	if err := command.Start(); err != nil {
		return fmt.Errorf("failed to start command for the FileCreator with error: %w", err)
	}
	if err := command.Wait(); err != nil {
		return fmt.Errorf("failed to run command for the FileCreator with error: %w: %s", err, stdErrBuffer.String())
	}

	return nil
}
