// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

// Walk a given directory and perform an action on the files found.

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/luci/luci-go/common/dirwalk"
)

func mainImpl() error {
	method := flag.String("method", "simple", "Method used to walk the tree")
	dir := flag.String("dir", "", "Directory to walk")
	do := flag.String("do", "nothing", "Action to perform on the files")
	smallfilesize := flag.Int64("smallfilesize", 64*1024, "Size to consider a small file")
	repeat := flag.Int("repeat", 1, "Repeat the walk x times")
	verbose := flag.Bool("v", false, "verbose mode")

	flag.Parse()

	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}

	if _, err := os.Stat(*dir); err != nil {
		return err
	}

	var proc *BaseFileProcessor
	switch *do {
	case "nothing":
		proc := &BaseFileProcessor{}
	case "print":
		proc := &PrintFileProcessor{obuf: os.Stderr}
	case "size":
		proc := &SizeFileProcessor{obuf: os.Stderr}
	case "read":
		proc := &ReadFileProcessor{}
	case "hash":
		proc := &HashFileProcessor{obuf: os.Stderr}
	case "phash":
		proc := CreateParallelHashFileProcessor(os.Stderr)
	default:
		log.Fatalf("Invalid action '%s'", *do)
	}

	for i := 0; i < *repeat; i++ {
		proc.smallfiles = 0
		proc.largefiles = 0

		switch *method {
		case "simple":
			dirwalk.WalkBasic(*dir, *smallfilesize, obs)
		case "nostat":
			dirwalk.WalkNoStat(*dir, *smallfilesize, obs)
		case "parallel":
			dirwalk.WalkParallel(*dir, *smallfilesize, obs)
		default:
			return errors.New(fmt.Sprintf("Invalid walk method '%s'", *method))
		}
		proc.Finished()
		fmt.Printf("Found %d small files and %d large files\n", proc.smallfiles, proc.largefiles)
	}

	fmt.Fprintf(os.Stderr, "Found %d small files and %d large files\n", proc.smallfiles, proc.largefiles)
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "walkdir: %s.\n", err)
		os.Exit(1)
	}
}
