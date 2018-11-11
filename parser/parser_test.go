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
func TestStage2(t *testing.T) {
	tests := []struct {
		SrcPath string
		Valid   bool
	}{
		// valid cases
		{SrcPath: "../testdata/stage_2/valid/bitwise_zero.c", Valid: true},
		{SrcPath: "../testdata/stage_2/valid/bitwise.c", Valid: true},
		{SrcPath: "../testdata/stage_2/valid/neg.c", Valid: true},
		{SrcPath: "../testdata/stage_2/valid/nested_ops.c", Valid: true},
		{SrcPath: "../testdata/stage_2/valid/nested_ops_2.c", Valid: true},
		{SrcPath: "../testdata/stage_2/valid/not_five.c", Valid: true},
		{SrcPath: "../testdata/stage_2/valid/not_zero.c", Valid: true},

		// invalid cases
		{SrcPath: "../testdata/stage_2/invalid/missing_const.c", Valid: false},
		{SrcPath: "../testdata/stage_2/invalid/missing_semicolon.c", Valid: false},
		{SrcPath: "../testdata/stage_2/invalid/nested_missing_const.c", Valid: false},
		{SrcPath: "../testdata/stage_2/invalid/wrong_order.c", Valid: false},
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
