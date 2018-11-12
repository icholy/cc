package parser

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"gotest.tools/assert"
)

func TestValidParsing(t *testing.T) {
	AssertParsingStage(t, 1)
	AssertParsingStage(t, 2)
}

type validityTest struct {
	SrcPath string
	Valid   bool
}

func AssertParsingStage(t *testing.T, stage int) {
	var tests []validityTest
	valid, err := filepath.Glob(fmt.Sprintf("../testdata/stage_%d/valid/*.c", stage))
	assert.NilError(t, err)
	for _, path := range valid {
		tests = append(tests, validityTest{path, true})
	}
	invalid, err := filepath.Glob(fmt.Sprintf("../testdata/stage_%d/invalid/*.c", stage))
	assert.NilError(t, err)
	for _, path := range invalid {
		tests = append(tests, validityTest{path, false})
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
