package parser

import (
	"io/ioutil"
	"testing"

	"gotest.tools/assert"
)

func TestStage1(t *testing.T) {
	tests := []struct {
		SrcPath string
		Valid   bool
	}{
		// valid cases
		{SrcPath: "../testdata/stage_1/valid/multi_digit.c", Valid: true},
		{SrcPath: "../testdata/stage_1/valid/newlines.c", Valid: true},
		{SrcPath: "../testdata/stage_1/valid/spaces.c", Valid: true},
		{SrcPath: "../testdata/stage_1/valid/return_0.c", Valid: true},
		{SrcPath: "../testdata/stage_1/valid/return_2.c", Valid: true},

		// valid cases
		{SrcPath: "../testdata/stage_1/invalid/missing_paren.c", Valid: false},
		{SrcPath: "../testdata/stage_1/invalid/missing_retval.c", Valid: false},
		{SrcPath: "../testdata/stage_1/invalid/no_brace.c", Valid: false},
		{SrcPath: "../testdata/stage_1/invalid/no_space.c", Valid: false},
		{SrcPath: "../testdata/stage_1/invalid/wrong_case.c", Valid: false},
	}
	for _, tt := range tests {
		t.Run(tt.SrcPath, func(t *testing.T) {
			src, err := ioutil.ReadFile(tt.SrcPath)
			assert.NilError(t, err)
			_, err = Parse(string(src))
			if tt.Valid {
				assert.NilError(t, err)
			} else {
				assert.Assert(t, err != nil)
			}
		})
	}
}
