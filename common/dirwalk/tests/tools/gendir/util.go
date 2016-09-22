// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

func min(a uint64, b uint64) uint64 {
	if a > b {
		return b
	}
	return a
}
func max(a uint64, b uint64) uint64 {
	if a < b {
		return b
	}
	return a
}
