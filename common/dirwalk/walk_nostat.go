// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// This function works strangely for performance reasons.
//
// File systems have been heavily optimised for doing a directory walk in inode
// order. It can be an order of magnitude faster to walk the directory in this
// order so we do. *However*, we want out output to be in sorted so it is
// deterministic.
//
// Calling `stat` is also one of the most expensive things you can do (it is
// roughly equivalent to reading 64/128k of data). Hence, if you have a lot of
// small files then just reading their contents directly is more efficient.
// Rather then doing the stat, we assume everything is a file and just try to
// read a chunk. If the file is smaller than the block size, we know that we
// have the entire contents. Otherwise we know the file is bigger and can
// decide to do the stat. If the name turned out to be a directory, then we
// will get an error.

package dirwalk

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func walkNoStatInternal(base string, files []string, smallfile_limit int64, callback WalkFunc) {
	for _, name := range files {
		path := filepath.Join(base, name)

		file, err := os.Open(path)
		if err != nil {
			callback(path, -1, nil, err)
			continue
		}

		block := make([]byte, smallfile_limit)
		count, err := file.Read(block)
		if err != io.EOF && err != nil {
			// It is probably a directory, try and list it.
			dir := file

			names, err := dir.Readdirnames(0)
			if err != nil {
				callback(path, -1, nil, err)
				continue
			}
			walkNoStatInternal(path, names, smallfile_limit, callback)
			callback(path, -1, nil, nil)
		}

		if int64(count) == smallfile_limit {
			// FIXME: Need to actually close file properly!!
			callback(path, -1, ioutil.NopCloser(io.MultiReader(bytes.NewReader(block), file)), nil)
		} else {
			// This file was smaller than the block size
			callback(path, int64(count), ioutil.NopCloser(bytes.NewReader(block[:count])), nil)
		}
	}
}

func WalkNoStat(root string, smallfile_limit int64, callback WalkFunc) {
	paths := []string{root}
	walkNoStatInternal("", paths, smallfile_limit, callback)
	callback(root, -1, nil, nil)
}
