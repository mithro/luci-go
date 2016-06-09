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
	"runtime"
	"sync/atomic"

	"github.com/dustin/go-humanize"
	"github.com/luci/luci-go/common/dirwalk"
	"github.com/luci/luci-go/common/isolated"
)

var method = flag.String("method", "simple", "Method used to walk the tree")
var dir = flag.String("dir", "", "Directory to walk")

//var do = flags.Choice("do", "null", ["null", "print", "read"])
var do = flag.String("do", "nothing", "Action to perform on the files")
var smallfilesize = flag.Int64("smallfilesize", 64*1024, "Size to consider a small file")
var repeat = flag.Int("repeat", 1, "Repeat the walk x times")

var maxworkers = flag.Int("maxworkers", 100, "Maximum number of workers to use.")

// NullWalker implements Walker. It counts the number of files of each type.
type NullWalker struct {
	smallfiles uint64
	largefiles uint64
}

func (n *NullWalker) SmallFile(filename string, alldata []byte) {
	atomic.AddUint64(&n.smallfiles, 1)
}
func (n *NullWalker) LargeFile(filename string) {
	atomic.AddUint64(&n.largefiles, 1)
}
func (n *NullWalker) Error(pathname string, err error) {
	log.Fatalf("%s:%s", pathname, err)
}
func (n *NullWalker) Finished() {
}

// PrintWalker implements Walker. It prints the filename of each found file.
type PrintWalker struct {
	NullWalker
	obuf io.Writer
}

func (p *PrintWalker) PrintFile(filename string) {
	fmt.Fprintln(p.obuf, filename)
}
func (p *PrintWalker) SmallFile(filename string, alldata []byte) {
	p.NullWalker.SmallFile(filename, alldata)
	p.PrintFile(filename)
}
func (p *PrintWalker) LargeFile(filename string) {
	p.NullWalker.LargeFile(filename)
	p.PrintFile(filename)
}

// SizeWalker implements Walker. It prints the filename and file size of each
// found file.
type SizeWalker struct {
	NullWalker
	obuf io.Writer
}

func (s *SizeWalker) SizeFile(filename string, size int64) {
	fmt.Fprintf(s.obuf, "%s: %d\n", filename, size)
}
func (s *SizeWalker) SmallFile(filename string, alldata []byte) {
	s.NullWalker.SmallFile(filename, alldata)
	s.SizeFile(filename, int64(len(alldata)))
}
func (s *SizeWalker) LargeFile(filename string) {
	s.NullWalker.LargeFile(filename)
	stat, err := os.Stat(filename)
	if err != nil {
		s.Error(filename, err)
	} else {
		s.SizeFile(filename, stat.Size())
	}
}

// ReadWalker implements Walker. It reads the contents of each found file.
type ReadWalker struct {
	NullWalker
}

func (r *ReadWalker) SmallFile(filename string, alldata []byte) {
	r.NullWalker.SmallFile(filename, alldata)
}
func (r *ReadWalker) LargeFile(filename string) {
	r.NullWalker.LargeFile(filename)
	_, err := ioutil.ReadFile(filename)
	if err != nil {
		r.Error(filename, err)
	}
}

// HashWalker implements Walker. It generates a hash for the contents of each
// found file.
type HashWalker struct {
	NullWalker
	obuf io.Writer
}

func (h *HashWalker) HashedFile(filename string, digest isolated.HexDigest) {
	fmt.Fprintf(h.obuf, "%s: %v\n", filename, digest)
}
func (h *HashWalker) SmallFile(filename string, alldata []byte) {
	h.NullWalker.SmallFile(filename, alldata)
	h.HashedFile(filename, isolated.HashBytes(alldata))
}
func (h *HashWalker) LargeFile(filename string) {
	h.NullWalker.LargeFile(filename)
	d, _ := isolated.HashFile(filename)
	h.HashedFile(filename, isolated.HexDigest(d.Digest))
}

// Walker which hashes using a worker tool
type ToHash struct {
	filename string
	hasdata  bool
	data     []byte
}

// ParallelHashWalker implements Walker. It generates a hash for the contents
// of each found file using multiple threads.
type ParallelHashWalker struct {
	NullWalker
	obuf     io.Writer
	workers  int
	queue    *chan ToHash
	finished chan bool
}

func ParallelHashWalkerWorker(name int, obuf io.Writer, queue <-chan ToHash, finished chan<- bool) {
	fmt.Fprintf(obuf, "Starting hash worker %d\n", name)

	var filecount uint64 = 0
	var bytecount uint64 = 0
	for tohash := range queue {
		filecount += 1

		var digest isolated.HexDigest
		if tohash.hasdata {
			bytecount += uint64(len(tohash.data))
			digest = isolated.HashBytes(tohash.data)
		} else {
			d, _ := isolated.HashFile(tohash.filename)
			bytecount += uint64(d.Size)
			digest = isolated.HexDigest(d.Digest)
		}
		fmt.Fprintf(obuf, "%s: %v\n", tohash.filename, digest)
	}
	fmt.Fprintf(obuf, "Finished hash worker %d (hashed %d files, %s)\n", name, filecount, humanize.Bytes(bytecount))
	finished <- true
}
func CreateParallelHashWalker(obuf io.Writer) *ParallelHashWalker {
	var max int = *maxworkers

	maxProcs := runtime.GOMAXPROCS(0)
	if maxProcs < max {
		max = maxProcs
	}

	numCPU := runtime.NumCPU()
	if numCPU < maxProcs {
		max = numCPU
	}

	if max < *maxworkers {
		// FIXME: Warn
	}

	h := ParallelHashWalker{obuf: obuf, workers: max, finished: make(chan bool)}
	return &h
}
func (h *ParallelHashWalker) Init() {
	if h.queue == nil {
		q := make(chan ToHash, h.workers)
		h.queue = &q
		for i := 0; i < h.workers; i++ {
			go ParallelHashWalkerWorker(i, h.obuf, *h.queue, h.finished)
		}
	}
}
func (h *ParallelHashWalker) SmallFile(filename string, alldata []byte) {
	h.NullWalker.SmallFile(filename, alldata)
	h.Init()
	*h.queue <- ToHash{filename: filename, hasdata: true, data: alldata}
}
func (h *ParallelHashWalker) LargeFile(filename string) {
	h.NullWalker.LargeFile(filename)
	h.Init()
	*h.queue <- ToHash{filename: filename, hasdata: false}
}
func (h *ParallelHashWalker) Finished() {
	h.Init()
	close(*h.queue)
	for i := 0; i < h.workers; i++ {
		<-h.finished
	}
	fmt.Fprintln(h.obuf, "All workers finished.")
	h.queue = nil
}

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
