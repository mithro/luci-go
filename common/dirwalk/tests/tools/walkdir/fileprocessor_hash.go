// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
)

// HashFileProcessor implements FileProcessor. It generates a hash for the contents of each
// found file.
type HashFileProcessor struct {
	BaseFileProcessor
	obuf io.Writer
}

func (p *HashFileProcessor) HashedFile(path string, r io.Reader) {
	digest, _, err := hash(r)
	if err != nil {
		p.Error(path, err)
		return
	}
	fmt.Fprintf(p.obuf, "%s: %v\n", path, digest)
}

func (p *HashFileProcessor) SmallFile(path string, r io.ReadCloser) {
	p.HashedFile(path, r)
	p.BaseFileProcessor.SmallFile(path, r)
}

func (p *HashFileProcessor) LargeFile(path string, r io.ReadCloser) {
	p.HashedFile(path, r)
	p.BaseFileProcessor.LargeFile(path, r)
}
