// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"io"
	"log"
	"os"
	"strings"
)

// VerifyFileProcessor implements FileProcessor. It verifies that things are called in the right order.
type VerifyFileProcessor struct {
	BaseFileProcessor
	rootDir      string
	finishedDirs []string
}

func (p *VerifyFileProcessor) check(path string) {
	for _, d := range p.finishedDirs {
		if strings.HasPrefix(path, d) {
			log.Fatalf("Directory called before content\n    Dir: %s\nContent: %s\n", path, d)
		}
	}
}

func (p *VerifyFileProcessor) checkFile(path string) {
	stat, err := os.Stat(path)
	if err != nil {
		log.Fatalf("Error while stating small file\n%s", path)
	}
	if stat.IsDir() {
		log.Fatalf("File function called with directory!\n%s", path)
	}
	p.check(path)
}

func (p *VerifyFileProcessor) Dir(path string) {
	stat, err := os.Stat(path)
	if err != nil {
		log.Fatalf("Error while stating directory\n%s", path)
	}
	if !stat.IsDir() {
		log.Fatalf("Dir() called with non-directory!\n%s", path)
	}
	p.check(path)
	p.finishedDirs = append(p.finishedDirs, path)
	p.BaseFileProcessor.Dir(path)
}

func (p *VerifyFileProcessor) SmallFile(path string, r io.ReadCloser) {
	p.checkFile(path)
	p.BaseFileProcessor.SmallFile(path, r)
}

func (p *VerifyFileProcessor) LargeFile(path string, r io.ReadCloser) {
	p.checkFile(path)
	p.BaseFileProcessor.LargeFile(path, r)
}

func (p *VerifyFileProcessor) Finished() {
	if strings.Compare(p.finishedDirs[len(p.finishedDirs)-1], p.rootDir) != 0 {
		log.Fatalf("Last directory should be the root\n%s\n%s", p.finishedDirs[len(p.finishedDirs)-1], p.rootDir)
	}
}
