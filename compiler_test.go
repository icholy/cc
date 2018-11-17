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
		{
			Name:     "multiple_vars",
			SrcPath:  "testdata/stage_5/valid/multiple_vars.c",
			ExitCode: 3,
		},
		{
			Name:     "exp_return_val",
			SrcPath:  "testdata/stage_5/valid/exp_return_val.c",
			ExitCode: 0,
		},
		{
			Name:     "if_nested_2",
			SrcPath:  "testdata/stage_6/valid/if_nested_2.c",
			ExitCode: 2,
		},
		{
			Name:     "if_nested_3",
			SrcPath:  "testdata/stage_6/valid/if_nested_3.c",
			ExitCode: 3,
		},
		{
			Name:     "if_nested_4",
			SrcPath:  "testdata/stage_6/valid/if_nested_4.c",
			ExitCode: 4,
		},
		{
			Name:     "consecutive_blocks.c",
			SrcPath:  "testdata/stage_7/valid/consecutive_blocks.c",
			ExitCode: 1,
		},
		{
			Name:     "consecutive_declarations.c",
			SrcPath:  "testdata/stage_7/valid/consecutive_declarations.c",
			ExitCode: 3,
		},
		{
			Name:     "multi_nesting.c",
			SrcPath:  "testdata/stage_7/valid/multi_nesting.c",
			ExitCode: 3,
		},
		{
			Name:     "nested_scope.c",
			SrcPath:  "testdata/stage_7/valid/nested_scope.c",
			ExitCode: 4,
		},
		{
			Name:     "while_single_statement.c",
			SrcPath:  "testdata/stage_8/valid/while_single_statement.c",
			ExitCode: 6,
		},
		{
			Name:     "while_multi_statement.c",
			SrcPath:  "testdata/stage_8/valid/while_multi_statement.c",
			ExitCode: 6,
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
	AssertValid(t, 5)
	AssertValid(t, 6)
	AssertValid(t, 7)
	AssertValid(t, 8)
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
