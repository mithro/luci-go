// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Write an ar archive file with BSD style filenames.

package ar

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

type Error interface {
	error
	Fatal() bool	// Is the error fatal and the archive is now corrupted?
}

type ErrWriteUsage struct {
	msg string
}
func (e *ErrWriteUsage) Error() string {
	return fmt.Sprintf("archive/ar: usage error, %s", e.msg)
}
func (e *ErrWriteUsage) Fatal() bool {
	return false
}

// ErrWriteIOError indicates an error occurred during IO operations.
// ErrWriteIOError is always fatal.
type ErrWriteIOError struct {
	err Error
}
func (e *ErrWriteIOError) Error() string {
	return fmt.Sprintf("archive/ar: *archive corrupted* -- invalid data written (%s)", e.err.String())
}
func (e *ErrWriteIOError) Fatal() bool {
	return true
}

// ErrWriteIOError indicates trying to write to a file after closing.
// ErrWriteIOError is never fatal.
type ErrWriteAfterClose struct {
	ErrWriteUsage
}

// ErrWriteToLong indicates trying to write the wrong amount of data into the archive.
// ErrWriteIOError is never fatal.
type ErrWriteToLong struct {
	needed int64,
	got int64,
}
func (e *ErrWriteToLong) Error() string {
	return fmt.Sprintf("archive/ar: invalid data length (needed %d, got %d)", e.requested, e.got)
}
func (e *ErrWriteToLong) Fatal() bool {
	return false
}

// ErrWriteToLongFatal indicates that the wrong amount of data *was* written into the archive.
// ErrWriteIOError is always fatal.
type ErrWriteToLongFatal struct {
	needed int64,
	got int64,
}
func (e *ErrWriteToLongFatal) Error() string {
	return fmt.Sprintf("archive/ar: *archive corrupted* -- invalid data written (needed %d, got %d)", e.requested, e.got)
}
func (e *ErrWriteToLongFatal) Fatal() bool {
	return true
}

const DefaultModifyTime = 1447140471
const DefaultUser = 1000
const DefaultGroup = 1000
const DefaultFileMode = 0100640 // 100640 -- Octal

type writerStage uint

const (
	writeStageHeader writerStage = iota
	writeStageBody
	writeStageClosed
)

type Writer struct {
	w     io.Writer
	stage writerStage

	bytesrequired int64
	needspadding  bool
}

func NewWriter(w io.Writer) *Writer, Error {
	if err := io.WriteString(w, "!<arch>\n"); err != nil {
		return nil, ErrWriteIOError{err: err}
	}
	return &Writer{w: w, stage: writeStageHeader}, nil
}

func (aw *Writer) Close() Error {
	switch aw.stage {
	case writeStageHeader:
		// Good
	case writeStageBody:
		return ErrWriteUsage{msg: "currently writing a file"}
	case writeStageClosed:
		return ErrWriteAfterClose{}
	default:
		log.Fatalf("unknown writer mode: %d", aw.stage)
	}
	aw.stage = writeStageClosed
	aw.w = nil
	return nil
}

func (aw *Writer) wroteBytes(numbytes int64) Error {
	if numbytes > aw.bytesrequired {
		return ErrWriteToLongFatal{aw.bytesrequired, numbytes}
	}

	aw.bytesrequired -= numbytes
	if aw.bytesrequired != 0 {
		return nil
	}

	// Padding to 16bit boundary
	if aw.needspadding {
		if _, err := io.WriteString(aw.w, "\n"); err != nil {
			return ErrWriteIOError{err: err}
		}
		aw.needspadding = false
	}
	aw.stage = writeStageHeader
	return nil
}

// canWriteContent returns nil if the stream is in a position to write a stream
// content.
func (aw *Writer) checkWrite() Error {
	switch aw.stage {
	case writeStageHeader:
		return ErrWriteUsage{msg: "need to write header first"}
		// Good
	case writeStageBody:
		return nil
	case writeStageClosed:
		return ErrWriteAfterClose{}
	default:
		log.Fatalf("unknown writer mode: %d", aw.stage)
	}
	return nil
}

// Check we have finished writing bytes
func (aw *Writer) checkFinished() Error {
	if aw.bytesrequired != 0 {
		return ErrWriteToLongFatal{aw.bytesrequired, -1}
	}
}

func (aw *Writer) writePartial(data []byte) Error {
	if err := aw.checkWrite(); err != nil {
		return err
	}

	datalen := int64(len(data))
	if datalen > aw.bytesrequired {
		return ErrWriteToLong{aw.bytesrequired, datalen}
	}

	if err := aw.w.Write(data); err != nil {
		return ErrWriteIOError{err: err}
	}
	if err := aw.wroteBytes(datalen); err != nil {
		return err
	}
	return nil
}


