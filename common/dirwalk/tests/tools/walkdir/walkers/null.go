// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"log"
	"sync/atomic"
)

// NullWalker implements Walker. It counts the number of files of each type.
type NullWalker struct {
	smallfiles uint64
	largefiles uint64
}

func (n *NullWalker) SmallFile(filename string, alldata []byte) {
	atomic.AddUint64(&n.smallfiles, 1)
}
func (n *NullWalker) LargeFile(filename string) {
	atomic.AddUint64(&n.largefiles, 1)
}
func (n *NullWalker) Error(pathname string, err error) {
	log.Fatalf("%s:%s", pathname, err)
}
func (n *NullWalker) Finished() {
}
