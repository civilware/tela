package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/civilware/tela"
	"github.com/civilware/tela/logger"
	"github.com/civilware/tela/shards"
	"github.com/deroproject/derohe/globals"
	"github.com/stretchr/testify/assert"
)

var mainPath string
var testDir string
var datashards string

func TestMain(m *testing.M) {
	dir, err := os.Getwd()
	if err != nil {
		logger.Fatalf("[TELA-CLI] %s\n", err)
	}

	mainPath = dir
	testDir = filepath.Join(mainPath, "cli_tests")
	datashards = filepath.Join(testDir, "datashards")

	err = os.MkdirAll(datashards, os.ModePerm)
	if err != nil {
		logger.Fatalf("[TELA-CLI] %s\n", err)
	}

	m.Run()
}

// Test functions of TELA-CLI
func TestCLIFunctions(t *testing.T) {
	var app tela_cli

	t.Cleanup(func() {
		os.RemoveAll(testDir)
		app.shutdownLocalServer()
	})

	os.RemoveAll(testDir)
	err := tela.SetShardPath(testDir)
	if err != nil {
		t.Fatalf("Could not set shard path for tests: %s", err)
	}

	// // Test parseFlags
	t.Run("ParseFlags", func(t *testing.T) {
		//  No flags
		app.parseFlags()
		assert.False(t, globals.Arguments["--testnet"].(bool), "Network should be mainnet default with no saved preferences")
		assert.False(t, globals.Arguments["--simulator"].(bool), "Network should be mainnet default with no saved preferences")

		// Set flags
		os.Args = []string{
			"--testnet",
			"--simulator",
			"--debug",
			"--db-type=boltdb",
			"--fastsync=true",
			"--num-parallel-blocks=4",
		}
		app.parseFlags()
		assert.True(t, globals.Arguments["--testnet"].(bool), "--testnet was not set")
		assert.True(t, globals.Arguments["--simulator"].(bool), "--simulator was not set")
		assert.True(t, globals.Arguments["--debug"].(bool), "--debug was not set")
		assert.Equal(t, "boltdb", shards.GetDBType(), "--db-type was not set")
		assert.True(t, gnomon.fastsync, "--fastsync was not set")
		assert.Equal(t, 4, gnomon.parallelBlocks, "--num-parallel-blocks was not set")
		globals.InitNetwork()
		assert.Equal(t, "Simulator", getNetworkInfo(), "getNetworkInfo should return simulator")

		// Set --mainnet override
		os.Args = []string{
			"--testnet",
			"--simulator",
			"--mainnet",
		}
		app.parseFlags()
		assert.False(t, globals.Arguments["--testnet"].(bool), "--mainnet did not override --testnet")
		assert.False(t, globals.Arguments["--simulator"].(bool), "--mainnet did not override --simulator")
		globals.InitNetwork()
		assert.Equal(t, "Mainnet", getNetworkInfo(), "getNetworkInfo should return mainnet")
	})

	// // Test completer functions
	t.Run("Completers", func(t *testing.T) {
		assert.Empty(t, completerNothing(), "completerNothing should return empty")
		assert.Empty(t, app.completerSearchExclusions(), "completerSearchFilters should return empty without exclusions")
		app.exclusions = []string{"exclude"}
		assert.NotEmpty(t, app.completerSearchExclusions(), "completerSearchFilters should not be empty with exclusions")
		assert.NotEmpty(t, completerNetworks(false), "completerNetworks should not be empty")
		assert.NotEmpty(t, completerNetworks(true), "completerNetworks should not be empty")
		assert.NotEmpty(t, completerTrueFalse(), "completerTrueFalse should not be empty")
		assert.NotEmpty(t, completerYesNo(), "completerYesNo should not be empty")
		assert.NotEmpty(t, completerDocType(), "completerDocType should not be empty")
		assert.NotEmpty(t, completerMODs("tx"), "completerMODs should not be empty")
		assert.NotEmpty(t, completerMODClasses(), "completerMODClasses should not be empty")
		// completerSCFunctionNames()
		assert.NotEmpty(t, completerFiles("."), "completerFiles should not be empty")
		// completerServers()
	})

	// // Test error functions
	t.Run("ErrorFunctions", func(t *testing.T) {
		eofErr := fmt.Errorf("EOF")
		interruptErr := fmt.Errorf("Interrupt")
		networkErr := fmt.Errorf("Mainnet/TestNet")

		// Test readError
		assert.True(t, readError(interruptErr), "readError should return interrupt as true")
		assert.False(t, readError(eofErr), "readError should return false when not interrupt")

		// Test networkError
		err := networkError(networkErr)
		assert.ErrorContains(t, err, "invalid network settings, run command", "Network mismatch should return info error")
		err = networkError(interruptErr)
		assert.EqualError(t, err, interruptErr.Error(), "networkError should pass through non network errors")

		// Test lineError
		assert.EqualError(t, app.lineError(eofErr), eofErr.Error(), "lineError should pass through EOF error")
		assert.Nil(t, app.lineError(nil), "lineError should pass through nil error")
	})

	// // Test findDocShardFiles
	t.Run("FindDocShards", func(t *testing.T) {
		moveTo := filepath.Join(datashards, "clone", "shard")
		var docShardFiles = []struct {
			Name   string
			Source string
			Path   string
		}{
			{
				Name:   "splitMain.go",
				Source: "main.go",
				Path:   filepath.Join(moveTo, "splitMain.go"),
			},
		}

		err := os.MkdirAll(moveTo, os.ModePerm)
		assert.NoError(t, err, "Creating directories should not error: %s", err)

		// Shard and then reconstruct the files
		for _, shardFile := range docShardFiles {
			content, _ := readFile(shardFile.Source)
			file, err := os.Create(shardFile.Path)
			assert.NoError(t, err, "Creating original file should not error: %s", err)

			_, err = file.Write([]byte(content))
			assert.NoError(t, err, "Writing original file should not error: %s", err)

			err = tela.CreateShardFiles(shardFile.Path)
			assert.NoError(t, err, "Sharding original file should not error: %s", err)

			shardEntrypoint := filepath.Join(moveTo, strings.ReplaceAll(shardFile.Name, ".go", "-1.go"))
			docShards, recreate, err := findDocShardFiles(shardEntrypoint)
			assert.NoError(t, err, "Finding shard files should not error: %s", err)
			assert.Equal(t, shardFile.Name, recreate, "Recreated file name should be the same")

			err = os.RemoveAll(filepath.Join(moveTo, shardFile.Name))
			assert.NoError(t, err, "Removing original file should not error: %s", err)

			err = tela.ConstructFromShards(docShards, recreate, moveTo)
			assert.NoError(t, err, "Recreating original file should not error: %s", err)

			newContent, err := readFile(shardFile.Source)
			assert.NoError(t, err, "Reading the recreated file should not error: %s", err)
			assert.Equal(t, content, newContent, "Recreated content should match original")
		}

		_, _, err = findDocShardFiles("main.go")
		assert.Error(t, err, "findDocShardFiles should error with non shard file")
	})

	// Test gitClone if instructed to
	t.Run("GitClone", func(t *testing.T) {
		if os.Getenv("RUN_GIT_TEST") != "true" {
			t.Skipf("Use %q to run Git test", "RUN_GIT_TEST=true go test . -v")
		}

		_, err := exec.LookPath("git")
		if err != nil {
			t.Skipf("Git is not installed for clone tests")
		}

		// Valid git clone
		err = gitClone("github.com/civilware/tela@main")
		assert.NoError(t, err, "Cloning valid repo should not error: %s", err)

		// Invalid git clone
		err = gitClone("https://github.com/civilware/tela")
		assert.Error(t, err, "Cloning with long form URL should error")
		err = gitClone("github.com/civilware/tela.git")
		assert.Error(t, err, "Cloning with .git URL suffix should error")
		err = gitClone("civilware/tela")
		assert.Error(t, err, "Cloning with invalid URL should error")
		err = gitClone("/ / /")
		assert.Error(t, err, "Cloning with invalid URL parts should error")
	})

	// // Test getFileDiff
	t.Run("GetDiff", func(t *testing.T) {
		diff, fileNames, err := getFileDiff("gnomon.go", "gnomon.go")
		assert.NoError(t, err, "getFileDiff should not error with valid files: %s", err)
		assert.Len(t, diff, 2, "getFileDiff should return 2 diffs")
		assert.Len(t, fileNames, 2, "getFileDiff should return 2 fileNames")
		assert.NotEmpty(t, diff[0], "getFileDiff code1 should not be empty")
		assert.NotEmpty(t, diff[1], "getFileDiff code2 should not be empty")

		_, _, err = getFileDiff(" ", " ")
		assert.Error(t, err, "getFileDiff should error with invalid files")
	})

	// // Test printDiff
	t.Run("PrintDiff", func(t *testing.T) {
		var diffTests = []struct {
			diff     []string // The things to compare
			expected string   // Expected result of comparison
		}{
			// No diff
			{
				diff: []string{
					`1
2
3
4
5
6`,
					`1
2
3
4
5
6`,
				},
				expected: "No diffs found",
			},
			// Diff line 4, same file len
			{
				diff: []string{
					`1
2
3
4
5
6`,
					`1
2
3
3
5
6`,
				},
				expected: fmt.Sprintf("%s\n%s", "4 - 4", "4 + 3"),
			},
			// Diff line 4 and 9 to hit context, same file len
			{
				diff: []string{
					`1
2
3
4
5
6
7
8
9
10`,
					`1
2
3
3
5
6
7
8
8
10`,
				},
				expected: fmt.Sprintf("%s\n%s\n\n%s\n%s", "4 - 4", "4 + 3", "9 - 9", "9 + 8"),
			},
			// Diff, file1 longer
			{
				diff: []string{
					`1
2
3
4
5
6`,
					`1
2
6`,
				},
				expected: fmt.Sprintf("%s\n%s\n%s\n%s\n%s", "3 - 3", "3 + 6", "4 - 4", "5 - 5", "6 - 6"),
			},
			// Diff, file1 longer same end
			{
				diff: []string{
					`1
2
3
4
5
6`,
					`1
2
3`,
				},
				expected: fmt.Sprintf("%s\n%s\n%s", "4 - 4", "5 - 5", "6 - 6"),
			},
			// Diff, file2 longer
			{
				diff: []string{
					`1
2
6`,
					`1
2
3
4
5
6`,
				},
				expected: fmt.Sprintf("%s\n%s\n%s\n%s\n%s", "3 - 6", "3 + 3", "4 + 4", "5 + 5", "6 + 6"),
			},
			// Diff, file2 longer same end
			{
				diff: []string{
					`1
2
3`,
					`1
2
3
4
5
6`,
				},
				expected: fmt.Sprintf("%s\n%s\n%s", "4 + 4", "5 + 5", "6 + 6"),
			},
		}

		app.pageSize = 20

		logger.EnableColors(false)
		outputHeader := "--- a/file1.txt\n+++ b/file2.txt"

		for i, d := range diffTests {
			output := captureOutput(func() {
				err := app.printDiff(d.diff, []string{"file1.txt", "file2.txt"})
				assert.NoError(t, err, "printDiff with valid inputs should not error: %s", err)
			})

			if i > 0 {
				assert.Contains(t, output, outputHeader, "Expected output %d to have matching header", i)
			}
			assert.Contains(t, output, d.expected, "Output diff %d does not match expected", i)
		}

		err := app.printDiff([]string{}, []string{"file1.txt", "file2.txt"})
		assert.Error(t, err, "printDiff with invalid diff should error")
		err = app.printDiff([]string{"1", "1"}, []string{})
		assert.Error(t, err, "printDiff with invalid filenames should error")
	})

	t.Run("ServeLocal", func(t *testing.T) {
		// Test serveLocal
		err := app.serveLocal(".")
		assert.NoError(t, err, "serveLocal should not error: %s", err)

		// Test openBrowser
		app.os = runtime.GOOS
		err = app.openBrowser("http://localhost:8080")
		assert.NoError(t, err, "openBrowser should not error opening localhost: %s", err)

		// Test getServerInfo
		_, _, localRunning := app.getServerInfo()
		assert.True(t, localRunning, "Local server should be running")

		// Test getCLIInfo
		assert.NotEmpty(t, app.getCLIInfo(), "getCLIInfo should not be empty")
	})
}

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

// Read files for test
func readFile(elem ...string) (string, error) {
	path := mainPath
	for _, e := range elem {
		path = filepath.Join(path, e)
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(file), nil
}
