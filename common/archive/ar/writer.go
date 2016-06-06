// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Write an ar archive file with BSD style filenames.

package ar

import (
	"fmt"
	"io"
	"log"
	"os"
)

// Special UsageError that indicates trying to write after closing.
var (
	ErrWriteAfterClose = UsageError{msg: "write after file closed"}
)

// WriteTooLongError indicates trying to write the wrong amount of data into
// the archive.
// WriteTooLongError is never fatal.
type WriteTooLongError struct {
	needed int64
	got    int64
}

func (e *WriteTooLongError) Error() string {
	return fmt.Sprintf("archive/ar: invalid data length (needed %d, got %d)", e.needed, e.got)
}

// Fatal is always false on WriteToLongError.
func (e *WriteTooLongError) Fatal() bool {
	return false
}

// WriteTooLongFatalError indicates that the wrong amount of data *was* written
// into the archive.
// WriteTooLongFatalError is always fatal.
type WriteTooLongFatalError struct {
	needed int64
	got    int64
}

func (e *WriteTooLongFatalError) Error() string {
	return fmt.Sprintf("archive/ar: *archive corrupted* -- invalid data written (needed %d, got %d)", e.needed, e.got)
}

// Fatal is always true on WriteToLongError.
func (e *WriteTooLongFatalError) Fatal() bool {
	return true
}

// DefaultModifyTime is the default modification time used when no value is
// provided.
const DefaultModifyTime = 1447140471

// DefaultUser is the default user id used when no value is provided.
const DefaultUser = 1000

// DefaultGroup is the default group id used when no value is provided.
const DefaultGroup = 1000

// DefaultFileMode is the default file mode used when no value is provided.
const DefaultFileMode = 0100640 // 100640 -- Octal

type writerStage uint

const (
	writeStageHeader writerStage = iota
	writeStageBody
	writeStageClosed
)

// Writer creates a new ar archive.
type Writer struct {
	w     io.Writer
	stage writerStage

	streamSizeNeeded int64
	bodyNeedsPadding bool
}

// NewWriter creates a new ar archive.
func NewWriter(w io.Writer) (*Writer, Error) {
	if _, err := io.WriteString(w, "!<arch>\n"); err != nil {
		return nil, &IOError{section: "archive header", err: err}
	}
	return &Writer{w: w, stage: writeStageHeader}, nil
}

// Close the archive. Archive will be valid after this function is called.
func (aw *Writer) Close() Error {
	switch aw.stage {
	case writeStageHeader:
		// Good
	case writeStageBody:
		return &UsageError{msg: "currently writing a file"}
	case writeStageClosed:
		return &ErrWriteAfterClose
	default:
		log.Fatalf("unknown writer mode: %d", aw.stage)
	}
	aw.stage = writeStageClosed
	aw.w = nil
	return nil
}

func (aw *Writer) wroteBytes(numbytes int64) Error {
	if numbytes > aw.streamSizeNeeded {
		return &WriteTooLongFatalError{aw.streamSizeNeeded, numbytes}
	}

	aw.streamSizeNeeded -= numbytes
	if aw.streamSizeNeeded != 0 {
		return nil
	}

	// Padding to 16bit boundary
	if aw.bodyNeedsPadding {
		if _, err := io.WriteString(aw.w, "\n"); err != nil {
			return &IOError{section: "body padding", err: err}
		}
		aw.bodyNeedsPadding = false
	}
	aw.stage = writeStageHeader
	return nil
}

// checkCanWriteContent returns nil if the stream is in a position to write a
// stream content.
func (aw *Writer) checkCanWriteContent() Error {
	switch aw.stage {
	case writeStageHeader:
		return &UsageError{msg: "need to write header first"}
	case writeStageBody:
		// Good
		return nil
	case writeStageClosed:
		return &ErrWriteAfterClose
	default:
		log.Fatalf("unknown writer mode: %d", aw.stage)
	}
	return nil
}

// Check we have finished writing bytes
func (aw *Writer) checkFinished() Error {
	if aw.streamSizeNeeded != 0 {
		return &WriteTooLongFatalError{aw.streamSizeNeeded, -1}
	}
	return nil
}

