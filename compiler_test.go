package cc

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"

	"gotest.tools/assert"
	"gotest.tools/fs"
)

func TestOutput(t *testing.T) {
	tests := []struct {
		Name     string
		SrcPath  string
		Ouput    string
		ExitCode int
	}{
		{
			Name:     "return_2",
			SrcPath:  "testdata/stage_1/valid/return_2.c",
			ExitCode: 2,
		},
		{
			Name:     "bitwise",
			SrcPath:  "testdata/stage_2/valid/bitwise.c",
			ExitCode: 0,
		},
		{
			Name:     "add",
			SrcPath:  "testdata/stage_3/valid/precedence.c",
			ExitCode: 14,
		},
		{
			Name:     "add",
			SrcPath:  "testdata/stage_4/valid/or_true.c",
			ExitCode: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			dir := fs.NewDir(t, tt.Name)
			defer dir.Remove()
			assertCompile(t, dir, tt.SrcPath, true)
			output, exitcode := assertRun(t, dir)
			assert.Equal(t, exitcode, tt.ExitCode)
			assert.Equal(t, output, tt.Ouput)
		})
	}
}

func TestStages(t *testing.T) {
	AssertValid(t, 1)
	AssertValid(t, 2)
	AssertValid(t, 3)
	AssertValid(t, 4)
}

func AssertValid(t *testing.T, stage int) {
	t.Run(fmt.Sprintf("stage_%d", stage), func(t *testing.T) {
		pattern := fmt.Sprintf("testdata/stage_%d/valid/*.c", stage)
		valid, err := filepath.Glob(pattern)
		assert.NilError(t, err)
		for _, srcpath := range valid {
			t.Run(filepath.Base(srcpath), func(t *testing.T) {
				dir := fs.NewDir(t, "cc")
				defer dir.Remove()
				assertCompile(t, dir, srcpath, true)
			})
		}
	})
}

const binName = "out.exe"

func assertRun(t *testing.T, dir *fs.Dir) (string, int) {
	bin := exec.Command(dir.Join(binName))
	output, err := bin.CombinedOutput()
	return string(output), exitCode(bin, err)
}

func assertCompile(t *testing.T, dir *fs.Dir, srcpath string, valid bool) {
	src, err := ioutil.ReadFile(srcpath)
	assert.NilError(t, err)
	// compile to assembly
	asm, err := Compile(string(src))
	assert.NilError(t, err)
	assertWriteFile(t, dir, "out.s", asm)
	gcc := exec.Command("gcc", "-m32", "out.s", "-o", binName)
	gcc.Dir = dir.Path()
	// check the output
	output, err := gcc.CombinedOutput()
	if valid {
		assert.NilError(t, err, string(output))
	} else {
		assert.Error(t, err, string(output))
	}
}

func assertWriteFile(t *testing.T, dir *fs.Dir, name, content string) {
	t.Helper()
	err := ioutil.WriteFile(dir.Join(name), []byte(content), os.ModePerm)
	assert.NilError(t, err)
}

func exitCode(cmd *exec.Cmd, err error) int {
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
