package logger

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/deroproject/derohe/globals"
	"github.com/stretchr/testify/assert"
)

// Capture the console output
func captureOutput(f func()) string {
	var buf bytes.Buffer
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = stdout
	buf.ReadFrom(r)
	return buf.String()
}

func TestEnableColors(t *testing.T) {
	EnableColors(false)
	assert.Empty(t, Color.Red(), "Color red should be empty when colors are disabled")
	assert.Empty(t, Color.End(), "Color end should be empty when colors are disabled")
	assert.Equal(t, INFO, tag.info, "INFO tag should match when colors are disabled")
	assert.Equal(t, ERROR, tag.err, "ERROR tag should match  when colors are disabled")

	EnableColors(true)
	expectedInfo := ANSIgreen + INFO + ANSIend
	expectedError := ANSIred + ERROR + ANSIend
	assert.Equal(t, ANSIblack, Color.Black(), "Color black should match ANSI when colors are enabled")
	assert.Equal(t, ANSIred, Color.Red(), "Color red should match ANSI when colors are enabled")
	assert.Equal(t, ANSIgreen, Color.Green(), "Color green should match ANSI when colors are enabled")
	assert.Equal(t, ANSIyellow, Color.Yellow(), "Color yellow should match ANSI when colors are enabled")
	assert.Equal(t, ANSIblue, Color.Blue(), "Color blue should match ANSI when colors are enabled")
	assert.Equal(t, ANSImagenta, Color.Magenta(), "Color magenta should match ANSI when colors are enabled")
	assert.Equal(t, ANSIcyan, Color.Cyan(), "Color cyan should match ANSI when colors are enabled")
	assert.Equal(t, ANSIwhite, Color.White(), "Color white should match ANSI when colors are enabled")
	assert.Equal(t, ANSIdefault, Color.Default(), "Color default should match ANSI when colors are enabled")
	assert.Equal(t, ANSIgrey, Color.Grey(), "Color grey should match ANSI when colors are enabled")
	assert.Equal(t, ANSIend, Color.End(), "Color end should match ANSI when colors are enabled")
	assert.Equal(t, expectedInfo, tag.info, "INFO tag should match when colors are enabled")
	assert.Equal(t, expectedError, tag.err, "ERROR tag should match  when colors are enabled")
}

func TestTimestamp(t *testing.T) {
	stamp := Timestamp()
	assert.Contains(t, stamp, "[", "Timestamp missing left format")
	assert.Contains(t, stamp, "]", "Timestamp missing right format")
}

func TestSprint(t *testing.T) {
	result := sprint("[source] message")
	expected := fmt.Sprintf("%s%s:%s message", Color.Cyan(), "source", Color.End())
	assert.Equal(t, expected, result, "Expected sprint to parse source tag")

	message := "message"
	result = sprint(message)
	assert.Equal(t, message, result, "Expected message without source to remain the same")
}

func TestDebugf(t *testing.T) {
	format := "Hello %s %d %t"
	a := []any{"TELA", 1, true}

	globals.Arguments["--debug"] = true
	output := captureOutput(func() {
		Debugf(format, a...)
	})

	assert.Contains(t, output, DEBUG, "Expected output to contain DEBUG tag")
	assert.Contains(t, output, fmt.Sprintf(format, a...), "Expecting output to match format")

	globals.Arguments["--debug"] = false
	output = captureOutput(func() {
		Debugf(format, a...)
	})

	assert.Empty(t, output, "Expected empty output when --debug is not enabled")
}

func TestPrintf(t *testing.T) {
	format := "Hello %s %d %t"
	a := []any{"TELA", 1, true}

	output := captureOutput(func() {
		Printf(format, a...)
	})

	assert.Contains(t, output, INFO, "Expected output to contain INFO tag")
	assert.Contains(t, output, fmt.Sprintf(format, a...), "Expecting output to match format")
}

func TestWarnf(t *testing.T) {
	format := "Warning %s %d %t"
	a := []any{"issued", 1, true}

	output := captureOutput(func() {
		Warnf(format, a...)
	})

	assert.Contains(t, output, WARN, "Expected output to contain WARN tag")
	assert.Contains(t, output, fmt.Sprintf(format, a...), "Expecting output to match format")
}

func TestErrorf(t *testing.T) {
	format := "Error %s %d %t"
	a := []any{"occurred", 1, true}

	output := captureOutput(func() {
		Errorf(format, a...)
	})

	assert.Contains(t, output, ERROR, "Expected output to contain ERROR tag")
	assert.Contains(t, output, fmt.Sprintf(format, a...), "Expecting output to match format")
}

func TestFatalf(t *testing.T) {
	if os.Getenv("FLAG") == "1" {
		Fatalf("Fatalf test")
	}

	// Run the test in a subprocess
	cmd := exec.Command(os.Args[0], "-test.run=TestFatalf")
	cmd.Env = append(os.Environ(), "FLAG=1")
	err := cmd.Run()

	_, ok := err.(*exec.ExitError)
	expectedErrorString := "exit status 1"
	assert.True(t, ok, "Error should be %T", new(exec.ExitError))
	assert.EqualError(t, err, expectedErrorString, "Error string should be equal")
}
