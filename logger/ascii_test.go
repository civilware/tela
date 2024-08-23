package logger

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestASCIIBlend(t *testing.T) {
	var margin []string
	margin = append(margin, "------------------------------")
	margin = append(margin, "------------------------")

	defaultSpace := "                                  "

	var info [][]string
	info = append(info, []string{margin[0]})
	info = append(info, []string{margin[1]})
	info = append(info, []string{"      "})
	info = append(info, []string{""})

	expectedOutput1 := margin[0] + `    ` + ASCIISmall[0] + "\n" +
		`Some info 0                       ` + ASCIISmall[1] + "\n" +
		`Some info 1                       ` + ASCIISmall[2] + "\n" +
		`Some info 2                       ` + ASCIISmall[3] + "\n" +
		defaultSpace + ASCIISmall[4] + "\n" +
		defaultSpace + ASCIISmall[5] + "\n" +
		defaultSpace + ASCIISmall[6] + "\n" +
		defaultSpace + ASCIISmall[7] + "\n" +
		defaultSpace + ASCIISmall[8] + "\n" +
		defaultSpace + ASCIISmall[9] + "\n" +
		defaultSpace + ASCIISmall[10] + "\n" +
		defaultSpace + ASCIISmall[11] + "\n"

	// Match info line for expectedOutput1
	for i := 0; i < 3; i++ {
		info[0] = append(info[0], fmt.Sprintf("Some info %d", i))
	}

	expectedOutput2 := margin[1] + `    ` + ASCIISmall[0] + "\n" +
		`Some info 0                 ` + ASCIISmall[1] + "\n" +
		`Some info 1                 ` + ASCIISmall[2] + "\n" +
		`Some info 2                 ` + ASCIISmall[3] + "\n" +
		`Some info 3                 ` + ASCIISmall[4] + "\n" +
		`Some info 4                 ` + ASCIISmall[5] + "\n" +
		`Some info 5                 ` + ASCIISmall[6] + "\n" +
		`Some info 6                 ` + ASCIISmall[7] + "\n" +
		`Some info 7                 ` + ASCIISmall[8] + "\n" +
		`Some info 8                 ` + ASCIISmall[9] + "\n" +
		`Some info 9                 ` + ASCIISmall[10] + "\n" +
		`Some info 10                ` + ASCIISmall[11] + "\n" +
		`Some info 11` + "\n" + `Some info 12` + "\n" + `Some info 13` + "\n" + `Some info 14` + "\n"

	// Match info line for expectedOutput2
	for i := 0; i < 15; i++ {
		info[1] = append(info[1], fmt.Sprintf("Some info %d", i))
	}

	var expectedOutput3 string
	for _, line := range ASCIISmall {
		expectedOutput3 = expectedOutput3 + defaultSpace + line + "\n"
	}

	var expectedOutput4 string
	for _, line := range ASCIISmall {
		expectedOutput4 = expectedOutput4 + defaultSpace + line + "\n"
	}

	var expectedOutputs []string
	expectedOutputs = append(expectedOutputs, expectedOutput1)
	expectedOutputs = append(expectedOutputs, expectedOutput2)
	expectedOutputs = append(expectedOutputs, expectedOutput3)
	expectedOutputs = append(expectedOutputs, expectedOutput4)

	for i, eo := range expectedOutputs {
		output := captureOutput(func() {
			ASCIIBlend(ASCIISmall, info[i])
		})

		assert.Equal(t, eo, output, "Blend output %d should be the same", i)

		// Test with nil info
		output = captureOutput(func() {
			ASCIIBlend(ASCIISmall, nil)
		})

		assert.Empty(t, output, "Blending nil %d should be empty", i)
	}
}

func TestASCIIPrint(t *testing.T) {
	var expectedOutputMain string
	for _, line := range ASCIIMain {
		expectedOutputMain = expectedOutputMain + line + "\n"
	}

	var expectedOutputSmall string
	for _, line := range ASCIISmall {
		expectedOutputSmall = expectedOutputSmall + line + "\n"
	}

	output := captureOutput(func() {
		ASCIIPrint(false)
	})

	assert.Equal(t, expectedOutputMain, output, "Main output should be the same")

	output = captureOutput(func() {
		ASCIIPrint(true)
	})

	assert.Equal(t, expectedOutputSmall, output, "Small output should be the same")
}
