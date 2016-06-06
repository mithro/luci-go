// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Write an ar archive file with BSD style filenames.

package ar

import (
	"fmt"
)

// Error indicates an error and if it is fatal or not.
type Error interface {
	error
	Fatal() bool // Is the error fatal and the archive is now corrupted?
}

// UsageError indicates an error with using the archive/ar API.
type UsageError struct {
	msg string
}

func (e *UsageError) Error() string {
	return fmt.Sprintf("archive/ar: usage error, %s", e.msg)
}

// Fatal is always false for Usage.
func (e *UsageError) Fatal() bool {
	return false
}

// IOError indicates an error occurred during IO operations.
type IOError struct {
	section string
	err     error
}

func (e *IOError) Error() string {
	return fmt.Sprintf("archive/ar: io error (%s) during %s -- *archive corrupted*", e.err.Error(), e.section)
}

// Fatal is always true for IOError.
func (e *IOError) Fatal() bool {
	return true
}
