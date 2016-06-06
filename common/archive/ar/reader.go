// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Read an ar file with BSD formatted file names.

package ar

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"
)

// Special UsageError that indicates trying to read after closing.
var (
	ErrReadAfterClose = UsageError{msg: "read after file closed"}
)

// ReadDataIOError indicates an error while reading data from the archive.
type ReadDataIOError struct {
	IOError
	wanted []byte
	got    []byte
}

func (e *ReadDataIOError) Error() string {
	return fmt.Sprintf("%s (wanted '%s', got '%s') -- *archive corrupted*", e.IOError.Error(), e.wanted, e.got)
}

// Fatal is always true for ReadDataIOError.
func (e *ReadDataIOError) Fatal() bool {
	return true
}

// FileInfo contains information about a file inside the ar archive.
type FileInfo interface {
	os.FileInfo
	UserID() int
	GroupID() int
}

type arFileInfoData struct {
	// os.FileInfo parts
	name    string
	size    int64
	mode    uint32
	modtime uint64
	// Extra parts
	uid int
	gid int
}

// os.FileInfo interface
func (fi *arFileInfoData) Name() string       { return fi.name }
func (fi *arFileInfoData) Size() int64        { return fi.size }
func (fi *arFileInfoData) Mode() os.FileMode  { return os.FileMode(fi.mode) }
func (fi *arFileInfoData) ModTime() time.Time { return time.Unix(int64(fi.modtime), 0) }
func (fi *arFileInfoData) IsDir() bool        { return fi.Mode().IsDir() }
func (fi *arFileInfoData) Sys() interface{}   { return fi }

// Extra
func (fi *arFileInfoData) UserID() int  { return fi.uid }
func (fi *arFileInfoData) GroupID() int { return fi.gid }

type readerStage uint

// Reader allows reading an existing ar archives.
type Reader struct {
	r              io.Reader
	bodyReader     io.LimitedReader
	bodyHasPadding bool
}

// NewReader allows reading an existing ar archives.
func NewReader(r io.Reader) (*Reader, Error) {
	reader := Reader{r: r}
	if err := reader.checkBytes("header", []byte("!<arch>\n")); err != nil {
		return nil, err
	}
	return &reader, nil
}

func (ar *Reader) checkBytes(section string, str []byte) Error {
	buffer := make([]byte, len(str))

	if _, err := io.ReadFull(ar.r, buffer); err != nil {
		return &IOError{section: section, err: err}
	}

	if !bytes.Equal(str, buffer) {
		return &ReadDataIOError{IOError: IOError{section: section}, wanted: str, got: buffer}
	}

	return nil
}

// Close the archive.
func (ar *Reader) Close() Error {
	if ar.r == nil {
		return &ErrReadAfterClose
	}
	ar.r = nil
	ar.bodyReader.R = nil
	return nil
}

func (ar *Reader) readHeaderInt64(section string, bytes int, formatstr string) (int64, Error) {
	data := make([]byte, bytes)
	if _, err := io.ReadFull(ar.r, data); err != nil {
		return -1, &IOError{section: section, err: err}
	}

	var output int64
	if _, err := fmt.Sscanf(string(data), formatstr, &output); err != nil {
		return -1, &ReadDataIOError{IOError: IOError{section: section, err: err}, wanted: []byte(formatstr), got: data}
	}

	if output <= 0 {
		return -1, &ReadDataIOError{IOError: IOError{section: section}, wanted: []byte(formatstr), got: data}
	}
	return output, nil
}

// Next file in the archive.
func (ar *Reader) Next() (FileInfo, Error) {
	if ar.r == nil {
		return nil, &ErrReadAfterClose
	}

	if ar.bodyReader.N > 0 {
		// Skip over any remaining part of previous file.
		if s, ok := ar.r.(io.Seeker); ok {
			if _, err := s.Seek(ar.bodyReader.N, io.SeekCurrent); err != nil {
				return nil, &ReadDataIOError{IOError: IOError{section: "body", err: err}, wanted: nil, got: nil}
			}
		} else {
			if _, err := io.Copy(ioutil.Discard, &ar.bodyReader); err != nil {
				return nil, &ReadDataIOError{IOError: IOError{section: "body", err: err}, wanted: nil, got: nil}
			}
		}

		// Padding to 16bit boundary
		if ar.bodyHasPadding {
			if err := ar.checkBytes("body padding", []byte{'\n'}); err != nil {
				return nil, err
			}
			ar.bodyHasPadding = false
		}
	}

	var fi arFileInfoData

	// File name length prefixed with '#1/' (BSD variant), 16 bytes
	namelen, err := ar.readHeaderInt64("file header file path length", 16, "#1/%13d")
	if err != nil {
		return nil, err
	}

	// Modtime, 12 bytes
	modtime, err := ar.readHeaderInt64("file modtime", 12, "%12d")
	if err != nil {
		return nil, err
	}
	fi.modtime = uint64(modtime)

	// Owner ID, 6 bytes
	ownerid, err := ar.readHeaderInt64("file header owner id", 6, "%6d")
	if err != nil {
		return nil, err
	}
	fi.uid = int(ownerid)

	// Group ID, 6 bytes
	groupid, err := ar.readHeaderInt64("file header group id", 6, "%6d")
	if err != nil {
		return nil, err
	}
	fi.gid = int(groupid)

	// File mode, 8 bytes
	filemod, err := ar.readHeaderInt64("file header file mode", 8, "%8o")
	if err != nil {
		return nil, err
	}
	fi.mode = uint32(filemod)

	// File size, 10 bytes
	size, err := ar.readHeaderInt64("file header data size", 10, "%10d")
	if err != nil {
		return nil, err
	}
	fi.size = size - namelen

	// File magic, 2 bytes
	if err = ar.checkBytes("file header file magic", []byte{'\x60', '\n'}); err != nil {
		return nil, err
	}

	ar.bodyReader = io.LimitedReader{R: ar.r, N: size}
	ar.bodyHasPadding = (size%2 != 0)

	// Filename - BSD variant
	filename := make([]byte, namelen)
	if _, err := io.ReadFull(&ar.bodyReader, filename); err != nil {
		return nil, &IOError{section: "body filename", err: err}
	}
	fi.name = string(filename)

	return &fi, nil
}

// Body for the current file in the archive.
func (ar *Reader) Body() io.Reader {
	return &ar.bodyReader
}