func (aw *Writer) writePartial(section string, data []byte) Error {
	if err := aw.checkCanWriteContent(); err != nil {
		return err
	}

	datalen := int64(len(data))
	if datalen > aw.streamSizeNeeded {
		return &WriteTooLongError{aw.streamSizeNeeded, datalen}
	}

	if _, err := aw.w.Write(data); err != nil {
		return &IOError{section: section, err: err}
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
func (aw *Writer) ReaderFrom(r io.Reader) (int64, Error) {
	if err := aw.checkCanWriteContent(); err != nil {
		return -1, err
	}

	count, err := io.Copy(aw.w, r)
	if err != nil {
		return -1, &IOError{section: "body file contents", err: err}
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
// Calling with wrong size data will return WriteTooLongError but the archive
// will still be valid.
// Calling after Close will return ErrWriteAfterClose.
func (aw *Writer) WriteBytes(data []byte) Error {
	if err := aw.checkCanWriteContent(); err != nil {
		return err
	}

	if datalen := int64(len(data)); datalen != aw.streamSizeNeeded {
		return &WriteTooLongError{aw.streamSizeNeeded, datalen}
	}

	if err := aw.writePartial("body content", data); err != nil {
		return err
	}
	return aw.checkFinished()
}

func (aw *Writer) writeHeaderInternal(filepath string, size int64, modtime uint64, ownerid uint, groupid uint, filemod uint) Error {
	switch aw.stage {
	case writeStageHeader:
		// Good
	case writeStageBody:
		return &UsageError{msg: "usage error, currently writing a file."}
	case writeStageClosed:
		return &ErrWriteAfterClose
	default:
		log.Fatalf("unknown writer mode: %d", aw.stage)
	}

	// File name length prefixed with '#1/' (BSD variant), 16 bytes
	if _, err := fmt.Fprintf(aw.w, "#1/%-13d", len(filepath)); err != nil {
		return &IOError{section: "file header filepath length", err: err}
	}

	// Modtime, 12 bytes
	if _, err := fmt.Fprintf(aw.w, "%-12d", modtime); err != nil {
		return &IOError{section: "file header modtime", err: err}
	}

	// Owner ID, 6 bytes
	if _, err := fmt.Fprintf(aw.w, "%-6d", ownerid); err != nil {
		return &IOError{section: "file header owner id", err: err}
	}

	// Group ID, 6 bytes
	if _, err := fmt.Fprintf(aw.w, "%-6d", groupid); err != nil {
		return &IOError{section: "file header group id", err: err}
	}

	// File mode, 8 bytes
	if _, err := fmt.Fprintf(aw.w, "%-8o", filemod); err != nil {
		return &IOError{section: "file header file mode", err: err}
	}

	// In BSD variant, file size includes the filepath length
	aw.streamSizeNeeded = int64(len(filepath)) + size

	// File size, 10 bytes
	if _, err := fmt.Fprintf(aw.w, "%-10d", aw.streamSizeNeeded); err != nil {
		return &IOError{section: "file header file size", err: err}
	}

	// File magic, 2 bytes
	if _, err := io.WriteString(aw.w, "\x60\n"); err != nil {
		return &IOError{section: "file header file magic", err: err}
	}

	aw.stage = writeStageBody
	aw.bodyNeedsPadding = (aw.streamSizeNeeded%2 != 0)

	// File path - BSD variant
	return aw.writePartial("body filepath", []byte(filepath))
}

// WriteHeaderDefault writes header information about a file to the archive
// using default values for everything apart from name and size.
// WriteBytes or ReaderFrom should be called after writing the header.
// Calling at the wrong time will return a UsageError.
func (aw *Writer) WriteHeaderDefault(filepath string, size int64) Error {
	return aw.writeHeaderInternal(filepath, size, DefaultModifyTime, DefaultUser, DefaultGroup, DefaultFileMode)
}

// WriteHeader writes header information about a file to the archive using the
// information from os.Stat.
// WriteBytes or ReaderFrom should be called after writing the header.
// Calling at the wrong time will return a UsageError.
func (aw *Writer) WriteHeader(stat os.FileInfo) Error {
	if stat.IsDir() {
		return &UsageError{msg: "only work with files, not directories"}
	}

	mode := stat.Mode()
	if mode&os.ModeSymlink == os.ModeSymlink {
		return &UsageError{msg: "only work with files, not symlinks"}
	}

	/* TODO(mithro): Should we also exclude other "special" files?
	if (stat.Mode().ModeType != 0) {
		return &argError{stat, "Only work with plain files."}
	}
	*/

	return aw.writeHeaderInternal(stat.Name(), stat.Size(), uint64(stat.ModTime().Unix()), DefaultUser, DefaultGroup, uint(mode&os.ModePerm))
}

// AddWithContent a file with given content to archive.
// Function is equivalent to calling "WriteHeaderDefault(filepath, len(data); WriteBytes(data)".
func (aw *Writer) AddWithContent(filepath string, data []byte) Error {
	if err := aw.WriteHeaderDefault(filepath, int64(len(data))); err != nil {
		return err
	}

	return aw.WriteBytes(data)
}
