package platform

import (
	"os"
	"strconv"
	"strings"
	"testing"

	testutils "github.com/Azure/azure-container-networking/test/utils"
)

func TestMain(m *testing.M) {
	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestGetLastRebootTime(t *testing.T) {
	fexec := testutils.GetFakeExecWithScripts([]testutils.TestCmd{})
	p := New(fexec)
	_, err := p.GetLastRebootTime()
	if err != nil {
		t.Errorf("GetLastRebootTime failed :%v", err)
	}
}

func TestGetOSDetails(t *testing.T) {
	fexec := testutils.GetFakeExecWithScripts([]testutils.TestCmd{})
	p := New(fexec)
	_, err := p.GetOSDetails()
	if err != nil {
		t.Errorf("GetOSDetails failed :%v", err)
	}
}

func TestGetProcessNameByID(t *testing.T) {
	fexec := testutils.GetFakeExecWithScripts([]testutils.TestCmd{})
	p := New(fexec)
	pName, err := p.GetProcessNameByID(strconv.Itoa(os.Getpid()))
	if err != nil {
		t.Errorf("GetProcessNameByID failed: %v", err)
	}

	if !strings.Contains(pName, "platform.test") {
		t.Errorf("Incorrect process name:%v\n", pName)
	}
}

func TestReadFileByLines(t *testing.T) {
	lines, err := readFileByLines("testfiles/test1")
	if err != nil {
		t.Errorf("ReadFileByLines failed: %v", err)
	}

	if len(lines) != 2 {
		t.Errorf("Line count %d didn't match expected count", len(lines))
	}

	lines, err = readFileByLines("testfiles/test2")
	if err != nil {
		t.Errorf("ReadFileByLines failed: %v", err)
	}

	if len(lines) != 1 {
		t.Errorf("Line count %d didn't match expected count", len(lines))
	}

	lines, err = readFileByLines("testfiles/test3")
	if err != nil {
		t.Errorf("ReadFileByLines failed: %v", err)
	}

	if len(lines) != 2 {
		t.Errorf("Line count %d didn't match expected count", len(lines))
	}

	if lines[1] != "" {
		t.Errorf("Expected empty line but got %s", lines[1])
	}
}

func TestFileExists(t *testing.T) {
	fexec := testutils.GetFakeExecWithScripts([]testutils.TestCmd{})
	p := New(fexec)
	isExist, err := p.CheckIfFileExists("testfiles/test1")
	if err != nil || !isExist {
		t.Errorf("Returned file not found %v", err)
	}

	isExist, err = p.CheckIfFileExists("testfiles/filenotfound")
	if err != nil || isExist {
		t.Errorf("Returned file found")
	}
}
