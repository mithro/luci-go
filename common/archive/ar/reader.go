// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Read an ar file with BSD formatted file names.

package ar

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

var (
    ErrHeader = errors.New("archive/tar: invalid tar header")
)

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

const (
	READ_HEADER readerStage = iota
	READ_BODY
	READ_CLOSED
)

type Reader struct {
	stage         readerStage
	r             io.Reader
	streamSizeRemaining int64
	needspadding  bool
}

func NewReader(r io.Reader) (*Reader, error) {
	reader := Reader{r: r, stage: READ_HEADER}
	if err := reader.checkBytes("header", []byte("!<arch>\n")); err != nil {
		return nil, err
	}
	return &reader, nil
}

func (ar *Reader) checkBytes(name string, str []byte) error {
	buffer := make([]byte, len(str))

	if _, err := io.ReadFull(ar.r, buffer); err != nil {
		return fmt.Errorf("%s: error in reading bytes (%v)", name, err)
	}

	if !bytes.Equal(str, buffer) {
		return fmt.Errorf("%s: error in bytes (wanted: %v got: %v)", name, buffer, str)
	}

	return nil
}

func (ar *Reader) Close() error {
	switch ar.stage {
	case READ_HEADER:
		// Good
	case READ_BODY:
		return errors.New("usage error, reading a file.")
	case READ_CLOSED:
		return errors.New("usage error, archive already closed.")
	default:
		log.Fatalf("unknown reader mode: %d", ar.stage)
	}
	ar.stage = READ_CLOSED
	ar.r = nil
	return nil
}

func (ar *Reader) completeReadBytes(numbytes int64) error {
	if numbytes > ar.streamSizeRemaining {
		return fmt.Errorf("to much data read, needed %d, got %d", ar.streamSizeRemaining, numbytes)
	}

	ar.streamSizeRemaining -= numbytes
	if ar.streamSizeRemaining != 0 {
		return nil
	}

	// Padding to 16bit boundary
	if ar.needspadding {
		if err := ar.checkBytes("padding", []byte{"\n"}); err != nil {
			return err
		}
		ar.needspadding = false
	}
	ar.stage = READ_HEADER
	return nil
}

// Check we have finished reading bytes
func (ar *Reader) checkFinished() error {
	if ar.streamSizeRemaining != 0 {
		return fmt.Errorf("didn't read enough %d bytes still needed", ar.streamSizeRemaining)
	}
	return nil
}

func (ar *Reader) readPartial(name string, data []byte) error {
	// Check you can read bytes from the ar at this moment.
	switch ar.stage {
	case READ_HEADER:
		return errors.New("usage error, need to read header first")
	case READ_BODY:
		// Good
	case READ_CLOSED:
		return errors.New("usage error, archive closed")
	default:
		log.Fatalf("unknown reader mode: %d", ar.stage)
	}

	if datalen := int64(len(data)); datalen > ar.streamSizeRemaining {
		return fmt.Errorf("to much data, wanted %d, but had %d", ar.streamSizeRemaining, datalen)
	}

	count, err := io.ReadFull(ar.r, data)
	if err != nil {
		return err
	}
	ar.completeReadBytes(int64(count))
	return nil
}

func (ar *Reader) readHeaderBytes(name string, bytes int, formatstr string) (int64, error) {
	data := make([]byte, bytes)
	if _, err := io.ReadFull(ar.r, data); err != nil {
		return -1, err
	}

	var output int64
	if _, err = fmt.Sscanf(string(data), formatstr, &output); err != nil {
		return -1, err
	}

	if output <= 0 {
		return -1, fmt.Errorf("%s: bad value %d", name, output)
	}
	return output, nil
}

func (ar *Reader) readHeader() (*arFileInfoData, error) {
	switch ar.stage {
	case READ_HEADER:
		// Good
	case READ_BODY:
		return nil, errors.New("usage error, already writing a file")
	case READ_CLOSED:
		return nil, errors.New("usage error, archive closed")
	default:
		log.Fatalf("unknown reader mode: %d", ar.stage)
	}

	var fi arFileInfoData

	// File name length prefixed with '#1/' (BSD variant), 16 bytes
	namelen, err := ar.readHeaderBytes("filename length", 16, "#1/%13d")
	if err != nil {
		return nil, err
	}

	// Modtime, 12 bytes
	modtime, err := ar.readHeaderBytes("modtime", 12, "%12d")
	if err != nil {
		return nil, err
	}
	fi.modtime = uint64(modtime)

	// Owner ID, 6 bytes
	ownerid, err := ar.readHeaderBytes("ownerid", 6, "%6d")
	if err != nil {
		return nil, err
	}
	fi.uid = int(ownerid)

	// Group ID, 6 bytes
	groupid, err := ar.readHeaderBytes("groupid", 6, "%6d")
	if err != nil {
		return nil, err
	}
	fi.gid = int(groupid)

	// File mode, 8 bytes
	filemod, err := ar.readHeaderBytes("filemod", 8, "%8o")
	if err != nil {
		return nil, err
	}
	fi.mode = uint32(filemod)

	// File size, 10 bytes
	size, err := ar.readHeaderBytes("datasize", 10, "%10d")
	if err != nil {
		return nil, err
	}
	fi.size = size - namelen

	// File magic, 2 bytes
	if err = ar.checkBytes("filemagic", []byte{"\x60", "\n"}); err != nil {
		return nil, err
	}

	ar.stage = READ_BODY
	ar.streamSizeRemaining = size
	ar.needspadding = (ar.streamSizeRemaining%2 != 0)

	// Filename - BSD variant
	filename := make([]byte, namelen)
	if err = ar.readPartial("filename", filename); err != nil {
		return nil, err
	}
	fi.name = string(filename)

	return &fi, nil
}

func (ar *Reader) Read(b []byte) (int, error) {
	if err = ar.readPartial("data", b); err != nil {
		return -1, err
	}
	if err = ar.checkFinished(); err != nil {
		return -1, err
	}
	return len(b), nil
}

func (ar *Reader) Next() (ArFileInfo, error) {
	return ar.readHeader()
}
