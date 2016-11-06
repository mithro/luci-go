// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"

	"github.com/luci/luci-go/common/isolated"
)

// HashFileProcessor implements FileProcessor. It generates a hash for the contents of each
// found file.
type HashFileProcessor struct {
	BaseFileProcessor
	obuf io.Writer
}

func (h *HashFileProcessor) HashedFile(filename string, digest isolated.HexDigest) {
	fmt.Fprintf(h.obuf, "%s: %v\n", filename, digest)
}

func (h *HashFileProcessor) SmallFile(filename string, alldata []byte) {
	h.BaseFileProcessor.SmallFile(filename, alldata)
	h.HashedFile(filename, isolated.HashBytes(alldata))
}

func (h *HashFileProcessor) LargeFile(filename string) {
	h.BaseFileProcessor.LargeFile(filename)
	d, _ := isolated.HashFile(filename)
	h.HashedFile(filename, isolated.HexDigest(d.Digest))
}
