// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"io/ioutil"
)

// ReadFileProcessor implements FileProcessor. It reads the contents of each found file.
type ReadFileProcessor struct {
	BaseFileProcessor
}

func (r *ReadFileProcessor) SmallFile(filename string, alldata []byte) {
	r.BaseFileProcessor.SmallFile(filename, alldata)
}

func (r *ReadFileProcessor) LargeFile(filename string) {
	r.BaseFileProcessor.LargeFile(filename)
	_, err := ioutil.ReadFile(filename)
	if err != nil {
		r.Error(filename, err)
	}
}
