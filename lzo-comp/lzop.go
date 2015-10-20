package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/rasky/go-lzo"
)

func decomp(args []string) error {
	in, err := ioutil.ReadFile(args[0])
	if err != nil {
		return err
	}
	out, err := lzo.Decompress1X(bytes.NewBuffer(in), 0, 0)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(out)
	return err
}

func main() {
	err := decomp(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}
}
