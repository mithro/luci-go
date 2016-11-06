// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
)

// PrintFileProcessor implements FileProcessor. It prints the filename of each found file.
type PrintFileProcessor struct {
	BaseFileProcessor
	obuf io.Writer
}

func (p *PrintFileProcessor) PrintFile(filename string) {
	fmt.Fprintln(p.obuf, filename)
}

func (p *PrintFileProcessor) SmallFile(filename string, alldata []byte) {
	p.BaseFileProcessor.SmallFile(filename, alldata)
	p.PrintFile(filename)
}

func (p *PrintFileProcessor) LargeFile(filename string) {
	p.BaseFileProcessor.LargeFile(filename)
	p.PrintFile(filename)
}
