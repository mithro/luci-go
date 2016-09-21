// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"io/ioutil"
)

// ReadWalker implements Walker. It reads the contents of each found file.
type ReadWalker struct {
	NullWalker
}

func (r *ReadWalker) SmallFile(filename string, alldata []byte) {
	r.NullWalker.SmallFile(filename, alldata)
}
func (r *ReadWalker) LargeFile(filename string) {
	r.NullWalker.LargeFile(filename)
	_, err := ioutil.ReadFile(filename)
	if err != nil {
		r.Error(filename, err)
	}
}
