// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
)

// PrintFileProcessor implements FileProcessor. It prints the path of each found file.
type PrintFileProcessor struct {
	BaseFileProcessor
	obuf io.Writer
}

func (p *PrintFileProcessor) Dir(path string) {
	fmt.Fprintln(p.obuf, "  dir", path)
	p.BaseFileProcessor.Dir(path)
}

func (p *PrintFileProcessor) SmallFile(path string, r io.ReadCloser) {
	fmt.Fprintln(p.obuf, "small", path)
	p.BaseFileProcessor.SmallFile(path, r)
}

func (p *PrintFileProcessor) LargeFile(path string, r io.ReadCloser) {
	fmt.Fprintln(p.obuf, "large", path)
	p.BaseFileProcessor.LargeFile(path, r)
}
