package dptestutils

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func AssertEqualFileStrings(t *testing.T, expectedFileString, actualFileString string) {
	if expectedFileString == actualFileString {
		return
	}
	fmt.Println("EXPECTED FILE STRING:")
	expectedLines := strings.Split(expectedFileString, "\n")
	for _, line := range expectedLines {
		fmt.Println(line)
	}
	fmt.Println("ACTUAL FILE STRING")
	actualLines := strings.Split(actualFileString, "\n")
	for _, line := range actualLines {
		fmt.Println(line)
	}
	if len(expectedLines) != len(actualLines) {
		fmt.Printf("expected %d lines, got %d\n", len(expectedLines), len(actualLines))
	}
	for k, expectedLine := range expectedLines {
		line := actualLines[k]
		if expectedLine != line {
			fmt.Printf("expected the next line, but got the one below it:\n%s\n%s\n", expectedLine, line)
			break
		}
	}
	require.FailNow(t, "got unexpected file string (see print contents above)")
}
