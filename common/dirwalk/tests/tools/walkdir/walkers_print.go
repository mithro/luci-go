// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
)

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
