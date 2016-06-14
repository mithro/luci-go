// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Read an ar file with BSD formatted file names.

package ar

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// Special UsageError that indicates trying to read after closing.
var (
	ErrReadAfterClose = UsageError{msg: "read after file closed"}
)

type ReadConstantIOError struct {
	IOError
	wanted []byte
	got []byte
}

func (e *ReadConstantIOError) Error() string {
	return fmt.Sprintf("archive/ar: io error (wanted %v, got %v) during %s -- *archive corrupted*", e.wanted, e.got, e.IOError.filesection)
}

type ReadDataIOError struct {
	IOError
	data interface{}
}

func (e *ReadDataIOError) Error() string {
	return fmt.Sprintf("archive/ar: io error (got %v) during %s -- *archive corrupted*", e.data, e.IOError.filesection)
}

// ReadTooLongFatalError indicates that the wrong amount of data *was* written into the archive.
// ReadTooLongFatalError is always fatal.
type ReadTooLongFatalError struct {
	needed int64
	got    int64
}

func (e *ReadTooLongFatalError) Error() string {
	return fmt.Sprintf("archive/ar: *reader broken* -- invalid data read (needed %d, got %d)", e.needed, e.got)
}
func (e *ReadTooLongFatalError) Fatal() bool {
	return true
}

type ReadTooLongError struct {
	needed int64
	got    int64
}

func (e *ReadTooLongError) Error() string {
	return fmt.Sprintf("archive/ar: invalid data read (needed %d, got %d)", e.needed, e.got)
}
func (e *ReadTooLongError) Fatal() bool {
	return false
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

const (
	readStageHeader readerStage = iota
	readStageBody
	readStageClosed
)

type Reader struct {
	stage               readerStage
	r                   io.Reader
	streamSizeRemaining int64
	bodyHasPadding        bool
}

func NewReader(r io.Reader) (*Reader, Error) {
	reader := Reader{r: r, stage: readStageHeader}
	if err := reader.checkBytes("header", []byte("!<arch>\n")); err != nil {
		return nil, err
	}
	return &reader, nil
}

func (ar *Reader) checkBytes(filesection string, str []byte) Error {
	buffer := make([]byte, len(str))

	if _, err := io.ReadFull(ar.r, buffer); err != nil {
		return &IOError{filesection: filesection, err:err}
	}

	if !bytes.Equal(str, buffer) {
		return &ReadConstantIOError{IOError: IOError{filesection: filesection}, wanted: str, got: buffer}
	}

	return nil
}

func (ar *Reader) Close() Error {
	switch ar.stage {
	case readStageHeader:
		// Good
	case readStageBody:
		return &UsageError{msg: "currently reading file body"}
	case readStageClosed:
		return &ErrReadAfterClose
	default:
		log.Fatalf("unknown reader mode: %d", ar.stage)
	}
	ar.stage = readStageClosed
	ar.r = nil
	return nil
}

func (ar *Reader) completeReadBytes(numbytes int64) Error {
	if numbytes > ar.streamSizeRemaining {
		return &ReadTooLongFatalError{needed: ar.streamSizeRemaining, got: numbytes}
	}

	ar.streamSizeRemaining -= numbytes
	if ar.streamSizeRemaining != 0 {
		return nil
	}

	// Padding to 16bit boundary
	if ar.bodyHasPadding {
		if err := ar.checkBytes("body padding", []byte{'\n'}); err != nil {
			return err
		}
		ar.bodyHasPadding = false
	}
	ar.stage = readStageHeader
	return nil
}

// Check we have finished reading bytes
func (ar *Reader) checkFinished() Error {
	return nil
}

func (ar *Reader) readPartial(filesection string, data []byte) Error {
	// Check you can read bytes from the ar at this moment.
	switch ar.stage {
	case readStageHeader:
		return &UsageError{msg: "need to read header first"}
	case readStageBody:
		// Good
	case readStageClosed:
		return &ErrReadAfterClose
	default:
		log.Fatalf("unknown reader mode: %d", ar.stage)
	}

	if datalen := int64(len(data)); datalen > ar.streamSizeRemaining {
		return &ReadTooLongError{needed: ar.streamSizeRemaining, got: datalen}
	}

	count, err := io.ReadFull(ar.r, data)
	if err != nil {
		return &IOError{filesection: filesection, err: err}
	}
	return ar.completeReadBytes(int64(count))
}

func (ar *Reader) readHeaderBytes(filesection string, bytes int, formatstr string) (int64, Error) {
	data := make([]byte, bytes)
	if _, err := io.ReadFull(ar.r, data); err != nil {
		return -1, &IOError{filesection: filesection, err: err}
	}

	var output int64
	if _, err := fmt.Sscanf(string(data), formatstr, &output); err != nil {
		return -1, &IOError{filesection: filesection, err: err}
	}

	if output <= 0 {
		return -1, &ReadDataIOError{IOError: IOError{filesection: filesection}, data: output}
	}
	return output, nil
}

func (ar *Reader) readHeader() (*arFileInfoData, Error) {
	switch ar.stage {
	case readStageHeader:
		// Good
	case readStageBody:
		return nil, &UsageError{msg:"currently writing a file"}
	case readStageClosed:
		return nil, &ErrReadAfterClose
	default:
		log.Fatalf("unknown reader mode: %d", ar.stage)
	}

	var fi arFileInfoData

	// File name length prefixed with '#1/' (BSD variant), 16 bytes
	namelen, err := ar.readHeaderBytes("file header filename length", 16, "#1/%13d")
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

	ar.stage = readStageBody
	ar.streamSizeRemaining = size
	ar.bodyHasPadding = (ar.streamSizeRemaining%2 != 0)

	// Filename - BSD variant
	filename := make([]byte, namelen)
	if err = ar.readPartial("body filename", filename); err != nil {
		return nil, err
	}
	fi.name = string(filename)

	return &fi, nil
}

func (ar *Reader) Read(b []byte) (int, Error) {
	if err := ar.readPartial("body contents", b); err != nil {
		return -1, err
	}
	return len(b), nil
}

func (ar *Reader) Next() (ArFileInfo, Error) {
	return ar.readHeader()
}
