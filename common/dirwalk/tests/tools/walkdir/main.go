// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

// Walk a given directory and perform an action on the files found.

import (
	"flag"
	"fmt"
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

	var stats *NullWalker
	var obs dirwalk.WalkObserver
	switch *do {
	case "nothing":
		o := &NullWalker{}
		stats = o
		obs = o
	case "print":
		o := &PrintWalker{obuf: os.Stderr}
		stats = &o.NullWalker
		obs = o
	case "size":
		o := &SizeWalker{obuf: os.Stderr}
		stats = &o.NullWalker
		obs = o
	case "read":
		o := &ReadWalker{}
		stats = &o.NullWalker
		obs = o
	case "hash":
		o := &HashWalker{obuf: os.Stderr}
		stats = &o.NullWalker
		obs = o
	case "phash":
		o := CreateParallelHashWalker(os.Stderr)
		stats = &o.NullWalker
		obs = o
	default:
		log.Fatalf("Invalid action '%s'", *do)
	}

	for i := 0; i < *repeat; i++ {
		stats.smallfiles = 0
		stats.largefiles = 0

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
		fmt.Printf("Found %d small files and %d large files\n", stats.smallfiles, stats.largefiles)
	}
	fmt.Fprintf(os.Stderr, "Found %d small files and %d large files\n", stats.smallfiles, stats.largefiles)
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "walkdir: %s.\n", err)
		os.Exit(1)
	}
}
