// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

// Tools for generating test directories.

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path"

	"github.com/dustin/go-humanize"
)

const (
	BLOCKSIZE uint64 = 1 * 1024 * 1024 // 1 Megabyte
)

func writeFile(filename string, filecontent io.Reader, filesize uint64) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var written uint64 = 0
	for written < filesize {
		content := make([]byte, min(filesize-written, BLOCKSIZE))

		// Generate a block of content
		read, err := filecontent.Read(content)
		if err != nil {
			log.Fatal(err)
		}

		// Write the block of content
		write, err := f.Write(content[0:read])
		if err != nil {
			log.Fatal(err)
		}

		written += uint64(write)
	}
}

const (
	FILENAME_MINSIZE uint64 = 4
	FILENAME_MAXSIZE uint64 = 20
)

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

// Generate num files between (min, max) size
func generateFiles(r *rand.Rand, dir string, num uint64, filesize_min uint64, filesize_max uint64) {
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
		filesize := randBetween(r, filesize_min, filesize_max)
		filetype := FileType(r.Intn(int(FILETYPE_MAX)))

		var filecontent io.Reader
		switch filetype {
		case FILETYPE_BIN_RAND:
			filecontent = RandomBinaryGenerator(r)
		case FILETYPE_TXT_RAND:
			filecontent = RandomTextGenerator(r)
		case FILETYPE_BIN_REPEAT:
			filecontent = RepeatedBinaryGenerator(r)
		case FILETYPE_TXT_REPEAT:
			filecontent = RepeatedTextGenerator(r)
		case FILETYPE_TXT_LOREM:
			filecontent = LoremTextGenerator()
		}

		if num < 1000 {
			fmt.Printf("File: %-40s %-20s (%s)\n", filename, filetype.String(), humanize.Bytes(filesize))
		}
		writeFile(filepath, filecontent, filesize)
	}
}

// Generate num directories
func generateDirs(r *rand.Rand, dir string, num uint64) []string {
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
			generateFiles(r, dir, numfiles, files.MinSize, files.MaxSize)
		}
	}

	// Generate another depth of directories
	if depth < uint64(len(settings.Dir.Number)) {
		numdirs := settings.Dir.Number[depth]
		fmt.Printf("%04d:%s: Generating %d directories\n", depth, dir, numdirs)
		for _, childpath := range generateDirs(r, dir, numdirs) {
			generateTreeInternal(r, childpath, depth+1, settings)
		}
	}
	fmt.Printf("%04d:%s <--\n", depth, dir)
}

func GenerateTree(r *rand.Rand, rootdir string, settings *TreeSettings) {
	generateTreeInternal(r, rootdir, 0, settings)
	return
}
