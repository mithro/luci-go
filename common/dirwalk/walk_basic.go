// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package dirwalk

import (
	"os"
	"path/filepath"
	"strings"
)

// WalkBasic is the trivial implementation of a directory tree walker using
// built in filepath.Walk function.
func WalkBasic(root string, callback WalkFunc) {
	dirs := newStringStack()
	dirs.push("")
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			callback(path, -1, nil, err)
			return nil
		}

		for true {
			if strings.HasPrefix(path, dirs.peek()) {
				break
			}
			callback(dirs.pop(), -1, nil, nil)
		}

		if info.IsDir() {
			dirs.push(path)
		} else {
			f, err := os.Open(path)
			if err != nil {
				callback(path, -1, nil, err)
				return nil
			}
			callback(path, info.Size(), f, nil)
		}
		return nil
	})
	for true {
		if dirs.peek() == "" {
			break
		}
		callback(dirs.pop(), -1, nil, nil)
	}
}
