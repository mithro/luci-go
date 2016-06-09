// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package dirwalk

import "io"

// WalkFunc is a callback for receiving the results of walking a directory
// tree.
//
// r.Close() should be called if r is not nil.
//
// Callback for a directory will have a nil reader and occur *after* all the
// paths contained inside the directory have already been provided.
type WalkFunc func(path string, size int64, r io.ReadCloser, err error)
