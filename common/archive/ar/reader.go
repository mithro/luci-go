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

type ReadDataIOError struct {
	IOError
	wanted []byte
	got    []byte
}

func (e *ReadDataIOError) Error() string {
	return fmt.Sprintf("%s (wanted '%s', got '%s') during %s -- *archive corrupted*", e.IOError.Error(), e.wanted, e.got)
}

type ArFileInfo interface {
	os.FileInfo
	UserId() int
	GroupId() int
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
func (fi *arFileInfoData) UserId() int  { return fi.uid }
func (fi *arFileInfoData) GroupId() int { return fi.gid }

type readerStage uint

type Reader struct {
	r              io.Reader
	bodyReader     io.LimitedReader
	bodyHasPadding bool
}

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

func (ar *Reader) Close() Error {
	if ar.r == nil {
		return &ErrReadAfterClose
	}
	ar.r = nil
	ar.bodyReader.R = nil
	return nil
}

func (ar *Reader) readHeaderBytes(section string, bytes int, formatstr string) (int64, Error) {
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

func (ar *Reader) Next() (ArFileInfo, Error) {
	if ar.r == nil {
		return nil, &ErrReadAfterClose
	}

	if ar.bodyReader.N > 0 {
		// Read any remains of the previous file
		io.Copy(ioutil.Discard, &ar.bodyReader)

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
	namelen, err := ar.readHeaderBytes("file header file path length", 16, "#1/%13d")
	if err != nil {
		return nil, err
	}

	// Modtime, 12 bytes
	modtime, err := ar.readHeaderBytes("file modtime", 12, "%12d")
	if err != nil {
		return nil, err
	}
	fi.modtime = uint64(modtime)

	// Owner ID, 6 bytes
	ownerid, err := ar.readHeaderBytes("file header owner id", 6, "%6d")
	if err != nil {
		return nil, err
	}
	fi.uid = int(ownerid)

	// Group ID, 6 bytes
	groupid, err := ar.readHeaderBytes("file header group id", 6, "%6d")
	if err != nil {
		return nil, err
	}
	fi.gid = int(groupid)

	// File mode, 8 bytes
	filemod, err := ar.readHeaderBytes("file header file mode", 8, "%8o")
	if err != nil {
		return nil, err
	}
	fi.mode = uint32(filemod)

	// File size, 10 bytes
	size, err := ar.readHeaderBytes("file header data size", 10, "%10d")
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

func (ar *Reader) Body() io.Reader {
	return &ar.bodyReader
}
