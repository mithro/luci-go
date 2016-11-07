// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// SizeFileProcessor implements FileProcessor. It prints the size of every file.
type SizeFileProcessor struct {
	BaseFileProcessor
	obuf io.Writer
}

func (p *SizeFileProcessor) SizeFile(path string, size int64) {
	fmt.Fprintf(p.obuf, "%s: %d\n", path, size)
}

func (p *SizeFileProcessor) SmallFile(path string, r io.ReadCloser) {
	bytes, err := io.Copy(ioutil.Discard, r)
	if err != nil {
		p.Error(path, err)
	}
	p.SizeFile(path, int64(bytes))

	p.BaseFileProcessor.SmallFile(path, r)
}

func (p *SizeFileProcessor) LargeFile(path string, r io.ReadCloser) {
	fi, err := os.Stat(path)
	if err != nil {
		p.Error(path, err)
	} else {
		p.SizeFile(path, fi.Size())
	}
	p.BaseFileProcessor.LargeFile(path, r)
}
