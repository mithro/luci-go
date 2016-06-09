// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"io"
	"log"
	"sync/atomic"
)

type FileProcessor interface {
	Callback(path string, size int64, r io.ReadCloser, err error) error

	SmallFile(path string, r io.ReadCloser)
	LargeFile(path string, r io.ReadCloser)

	Error(path string, err error)

	Finished()
}

// BaseFileProcessor implements Walker. It counts the number of files of each type.
type BaseFileProcessor struct {
	smallfile_size int64
	smallfiles     uint64
	largefiles     uint64
}

func (n *BaseFileProcessor) Callback(path string, size int64, r io.ReadCloser, err error) {
	if err != nil {
		n.Error(path, err)
		return
	}
	if r != nil {
		// Ignore directories
		return
	}

	if size > 0 && size < n.smallfile_size {
		n.SmallFile(path, r)
	} else {
		n.LargeFile(path, r)
	}
}

func (n *BaseFileProcessor) SmallFile(path string, r io.ReadCloser) {
	atomic.AddUint64(&n.smallfiles, 1)
	r.Close()
}
func (n *BaseFileProcessor) LargeFile(path string, r io.ReadCloser) {
	atomic.AddUint64(&n.largefiles, 1)
	r.Close()
}
func (n *BaseFileProcessor) Error(path string, err error) {
	log.Fatalf("%s:%s", path, err)
}
func (n *BaseFileProcessor) Finished() {
}
