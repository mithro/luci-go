// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

/**
 * Write an ar archive file.
 */
package ar

import (
	"errors"
	"fmt"
	"io"
	"os"
)

const DEFAULT_MODTIME = 1447140471
const DEFAULT_USER = 1000
const DEFAULT_GROUP = 1000
const DEFAULT_MODE = 0100640 // 100640 -- Octal

type WriterStage uint

const (
	WRITE_HEADER WriterStage = iota
	WRITE_BODY               = iota
	WRITE_CLOSED             = iota
)

type Writer struct {
	w     io.Writer
	stage WriterStage

	bytesrequired int64
	needspadding  bool
}

func NewWriter(w io.Writer) *Writer {
	io.WriteString(w, "!<arch>\n")
	return &Writer{w: w, stage: WRITE_HEADER, bytesrequired: 0, needspadding: false}
}

func (aw *Writer) Close() error {
	switch aw.stage {
	case WRITE_HEADER:
		// Good
	case WRITE_BODY:
		return errors.New("Usage error, writing a file.")
	case WRITE_CLOSED:
		return errors.New("Usage error, archive already closed.")
	default:
		panic(fmt.Sprintf("Unknown writer mode: %d", aw.stage))
	}
	//aw.w.Close()
	aw.stage = WRITE_CLOSED
	return nil
}

func (aw *Writer) wroteBytes(numbytes int64) error {
	if numbytes > aw.bytesrequired {
		panic(fmt.Sprintf("To much data written! Needed %d, got %d", aw.bytesrequired, numbytes))
	}

	aw.bytesrequired -= numbytes
	if aw.bytesrequired != 0 {
		return nil
	}

	// Padding to 16bit boundary
	if aw.needspadding {
		io.WriteString(aw.w, "\n")
		aw.needspadding = false
	}
	aw.stage = WRITE_HEADER
	return nil
}

// Check you can write bytes to the ar at this moment.
func (aw *Writer) checkWrite() error {
	switch aw.stage {
	case WRITE_HEADER:
		return errors.New("Usage error, need to write header first.")
		// Good
	case WRITE_BODY:
		return nil
	case WRITE_CLOSED:
		return errors.New("Usage error, archive closed.")
	default:
		panic(fmt.Sprintf("Unknown writer mode: %d", aw.stage))
	}
}

// Check we have finished writing bytes
func (aw *Writer) checkFinished() {
	if aw.bytesrequired != 0 {
		panic(fmt.Sprintf("Didn't write enough bytes %d still needed, archive corrupted!", aw.bytesrequired))
	}
}

func (aw *Writer) writePartial(data []byte) error {
	err := aw.checkWrite()
	if err != nil {
		return err
	}

	datalen := int64(len(data))
	if datalen > aw.bytesrequired {
		return errors.New(fmt.Sprintf("To much data! Needed %d, got %d", aw.bytesrequired, datalen))
	}

	aw.w.Write(data)
	aw.wroteBytes(datalen)
	return nil
}

func (aw *Writer) WriteReader(data io.Reader) error {
	err := aw.checkWrite()
	if err != nil {
		return err
	}

	count, err := io.Copy(aw.w, data)
	if err != nil {
		panic(fmt.Sprintf("err while copying (%s), archive is probably corrupted!", err))
	}
	aw.wroteBytes(count)
	aw.checkFinished()

	return nil
}

func (aw *Writer) WriteBytes(data []byte) error {
	err := aw.checkWrite()
	if err != nil {
		return err
	}

	datalen := int64(len(data))
	if datalen != aw.bytesrequired {
		return errors.New(fmt.Sprintf("Wrong amount of data! Needed %d, got %d", aw.bytesrequired, datalen))
	}

	aw.writePartial(data)
	aw.checkFinished()
	return nil
}

func (aw *Writer) writeHeaderBytes(name string, size int64, modtime uint64, ownerid uint, groupid uint, filemod uint) error {
	switch aw.stage {
	case WRITE_HEADER:
		// Good
	case WRITE_BODY:
		return errors.New("Usage error, already writing a file.")
	case WRITE_CLOSED:
		return errors.New("Usage error, archive closed.")
	default:
		panic(fmt.Sprintf("Unknown writer mode: %d", aw.stage))
	}

	// File name length prefixed with '#1/' (BSD variant), 16 bytes
	fmt.Fprintf(aw.w, "#1/%-13d", len(name))

	// Modtime, 12 bytes
	fmt.Fprintf(aw.w, "%-12d", modtime)

	// Owner ID, 6 bytes
	fmt.Fprintf(aw.w, "%-6d", ownerid)

	// Group ID, 6 bytes
	fmt.Fprintf(aw.w, "%-6d", groupid)

	// File mode, 8 bytes
	fmt.Fprintf(aw.w, "%-8o", filemod)

	// In BSD variant, file size includes the filename length
	aw.bytesrequired = int64(len(name)) + size

	// File size, 10 bytes
	fmt.Fprintf(aw.w, "%-10d", aw.bytesrequired)

	// File magic, 2 bytes
	io.WriteString(aw.w, "\x60\n")

	aw.stage = WRITE_BODY
	aw.needspadding = (aw.bytesrequired%2 != 0)

	// Filename - BSD variant
	return aw.writePartial([]byte(name))
}

func (aw *Writer) WriteHeaderDefault(name string, size int64) error {
	return aw.writeHeaderBytes(name, size, 1447140471, DEFAULT_USER, DEFAULT_GROUP, DEFAULT_MODE)
}

func (aw *Writer) WriteHeader(stat os.FileInfo) error {
	if stat.IsDir() {
		return errors.New("Only work with files, not directories.")
	}

	mode := stat.Mode()
	if mode&os.ModeSymlink == os.ModeSymlink {
		return errors.New("Only work with files, not symlinks.")
	}

	/* FIXME: Should we also exclude other "special" files?
	if (stat.Mode().ModeType != 0) {
		return &argError{stat, "Only work with plain files."}
	}
	*/

	// FIXME: Where do we get user/group from - they don't appear to be in Go's Mode() object?
	return aw.writeHeaderBytes(stat.Name(), stat.Size(), uint64(stat.ModTime().Unix()), DEFAULT_USER, DEFAULT_GROUP, uint(mode&os.ModePerm))
}

func (aw *Writer) Add(name string, data []byte) error {
	err := aw.WriteHeaderDefault(name, int64(len(data)))
	if err != nil {
		return err
	}

	return aw.WriteBytes(data)
}
