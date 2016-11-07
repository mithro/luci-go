// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"math/rand"
)

var textChars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-.")

func randChar(r *rand.Rand, runes []rune) rune {
	return runes[r.Intn(len(runes))]
}

func randStr(r *rand.Rand, length uint64, runes []rune) string {
	str := make([]rune, length)
	for i := range str {
		str[i] = randChar(r, runes)
	}
	return string(str)
}

func randBetween(r *rand.Rand, min uint64, max uint64) uint64 {
	if min == max {
		return min
	}
	return uint64(r.Int63n(int64(max-min))) + min
}

// FIXME: Maybe some UTF-8 characters?
var fileNameChars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-")

func fileNameRandom(r *rand.Rand, length uint64) string {
	return randStr(r, length, fileNameChars)
}
