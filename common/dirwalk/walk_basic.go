// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package dirwalk

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// Trivial implementation of a directory tree walker using the WalkObserver
// interface.
func WalkBasic(root string, smallfile_limit int64, obs WalkObserver) {
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			obs.Error(path, err)
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if info.Size() < smallfile_limit {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				obs.Error(path, err)
				return nil
			}
			if int64(len(data)) != info.Size() {
				panic("file size was wrong!")
			}
			obs.SmallFile(path, data)
		} else {
			obs.LargeFile(path)
		}
		return nil
	})
	obs.Finished()
}
