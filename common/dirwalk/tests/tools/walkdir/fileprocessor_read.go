// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"io"
	"io/ioutil"
)

// ReadFileProcessor implements FileProcessor. It reads the contents of each found file.
type ReadFileProcessor struct {
	BaseFileProcessor
}

func (p *ReadFileProcessor) ReadFile(path string, r io.Reader) {
	_, err := io.Copy(ioutil.Discard, r)
	if err != nil {
		p.Error(path, err)
	}
}

func (p *ReadFileProcessor) SmallFile(path string, r io.ReadCloser) {
	p.ReadFile(path, r)
	p.BaseFileProcessor.SmallFile(path, r)
}

func (p *ReadFileProcessor) LargeFile(path string, r io.ReadCloser) {
	p.ReadFile(path, r)
	p.BaseFileProcessor.LargeFile(path, r)
}
