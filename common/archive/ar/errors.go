// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Write an ar archive file with BSD style filenames.

package ar

import (
	"fmt"
)

type Error interface {
	error
	Fatal() bool // Is the error fatal and the archive is now corrupted?
}

// Indicates an error with using the archive/ar API.
type UsageError struct {
	msg string
}
func (e *UsageError) Error() string {
	return fmt.Sprintf("archive/ar: usage error, %s", e.msg)
}
func (e *UsageError) Fatal() bool {
	return false
}

// Indicates an error with IO while using the archive/ar. This is always fatal.
// IOError indicates an error occurred during IO operations.
// IOError is always fatal.
type IOError struct {
	filesection string
	err error
}

func (e *IOError) Error() string {
	return fmt.Sprintf("archive/ar: io error (%s) during %s -- *archive corrupted*", e.err.Error(), e.filesection)
}

func (e *IOError) Fatal() bool {
	return true
}
