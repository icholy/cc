package compiler

import (
	"fmt"
	"regexp"
	"strconv"
)

const outputfmt = `
.globl  _main
_main:
	movl    $%d, %%eax
	leave
	ret
`

var re = regexp.MustCompile(`int main\s*\(\s*\)\s*{\s*return\s+(?P<ret>[0-9]+)\s*;\s*}`)

func Compile(src string) (string, error) {
	match := re.FindStringSubmatch(src)
	if len(match) != 2 {
		return "", fmt.Errorf("cannot find return value")
	}
	n, err := strconv.Atoi(match[1])
	if err != nil {
		return "", nil
	}
	return fmt.Sprintf(outputfmt, n), nil
}
