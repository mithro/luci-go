// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

/**
 * Read an ar archive file.
 */
package ar

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

type arFileInfo struct {
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
func (fi arFileInfo) Name() string       { return fi.name }
func (fi arFileInfo) Size() int64        { return fi.size }
func (fi arFileInfo) Mode() os.FileMode  { return os.FileMode(fi.mode) }
func (fi arFileInfo) ModTime() time.Time { return time.Unix(int64(fi.modtime), 0) }
func (fi arFileInfo) IsDir() bool        { return fi.Mode().IsDir() }
func (fi arFileInfo) Sys() interface{}   { return fi }

// Extra
func (fi arFileInfo) UserId() int  { return fi.uid }
func (fi arFileInfo) GroupId() int { return fi.gid }

var (
	ErrHeader = errors.New("archive/ar: invalid ar header")
)

type ReaderStage uint

const (
	READ_HEADER ReaderStage = iota
	READ_BODY               = iota
	READ_CLOSED             = iota
)

type Reader struct {
	stage         ReaderStage
	r             io.Reader
	bytesrequired int64
	needspadding  bool
}

func NewReader(r io.Reader) (*Reader, error) {
	reader := Reader{r: r, bytesrequired: 0, needspadding: false}
	err := reader.checkBytes("header", []byte("!<arch>\n"))
	if err != nil {
		return nil, err
	} else {
		return &reader, nil
	}
}

func (ar *Reader) checkBytes(name string, str []byte) error {
	buffer := make([]byte, len(str))

	count, err := io.ReadFull(ar.r, buffer)
	if err != nil {
		return err
	}

	if count != len(buffer) {
		return errors.New(fmt.Sprintf("%s: Not enough data read (only %d, needed %d)", name, count, len(buffer)))
	}

	if bytes.Equal(str, buffer) {
		return nil
	} else {
		return errors.New(fmt.Sprintf("%s: error in bytes (wanted: %v got: %v)", name, buffer, str))
	}
}

func (ar *Reader) Close() error {
	switch ar.stage {
	case READ_HEADER:
		// Good
	case READ_BODY:
		return errors.New("Usage error, reading a file.")
	case READ_CLOSED:
		return errors.New("Usage error, archive already closed.")
	default:
		panic(fmt.Sprintf("Unknown reader mode: %d", ar.stage))
	}
	//ar.r.Close()
	ar.stage = READ_CLOSED
	return nil
}

func (ar *Reader) readBytes(numbytes int64) error {
	if numbytes > ar.bytesrequired {
		return errors.New(fmt.Sprintf("To much data read! Needed %d, got %d", ar.bytesrequired, numbytes))
	}

	ar.bytesrequired -= numbytes
	if ar.bytesrequired != 0 {
		return nil
	}

	// Padding to 16bit boundary
	if ar.needspadding {
		err := ar.checkBytes("padding", []byte("\n"))
		if err != nil {
			return err
		}
		ar.needspadding = false
	}
	ar.stage = READ_HEADER
	return nil
}

// Check you can write bytes to the ar at this moment.
func (ar *Reader) checkRead() error {
	switch ar.stage {
	case READ_HEADER:
		return errors.New("Usage error, need to read header first.")
		// Good
	case READ_BODY:
		return nil
	case READ_CLOSED:
		return errors.New("Usage error, archive closed.")
	default:
		panic(fmt.Sprintf("Unknown reader mode: %d", ar.stage))
	}
}

// Check we have finished writing bytes
func (ar *Reader) checkFinished() error {
	if ar.bytesrequired != 0 {
		return errors.New(fmt.Sprintf("Didn't read enough bytes %d still needed!", ar.bytesrequired))
	}
	return nil
}

func (ar *Reader) readPartial(name string, data []byte) error {
	err := ar.checkRead()
	if err != nil {
		return err
	}

	datalen := int64(len(data))
	if datalen > ar.bytesrequired {
		return errors.New(fmt.Sprintf("To much data! Wanted %d, but had %d", ar.bytesrequired, datalen))
	}

	count, err := io.ReadFull(ar.r, data)
	ar.readBytes(int64(count))
	return nil
}

func (ar *Reader) readHeaderBytes(name string, bytes int, formatstr string) (int64, error) {
	data := make([]byte, bytes)
	_, err := io.ReadFull(ar.r, data)
	if err != nil {
		return -1, err
	}

	var output int64 = 0
	_, err = fmt.Sscanf(string(data), formatstr, &output)
	if err != nil {
		return -1, err
	}

	if output <= 0 {
		return -1, errors.New(fmt.Sprintf("%s: bad value %d", name, output))
	}
	return output, nil
}

func (ar *Reader) readHeader() (*arFileInfo, error) {
	switch ar.stage {
	case READ_HEADER:
		// Good
	case READ_BODY:
		return nil, errors.New("Usage error, already writing a file.")
	case READ_CLOSED:
		return nil, errors.New("Usage error, archive closed.")
	default:
		panic(fmt.Sprintf("Unknown writer mode: %d", ar.stage))
	}

	var fi arFileInfo

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
	filemod, err := ar.readHeaderBytes("groupid", 8, "%8o")
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

	ar.stage = READ_BODY
	ar.bytesrequired = size
	ar.needspadding = (ar.bytesrequired%2 != 0)

	// File magic, 2 bytes
	err = ar.checkBytes("filemagic", []byte("\x60\n"))
	if err != nil {
		return nil, err
	}

	// Filename - BSD variant
	filename := make([]byte, namelen)
	err = ar.readPartial("filename", filename)
	if err != nil {
		return nil, err
	}
	fi.name = string(filename)

	return &fi, nil
}

func (ar *Reader) Read(b []byte) (n int, err error) {
	err = ar.readPartial("data", b)
	if err != nil {
		return -1, err
	}
	err = ar.checkFinished()
	if err != nil {
		return -1, err
	}
	return len(b), nil
}
func (ar *Reader) Next() (*arFileInfo, error) {
	return ar.readHeader()
}
