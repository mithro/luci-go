// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ar

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
	"time"
)

var (
	TestFile1 = ("" +
		// ar file header
		"!<arch>\n" +
		// filename len	- 16 bytes
		"#1/9            " +
		// modtime		- 12 bytes
		"1447140471  " +
		// owner id		- 6 bytes
		"1000  " +
		// group id		- 6 bytes
		"1000  " +
		// file mode	- 8 bytes
		"100640  " +
		// Data size	- 10 bytes
		"15        " +
		// File magic	- 2 bytes
		"\x60\n" +
		// File name	- 9 bytes
		"filename1" +
		// File data	- 6 bytes
		"abc123" +
		// Padding		- 1 byte
		"\n" +
		"")
)

func TestWriterCreatesTestFile1(t *testing.T) {
	b := bytes.NewBufferString("")
	data := []byte("abc123")

	ar := NewWriter(b)
	err := ar.Add("filename1", data)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	err = ar.Close()
	if err != nil {
		t.Fatalf("Close: %v", err)
	}

	if g, e := b.Bytes(), []byte(TestFile1); !bytes.Equal(g, e) {
		t.Errorf("\ndata = \n%v\n%v", g, e)
	}
}

func TestReaderOnTestFile1(t *testing.T) {
	r := strings.NewReader(TestFile1)

	ar, err := NewReader(r)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}

	h, err := ar.Next()
	if err != nil {
		t.Fatalf("Header: %v", err)
	}
	if g, e := h.Name(), "filename1"; g != e {
		t.Errorf("Name() = %q; want %q", g, e)
	}
	if g, e := h.Size(), int64(6); g != e {
		t.Errorf("Size() = %d; want %d", g, e)
	}
	if g, e := h.ModTime(), time.Unix(1447140471, 0); !g.Equal(e) {
		t.Errorf("ModTime() = %v; want %v", g, e)
	}

	data := make([]byte, 6)
	n, err := ar.Read(data)
	if err != nil {
		t.Fatalf("Data: %v", err)
	}
	if g, e := n, 6; g != e {
		t.Errorf("n = %v; want %v", g, e)
	}
	if g, e := data, []byte("abc123"); !bytes.Equal(g, e) {
		t.Errorf("data = %v; want %v", g, e)
	}

	err = ar.Close()
	if err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestWithSystemArCommandList(t *testing.T) {
	_, err := exec.LookPath("ar")
	if err != nil {
		t.Skipf("ar command not found: %v", err)
	}

	// Write out to an archive file
	tmpfile, err := ioutil.TempFile("", "go-ar-test.")
	defer os.Remove(tmpfile.Name()) // clean up
	ar := NewWriter(tmpfile)
	ar.Add("file1.txt", []byte("file1 contents"))
	ar.Add("file2.txt", []byte("file2 contents"))
	ar.Add("dir1/file3.txt", []byte("file3 contents"))
	ar.Close()

	// Use the ar command to list the file
	cmd_list := exec.Command("ar", "t", tmpfile.Name())
	var cmd_list_out_buf bytes.Buffer
	cmd_list.Stdout = &cmd_list_out_buf
	err = cmd_list.Run()
	if err != nil {
		t.Fatalf("ar command failed: %v\n%s", err, cmd_list_out_buf.String())
	}

	cmd_list_actual_out := cmd_list_out_buf.String()
	cmd_list_expect_out := `file1.txt
file2.txt
dir1/file3.txt
`
	if strings.Compare(cmd_list_actual_out, cmd_list_expect_out) != 0 {
		t.Fatalf("ar command output: '%s'", cmd_list_actual_out)
	}
}

func TestWithSystemArCommandExtract(t *testing.T) {
	arpath, err := exec.LookPath("ar")
	if err != nil {
		t.Skipf("ar command not found: %v", err)
	}

	// Write out to an archive file
	tmpfile, err := ioutil.TempFile("", "go-ar-test.")
	defer os.Remove(tmpfile.Name()) // clean up
	ar := NewWriter(tmpfile)
	ar.Add("file1.txt", []byte("file1 contents"))
	ar.Add("file2.txt", []byte("file2 contents"))
	ar.Close()

	// Extract the ar
	tmpdir, err := ioutil.TempDir("", "go-ar-test.")
	defer os.RemoveAll(tmpdir)
	cmd_extract := exec.Cmd{
		Path: arpath,
		Args: []string{"ar", "x", tmpfile.Name()},
		Dir:  tmpdir,
	}
	err = cmd_extract.Run()
	var cmd_extract_out_buf bytes.Buffer
	cmd_extract.Stdout = &cmd_extract_out_buf
	if err != nil {
		t.Fatalf("ar command failed: %v\n%s", err, cmd_extract_out_buf.String())
	}

	// Compare the directory output
	dir_contents, err := ioutil.ReadDir(tmpdir)
	if err != nil {
		t.Fatalf("Unable to read the output directory: %v", err)
	}
	for _, fi := range dir_contents {
		if fi.Name() != "file1.txt" && fi.Name() != "file2.txt" {
			t.Errorf("Found unexpected file '%s'", fi.Name())
		}
	}

	file1_contents, err := ioutil.ReadFile(path.Join(tmpdir, "file1.txt"))
	file1_expected := []byte("file1 contents")
	if err != nil {
		t.Errorf("%v", err)
	} else {
		if bytes.Compare(file1_contents, file1_expected) != 0 {
			t.Errorf("file1.txt content incorrect. Got:\n%v\n%v\n", file1_contents, file1_expected)
		}
	}

	file2_contents, err := ioutil.ReadFile(path.Join(tmpdir, "file2.txt"))
	file2_expected := []byte("file2 contents")
	if err != nil {
		t.Errorf("%v", err)
	} else {
		if bytes.Compare(file2_contents, file2_expected) != 0 {
			t.Errorf("file2.txt content incorrect. Got:\n%v\n%v\n", file2_contents, file2_expected)
		}
	}
}
