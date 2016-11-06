// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"os"
)

// SizeFileProcessor implements FileProcessor. It prints the size of every file.
type SizeFileProcessor struct {
	BaseFileProcessor
	obuf io.Writer
}

func (s *SizeFileProcessor) SizeFile(filename string, size int64) {
	fmt.Fprintf(s.obuf, "%s: %d\n", filename, size)
}

func (s *SizeFileProcessor) SmallFile(filename string, alldata []byte) {
	s.BaseFileProcessor.SmallFile(filename, alldata)
	s.SizeFile(filename, int64(len(alldata)))
}

func (s *SizeFileProcessor) LargeFile(filename string) {
	s.BaseFileProcessor.LargeFile(filename)
	stat, err := os.Stat(filename)
	if err != nil {
		s.Error(filename, err)
	} else {
		s.SizeFile(filename, stat.Size())
	}
}
