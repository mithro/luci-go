// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"runtime"

	"github.com/luci/luci-go/common/isolated"
)

var maxworkers = flag.Int("maxworkers", 100, "Maximum number of workers to use.")

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
