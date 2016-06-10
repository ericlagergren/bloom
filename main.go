// +build ignore

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/EricLagergren/bloom"
)

func main() {
	f := bloom.New(119095, 0.01)

	file, err := os.Open("/usr/share/dict/usa")
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	s := bufio.NewScanner(file)
	for s.Scan() {
		f.Add(s.Text())
	}

	_, err = file.Seek(0, os.SEEK_SET)
	if err != nil {
		log.Fatalln(err)
	}

	s = bufio.NewScanner(file)
	for s.Scan() {
		if !f.Has(s.Text()) {
			fmt.Printf("couldn't find %q", s.Text())
			os.Exit(1)
		}
		os.Exit(1)
	}
}
