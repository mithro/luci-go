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

var method = flag.String("method", "simple", "Method used to walk the tree")
var dir = flag.String("dir", "", "Directory to walk")

//var do = flags.Choice("do", "null", ["null", "print", "read"])
var do = flag.String("do", "nothing", "Action to perform on the files")
var smallfilesize = flag.Int64("smallfilesize", 64*1024, "Size to consider a small file")
var repeat = flag.Int("repeat", 1, "Repeat the walk x times")

func main() {
	flag.Parse()

	if _, err := os.Stat(*dir); err != nil {
		log.Fatalf("Directory not found: %s", err)
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
			log.Fatalf("Invalid walk method '%s'", *method)
		}
		fmt.Printf("Found %d small files and %d large files\n", stats.smallfiles, stats.largefiles)
	}
	fmt.Fprintf(os.Stderr, "Found %d small files and %d large files\n", stats.smallfiles, stats.largefiles)
}
