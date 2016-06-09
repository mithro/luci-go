// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package dirwalk

// Interface for receiving the results of walking a directory tree.
//
// For performance reasons, small files and large files are treated
// differently.
//
// SmallFile and LargeFile must be called in sorted order.

type WalkObserver interface {
	SmallFile(filename string, alldata []byte)
	LargeFile(filename string)

	//StartDir(dirname string) error
	//FinishDir(dirname string)

	Error(pathname string, err error)

	Finished()
}
