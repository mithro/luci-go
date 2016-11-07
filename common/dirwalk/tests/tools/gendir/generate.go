// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

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
	blockSize uint64 = 1 * 1024 * 1024 // 1 Megabyte
)

func writeFile(fileName string, fileContent io.Reader, fileSize uint64) {
	f, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var written uint64
	for written < fileSize {
		content := make([]byte, min(fileSize-written, blockSize))

		// Generate a block of content
		read, err := fileContent.Read(content)
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
	fileNameMinSize uint64 = 4
	fileNameMaxSize uint64 = 20
)

type fileType int

const (
	fileTypeBinaryRandom   fileType = iota // Truly random binary data (totally uncompressible)
	fileTypeTextRandom                     // Truly random text data (mostly uncompressible)
	fileTypeBinaryRepeated                 // Repeated binary data (compressible)
	fileTypeTextRepeated                   // Repeated text data (very compressible)
	fileTypeTextLorem                      // Lorem Ipsum txt data (very compressible)

	fileTypeMax
)

var fileTypeName = []string{
	"Random Binary",
	"Random Text",
	"Repeated Binary",
	"Repeated Text",
	"Lorem Text",
}

func (f fileType) String() string {
	return fileTypeName[int(f)]
}

// Generate num files between (min, max) size
func generateFiles(r *rand.Rand, dir string, num uint64, fileSizeMin uint64, fileSizeMax uint64) {
	for i := uint64(0); i < num; i++ {
		var fileName string
		var filePath string
		for true {
			fileName = fileNameRandom(r, randBetween(r, fileNameMinSize, fileNameMaxSize))
			filePath = path.Join(dir, fileName)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				break
			}
		}
		fileSize := randBetween(r, fileSizeMin, fileSizeMax)
		filetype := fileType(r.Intn(int(fileTypeMax)))

		var fileContent io.Reader
		switch filetype {
		case fileTypeBinaryRandom:
			fileContent = RandomBinaryGenerator(r)
		case fileTypeTextRandom:
			fileContent = RandomTextGenerator(r)
		case fileTypeBinaryRepeated:
			fileContent = RepeatedBinaryGenerator(r)
		case fileTypeTextRepeated:
			fileContent = RepeatedTextGenerator(r)
		case fileTypeTextLorem:
			fileContent = LoremTextGenerator()
		}

		if num < 1000 {
			fmt.Printf("File: %-40s %-20s (%s)\n", fileName, filetype.String(), humanize.Bytes(fileSize))
		}
		writeFile(filePath, fileContent, fileSize)
	}
}

// Generate num directories
func generateDirs(r *rand.Rand, dir string, num uint64) []string {
	var result []string

	for i := uint64(0); i < num; i++ {
		var dirname string
		var dirpath string
		for true {
			dirname = fileNameRandom(r, randBetween(r, fileNameMinSize, fileNameMaxSize))
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

type fileSettings struct {
	MinNumber uint64
	MaxNumber uint64
	MinSize   uint64
	MaxSize   uint64
}

type dirSettings struct {
	Number       []uint64
	MinFileDepth uint64
}

type treeSettings struct {
	Files []fileSettings
	Dir   dirSettings
}

func generateTreeInternal(r *rand.Rand, dir string, depth uint64, settings *treeSettings) {
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

func generateTree(r *rand.Rand, rootdir string, settings *treeSettings) {
	generateTreeInternal(r, rootdir, 0, settings)
	return
}
