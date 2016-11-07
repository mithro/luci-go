// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"io"
	"log"
	"sync/atomic"
)

// FileProcessor is a basic interface for processing files provided by the
// directory walking functions.
type FileProcessor interface {
	// Dir is called when a directory has been finished.
	Dir(path string)

	// SmallFile is called for processing a file which has been classed as "small".
	SmallFile(path string, r io.ReadCloser)

	// LargeFile is called for processing a file which has been classed as "large".
	LargeFile(path string, r io.ReadCloser)

	// Error is called when an error occurs on a path.
	Error(path string, err error)

	// Finished is called when the directory walk is finished.
	Finished()
}

// BaseFileProcessor implements FileProcessor and counts the number of files of
// each type.
type BaseFileProcessor struct {
	smallFiles uint64
	largeFiles uint64
	dirs       uint64
}

func (p *BaseFileProcessor) Dir(path string) {
	atomic.AddUint64(&p.dirs, 1)
}

func (p *BaseFileProcessor) SmallFile(path string, r io.ReadCloser) {
	atomic.AddUint64(&p.smallFiles, 1)
	if r != nil {
		r.Close()
	}
}

func (p *BaseFileProcessor) LargeFile(path string, r io.ReadCloser) {
	atomic.AddUint64(&p.largeFiles, 1)
	if r != nil {
		r.Close()
	}
}

func (p *BaseFileProcessor) Error(path string, err error) {
	log.Fatalf("%s:%s", path, err)
}

func (p *BaseFileProcessor) Finished() {
}
