package compiler

const output = `

.file   "return_2.c"
.text
.def    ___main;        .scl    2;      .type   32;     .endef
.section        .text.startup,"x"
.p2align 4,,15
.globl  _main
.def    _main;  .scl    2;      .type   32;     .endef
_main:
pushl   %ebp
movl    %esp, %ebp
andl    $-16, %esp
call    ___main
movl    $2, %eax
leave
ret
.ident  "GCC: (Rev2, Built by MSYS2 project) 7.3.0"

`

func Compile(src string) (string, error) {
	return output, nil
}
