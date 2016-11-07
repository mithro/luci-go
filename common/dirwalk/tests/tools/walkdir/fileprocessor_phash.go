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

	"github.com/dustin/go-humanize"
)

var maxworkers = flag.Int("maxworkers", 100, "Maximum number of workers to use.")

type ToHash struct {
	path string
	r    io.ReadCloser
}

// ParallelHashFileProcessor implements FileProcessor. It generates a hash for the contents
// of each found file using multiple threads.
type ParallelHashFileProcessor struct {
	BaseFileProcessor
	obuf     io.Writer
	workers  int
	queue    *chan ToHash
	finished chan bool
}

func ParallelHashFileProcessorWorker(name int, obuf io.Writer, queue <-chan ToHash, finished chan<- bool) {
	fmt.Fprintf(obuf, "Starting hash worker %d\n", name)

	var filecount uint64
	var bytecount uint64
	for tohash := range queue {
		filecount++

		digest, bytes, err := hash(tohash.r)
		tohash.r.Close()
		if err != nil {
			// FIXME(mithro): Do something here?
			continue
		}
		bytecount += bytes
		fmt.Fprintf(obuf, "%s: %v\n", tohash.path, digest)
	}
	fmt.Fprintf(obuf, "Finished hash worker %d (hashed %d files, %s)\n", name, filecount, humanize.Bytes(bytecount))
	finished <- true
}

func CreateParallelHashFileProcessor(obuf io.Writer) *ParallelHashFileProcessor {
	max := *maxworkers

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

	p := ParallelHashFileProcessor{obuf: obuf, workers: max, finished: make(chan bool)}
	q := make(chan ToHash, p.workers)
	p.queue = &q
	for i := 0; i < p.workers; i++ {
		go ParallelHashFileProcessorWorker(i, p.obuf, *p.queue, p.finished)
	}
	return &p
}

func (p *ParallelHashFileProcessor) SmallFile(path string, r io.ReadCloser) {
	*p.queue <- ToHash{path: path, r: r}
	p.BaseFileProcessor.SmallFile(path, ioutil.NopCloser(r))
}

func (p *ParallelHashFileProcessor) LargeFile(path string, r io.ReadCloser) {
	*p.queue <- ToHash{path: path, r: r}
	p.BaseFileProcessor.LargeFile(path, ioutil.NopCloser(r))
}

func (p *ParallelHashFileProcessor) Complete(path string) {
	close(*p.queue)
	for i := 0; i < p.workers; i++ {
		<-p.finished
	}
	fmt.Fprintln(p.obuf, "All workers finished.")
	p.queue = nil
}
