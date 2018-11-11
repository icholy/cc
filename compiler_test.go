package compiler

import (
	"io/ioutil"
	"os/exec"
	"syscall"
	"testing"

	"gotest.tools/assert"
	"gotest.tools/fs"
)

func TestOutput(t *testing.T) {
	tests := []OutputTest{
		{
			Name:     "return_2",
			SrcPath:  "stage_1/return_2.c",
			ExitCode: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

type OutputTest struct {
	Name     string
	SrcPath  string
	Ouput    string
	ExitCode int
}

func (tt *OutputTest) Run(t *testing.T) {
	// read the source file
	src, err := ioutil.ReadFile(tt.SrcPath)
	assert.NilError(t, err)
	// compile to assembly
	code, err := Compile(string(src))
	assert.NilError(t, err)
	// assembly & link with gcc
	dir := fs.NewDir(t, "cc", fs.WithFile("out.s", code))
	defer dir.Remove()
	gcc := exec.Command("gcc", "-m32", "out.s", "-o", "out")
	gcc.Dir = dir.Path()
	// check the output
	output, err := gcc.CombinedOutput()
	assert.NilError(t, err, string(output))
	// run the binary
	bin := exec.Command("out")
	bin.Dir = dir.Path()
	output, err = bin.CombinedOutput()
	assert.NilError(t, err, string(output))
	// check the output
	assert.Equal(t, ExitCode(bin, err), tt.ExitCode)
	assert.Equal(t, string(output), tt.Ouput)
}

const defaultFailedCode = 1

func ExitCode(cmd *exec.Cmd, err error) int {
	if err != nil {
		// try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			return ws.ExitStatus()
		} else {
			return 1
		}
	} else {
		// success, exitCode should be 0 if go is ok
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		return ws.ExitStatus()
	}
}