// ReaderFrom writes all the data from r (till EOF) into the archive. 
// The size of data should match the value given previously to WriteHeader*
// functions.
// ReaderFrom returns the number of bytes written on success.
// Calling with wrong size data will return ErrFatalWriteToLong, the archive
// should be considered broken.
// Calling after Close will return ErrWriteAfterClose
func (aw *Writer) ReaderFrom(r io.Reader) int64, Error {
	if err := aw.checkWrite(); err != nil {
		return -1, err
	}

	count, err := io.Copy(aw.w, data)
	if err != nil {
		return ErrWriteIOError{err: err}
	}
	if err := aw.wroteBytes(count); err != nil {
		return -1, err
	}
	if err := aw.checkFinished(); err != nil {
		return -1, err
	}

	return count, nil
}

// WriteBytes writes the given byte data into the archive.
// The size of data array should match the value given previously to
// WriteHeader* functions.
// WriteBytes returns nil on success.
// Calling with wrong size data will return ErrWriteToLong but the archive will
// still be valid.
// Calling after Close will return ErrWriteAfterClose.
func (aw *Writer) WriteBytes(data []byte) error {
	if err := aw.checkWrite(); err != nil {
		return err
	}

	if datalen := int64(len(data)); datalen != aw.bytesrequired {
		return ErrWriteToLong{aw.bytesrequired, datalen}
	}

	if err := aw.writePartial(data); err != nil {
		return err
	}
	if err := aw.checkFinished(); err != nil {
		return err
	}
	return nil
}

func (aw *Writer) writeHeaderInternal(name string, size int64, modtime uint64, ownerid uint, groupid uint, filemod uint) Error {
	switch aw.stage {
	case writeStageHeader:
		// Good
	case writeStageBody:
		return ErrWriteUsage{msg: "usage error, currently writing a file."}
	case writeStageClosed:
		return ErrWriteAfterClose{}
	default:
		log.Fatalf("unknown writer mode: %d", aw.stage)
	}

	// File name length prefixed with '#1/' (BSD variant), 16 bytes
	if _, err := fmt.Fprintf(aw.w, "#1/%-13d", len(name)); err != nil {
		return ErrWriteIOError{err: err}
	}

	// Modtime, 12 bytes
	if _, err := fmt.Fprintf(aw.w, "%-12d", modtime); err != nil {
		return ErrWriteIOError{err: err}
	}

	// Owner ID, 6 bytes
	if _, err := fmt.Fprintf(aw.w, "%-6d", ownerid); err != nil {
		return ErrWriteIOError{err: err}
	}

	// Group ID, 6 bytes
	if _, err := fmt.Fprintf(aw.w, "%-6d", groupid); err != nil {
		return ErrWriteIOError{err: err}
	}

	// File mode, 8 bytes
	if _, err := fmt.Fprintf(aw.w, "%-8o", filemod); err != nil {
		return ErrWriteIOError{err: err}
	}

	// In BSD variant, file size includes the filename length
	aw.bytesrequired = int64(len(name)) + size

	// File size, 10 bytes
	if _, err := fmt.Fprintf(aw.w, "%-10d", aw.bytesrequired); err != nil {
		return ErrWriteIOError{err: err}
	}

	// File magic, 2 bytes
	if _, err := io.WriteString(aw.w, "\x60\n"); err != nil {
		return ErrWriteIOError{err: err}
	}

	aw.stage = writeStageBody
	aw.needspadding = (aw.bytesrequired%2 != 0)

	// Filename - BSD variant
	return aw.writePartial([]byte(name))
}

// WriteHeaderDefault writes header information about a file to the archive
// using default values for everything apart from name and size. 
// WriteBytes or ReaderFrom should be called after writing the header.
// Calling at the wrong time will return a ErrWriteUsage.
func (aw *Writer) WriteHeaderDefault(name string, size int64) Error {
	return aw.writeHeaderInternal(name, size, DefaultModifyTime, DefaultUser, DefaultGroup, DefaultFileMode)
}

// WriteHeader writes header information about a file to the archive using the
// information from os.Stat.
// WriteBytes or ReaderFrom should be called after writing the header.
// Calling at the wrong time will return a ErrWriteUsage.
func (aw *Writer) WriteHeader(stat os.FileInfo) Error {
	if stat.IsDir() {
		return ErrWriteUsage{msg: "only work with files, not directories"}
	}

	mode := stat.Mode()
	if mode&os.ModeSymlink == os.ModeSymlink {
		return ErrWriteUsage{msg: "only work with files, not symlinks"}
	}

	/* FIXME: Should we also exclude other "special" files?
	if (stat.Mode().ModeType != 0) {
		return &argError{stat, "Only work with plain files."}
	}
	*/

	return aw.writeHeaderInternal(stat.Name(), stat.Size(), uint64(stat.ModTime().Unix()), DefaultUser, DefaultGroup, uint(mode&os.ModePerm))
}

// Add a file to the archive.
// Function is equivalent to calling "WriteHeaderDefault(name, len(data); WriteBytes(data)".
func (aw *Writer) Add(name string, data []byte) Error {
	if err := aw.WriteHeaderDefault(name, int64(len(data))); err != nil {
		return err
	}

	return aw.WriteBytes(data)
}
