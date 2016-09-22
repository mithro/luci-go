// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"io"
	"math/rand"
)

// io.Reader which produces the given byte array repetitively.
type repeatedByteGenerator struct {
	data []byte
	index int
}

func (g *repeatedByteGenerator) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = g.data[g.index]
		g.index = g.index+1 % len(g.data)
	}
	return len(p), nil
}


// io.Reader which produces a random text.
type textRandomGenerator struct {
	r *rand.Rand
}

func (g *textRandomGenerator) Read(p []byte) (n int, err error) {
	i := 0
	for {
		bytes := []byte(string(randChar(g.r, textChars)))
		if i+len(bytes) > len(p) {
			break
		}

		for j := range bytes {
			p[i+j] = bytes[j]
		}
		i += len(bytes)
	}
	return i, nil
}

// io.Reader which produces truly random binary data (totally uncompressible).
func RandomBinaryGenerator(r *rand.Rand) io.Reader {
	// rand.Rand already produces random binary data via Read()
	return r
}

// io.Reader which produces repeated binary data (some what compressible).
func RepeatedBinaryGenerator(r *rand.Rand) io.Reader {
	// Figure out how big the repeated sequence will be
	sequence_size := randBetween(r, SEQUENCE_MINSIZE, SEQUENCE_MAXSIZE)

	repeater := repeatedByteGenerator{index: 0, data:make([]byte, sequence_size)}
	r.Read(repeater.data)

	return &repeater
}

// io.Reader which produces truly random text data (mostly uncompressible).
func RandomTextGenerator(r *rand.Rand) io.Reader {
	reader := textRandomGenerator{r: r}
	return &reader
}

const (
	// Maximum 4k long repeated sequences
	SEQUENCE_MINSIZE uint64 = 16
	SEQUENCE_MAXSIZE uint64 = 4 * 1024
)

// io.Reader which produces repeated text data (very compressible).
func RepeatedTextGenerator(r *rand.Rand) io.Reader {
	// Figure out how big the repeated sequence will be
	sequence_size := randBetween(r, SEQUENCE_MINSIZE, SEQUENCE_MAXSIZE)

	repeater := repeatedByteGenerator{index: 0, data: []byte(randStr(r, sequence_size, textChars))}

	return &repeater
}

// io.Reader which produces repeated Lorem Ipsum text data (very compressible).
func LoremTextGenerator() io.Reader {
	repeater := repeatedByteGenerator{index: 0, data:[]byte(lorem)}
	return &repeater
}
