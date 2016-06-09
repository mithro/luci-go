// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

// Tools for generating test directories.

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path"

	"github.com/dustin/go-humanize"
)

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
var filenameChars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-")

func filenameRandom(r *rand.Rand, length uint64) string {
	return randStr(r, length, filenameChars)
}

type DirGen interface {
	Create(seed uint64, num int)
}

type FileType int

const (
	FILETYPE_BIN_RAND   FileType = iota // Truly random binary data (totally uncompressible)
	FILETYPE_TXT_RAND                   // Truly random text data (mostly uncompressible)
	FILETYPE_BIN_REPEAT                 // Repeated binary data (compressible)
	FILETYPE_TXT_REPEAT                 // Repeated text data (very compressible)
	FILETYPE_TXT_LOREM                  // Lorem Ipsum txt data (very compressible)

	FILETYPE_MAX
)

var FileTypeName []string = []string{
	"Random Binary",
	"Random Text",
	"Repeated Binary",
	"Repeated Text",
	"Lorem Text",
}

func (f FileType) String() string {
	return FileTypeName[int(f)]
}

// FIXME: Maybe some UTF-8 characters?
var textChars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-.")

const (
	BLOCKSIZE uint64 = 1 * 1024 * 1024 // 1 Megabyte

	// Maximum 4k long repeated sequences
	SEQUENCE_MINSIZE uint64 = 16
	SEQUENCE_MAXSIZE uint64 = 4 * 1024
)

func writeFile(r *rand.Rand, filename string, filetype FileType, filesize uint64) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var written uint64 = 0
	for written < filesize {
		content := make([]byte, min(filesize-written, BLOCKSIZE))

		// Generate a block of content
		switch filetype {
		case FILETYPE_BIN_RAND:
			r.Read(content)

		case FILETYPE_TXT_RAND:
			// Runes can be multiple bytes long
			for i := 0; i < len(content); {
				bytes := []byte(string(randChar(r, textChars)))
				for j := range bytes {
					content[i+j] = bytes[j]
				}
				i += len(bytes)
			}

		case FILETYPE_BIN_REPEAT:
			var sequence []byte = make([]byte, randBetween(r, SEQUENCE_MINSIZE, SEQUENCE_MAXSIZE))
			r.Read(sequence)

			for i := range content {
				content[i] = sequence[i%len(sequence)]
			}

		case FILETYPE_TXT_REPEAT, FILETYPE_TXT_LOREM:
			var sequence []byte

			switch filetype {
			case FILETYPE_TXT_REPEAT:
				// FIXME: As runes can be multiple bytes long, this could technical
				// be longer then SEQUENCE_MAXSIZE, but don't think we care?
				sequence = []byte(randStr(r, randBetween(r, SEQUENCE_MINSIZE, SEQUENCE_MAXSIZE), textChars))
			case FILETYPE_TXT_LOREM:
				sequence = []byte(lorem)
			}

			for i := range content {
				content[i] = sequence[i%len(sequence)]
			}
		}
		f.Write(content)
		written += uint64(len(content))
	}
}

const (
	FILENAME_MINSIZE uint64 = 4
	FILENAME_MAXSIZE uint64 = 20
)

// Generate num files between (min, max) size
func GenerateFiles(r *rand.Rand, dir string, num uint64, filesize_min uint64, filesize_max uint64) {
	for i := uint64(0); i < num; i++ {
		var filename string
		var filepath string
		for true {
			filename = filenameRandom(r, randBetween(r, FILENAME_MINSIZE, FILENAME_MAXSIZE))
			filepath = path.Join(dir, filename)
			if _, err := os.Stat(filepath); os.IsNotExist(err) {
				break
			}
		}
		filetype := FileType(r.Intn(int(FILETYPE_MAX)))
		filesize := randBetween(r, filesize_min, filesize_max)

		if num < 1000 {
			fmt.Printf("File: %-40s %-20s (%s)\n", filename, filetype.String(), humanize.Bytes(filesize))
		}
		writeFile(r, filepath, filetype, filesize)
	}
}

// Generate num directories
func GenerateDirs(r *rand.Rand, dir string, num uint64) []string {
	var result []string

	for i := uint64(0); i < num; i++ {
		var dirname string
		var dirpath string
		for true {
			dirname = filenameRandom(r, randBetween(r, FILENAME_MINSIZE, FILENAME_MAXSIZE))
			dirpath = path.Join(dir, dirname)
			if _, err := os.Stat(dirpath); os.IsNotExist(err) {
				break
			}
		}

		if err := os.MkdirAll(dirpath, 0755); err != nil {
			log.Fatal(err)
		}
		result = append(result, dirpath)
	}
	return result
}

type FileSettings struct {
	MinNumber uint64
	MaxNumber uint64
	MinSize   uint64
	MaxSize   uint64
}

type DirSettings struct {
	Number       []uint64
	MinFileDepth uint64
}

type TreeSettings struct {
	Files []FileSettings
	Dir   DirSettings
}

func generateTreeInternal(r *rand.Rand, dir string, depth uint64, settings *TreeSettings) {
	fmt.Printf("%04d:%s -->\n", depth, dir)
	// Generate the files in this directory
	if depth >= settings.Dir.MinFileDepth {
		for _, files := range settings.Files {
			numfiles := randBetween(r, files.MinNumber, files.MaxNumber)
			fmt.Printf("%04d:%s: Generating %d files (between %s and %s)\n", depth, dir, numfiles, humanize.Bytes(files.MinSize), humanize.Bytes(files.MaxSize))
			GenerateFiles(r, dir, numfiles, files.MinSize, files.MaxSize)
		}
	}

	// Generate another depth of directories
	if depth < uint64(len(settings.Dir.Number)) {
		numdirs := settings.Dir.Number[depth]
		fmt.Printf("%04d:%s: Generating %d directories\n", depth, dir, numdirs)
		for _, childpath := range GenerateDirs(r, dir, numdirs) {
			generateTreeInternal(r, childpath, depth+1, settings)
		}
	}
	fmt.Printf("%04d:%s <--\n", depth, dir)
}

func GenerateTree(r *rand.Rand, rootdir string, settings *TreeSettings) {
	generateTreeInternal(r, rootdir, 0, settings)
	return
}
