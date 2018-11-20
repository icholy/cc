package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/icholy/cc/compiler"
)

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		log.Fatalf("no input files")
	}
	for _, file := range flag.Args() {
		if err := compile(file); err != nil {
			log.Fatalf("%s: %v", file, err)
		}
	}
}

func compile(file string) error {
	src, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	asm, err := compiler.Compile(string(src))
	if err != nil {
		return err
	}
	name := outputName(file)
	return ioutil.WriteFile(name, []byte(asm), os.ModePerm)
}

func outputName(file string) string {
	ext := filepath.Ext(file)
	return fmt.Sprintf("%s.s", file[:len(file)-len(ext)])
}
