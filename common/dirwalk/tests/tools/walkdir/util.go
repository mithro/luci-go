// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"io"

	"github.com/luci/luci-go/common/isolated"
)

func hash(src io.Reader) (isolated.HexDigest, uint64, error) {
	h := isolated.GetHash()
	bytes, err := io.Copy(h, src)
	if err != nil {
		return isolated.HexDigest(""), 0, err
	}
	return isolated.Sum(h), uint64(bytes), nil
}
