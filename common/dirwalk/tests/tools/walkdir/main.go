// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

// Walk a given directory and perform an action on the files found.

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/luci/luci-go/common/dirwalk"
)

func mainImpl() error {
	method := flag.String("method", "basic", "Method used to walk the tree")
	dir := flag.String("dir", "", "Directory to walk")
	do := flag.String("do", "nothing", "Action to perform on the files")
	smallFileSize := flag.Int64("smallfilesize", 64*1024, "Size to consider a small file")
	repeat := flag.Int("repeat", 1, "Repeat the walk x times")
	verbose := flag.Bool("v", false, "verbose mode")

	flag.Parse()

	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}

	if _, err := os.Stat(*dir); err != nil {
		return err
	}

	var base *BaseFileProcessor
	var proc FileProcessor
	switch *do {
	case "nothing":
		base = &BaseFileProcessor{}
		proc = base
	case "print":
		p := &PrintFileProcessor{obuf: os.Stderr}
		base = &p.BaseFileProcessor
		proc = p
	case "size":
		p := &SizeFileProcessor{obuf: os.Stderr}
		base = &p.BaseFileProcessor
		proc = p
	case "read":
		p := &ReadFileProcessor{}
		base = &p.BaseFileProcessor
		proc = p
	case "hash":
		p := &HashFileProcessor{obuf: os.Stderr}
		base = &p.BaseFileProcessor
		proc = p
	case "phash":
		p := CreateParallelHashFileProcessor(os.Stderr)
		base = &p.BaseFileProcessor
		proc = p
	case "verify":
		p := &VerifyFileProcessor{rootDir: *dir}
		base = &p.BaseFileProcessor
		proc = p
	default:
		log.Fatalf("Invalid action '%s'", *do)
	}

	for i := 0; i < *repeat; i++ {
		base.smallFiles = 0
		base.largeFiles = 0

		callback := func(path string, size int64, r io.ReadCloser, err error) {
			if err != nil {
				proc.Error(path, err)
				return
			}
			if r == nil {
				proc.Dir(path)
				return
			}

			if (size >= 0) && (size < *smallFileSize) {
				proc.SmallFile(path, r)
			} else {
				proc.LargeFile(path, r)
			}
		}

		switch *method {
		case "basic":
			dirwalk.WalkBasic(*dir, callback)
		case "nostat":
			dirwalk.WalkNoStat(*dir, *smallFileSize, callback)
		case "parallel":
			dirwalk.WalkParallel(*dir, callback)
		default:
			return fmt.Errorf("Invalid walk method '%s'", *method)
		}
		proc.Finished()
		fmt.Printf("Found %d small files and %d large files in %d dirs\n", base.smallFiles, base.largeFiles, base.dirs)
	}

	fmt.Fprintf(os.Stderr, "Found %d small files and %d large files in %d dirs\n", base.smallFiles, base.largeFiles, base.dirs)
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "walkdir: %s.\n", err)
		os.Exit(1)
	}
}
