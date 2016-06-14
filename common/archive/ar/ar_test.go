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

	"github.com/maruel/ut"
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

	TestFile2 = ("" +
		// ar file header
		"!<arch>\n" +

		// File 1
		// ----------------------
		// filename len	- 16 bytes
		"#1/5            " +
		// modtime		- 12 bytes
		"1447140471  " +
		// owner id		- 6 bytes
		"1000  " +
		// group id		- 6 bytes
		"1000  " +
		// file mode	- 8 bytes
		"100640  " +
		// Data size	- 10 bytes
		"13        " +
		// File magic	- 2 bytes
		"\x60\n" +
		// File name	- 9 bytes
		"file1" +
		// File data	- 6 bytes
		"contents" +
		// Padding		- 1 byte
		"\n" +

		// File 2
		// ----------------------
		// filename len	- 16 bytes
		"#1/7            " +
		// modtime		- 12 bytes
		"1447140471  " +
		// owner id		- 6 bytes
		"1000  " +
		// group id		- 6 bytes
		"1000  " +
		// file mode	- 8 bytes
		"100640  " +
		// Data size	- 10 bytes
		"10        " +
		// File magic	- 2 bytes
		"\x60\n" +
		// File name	- 9 bytes
		"fileabc" +
		// File data	- 6 bytes
		"123" +
		// No padding	- 0 byte

		// File 3
		// ----------------------
		// filename len	- 16 bytes
		"#1/10           " +
		// modtime		- 12 bytes
		"1447140471  " +
		// owner id		- 6 bytes
		"1000  " +
		// group id		- 6 bytes
		"1000  " +
		// file mode	- 8 bytes
		"100640  " +
		// Data size	- 10 bytes
		"16        " +
		// File magic	- 2 bytes
		"\x60\n" +
		// File name	- 9 bytes
		"dir1/file1" +
		// File data	- 6 bytes
		"123abc" +
		// No padding	- 0 byte
		"")
)

func TestWriterCreatesTestFile1(t *testing.T) {
	b := &bytes.Buffer{}
	data := []byte("abc123")

	ar, err := NewWriter(b)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	if err := ar.Add("filename1", data); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := ar.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	ut.AssertEqual(t, []byte(TestFile1), b.Bytes())
}

func TestReaderOnTestFile1(t *testing.T) {
	r := strings.NewReader(TestFile1)

	ar, err := NewReader(r)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}

	h, herr := ar.Next()
	if herr != nil {
		t.Fatalf("Header: %v", herr)
	}

	ut.AssertEqual(t, time.Unix(1447140471, 0), h.ModTime())
	ut.AssertEqual(t, 1000, h.UserId())
	ut.AssertEqual(t, 1000, h.GroupId())
	ut.AssertEqual(t, "filename1", h.Name())
	ut.AssertEqual(t, int64(6), h.Size())

	data1 := make([]byte, 3)
	data2 := make([]byte, 4)
	n1, berr := ar.Body().Read(data1)
	if berr != nil {
		t.Fatalf("Data: %v", berr)
	}
	ut.AssertEqual(t, 3, n1)
	n2, berr := ar.Body().Read(data2)
	if berr != nil {
		t.Fatalf("Data: %v", berr)
	}
	ut.AssertEqual(t, 3, n2)

	ut.AssertEqual(t, []byte("abc"), data1)
	ut.AssertEqual(t, []byte{'1', '2', '3', 0}, data2)

	if err := ar.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestReaderOnTestFile2(t *testing.T) {
	r := strings.NewReader(TestFile2)

	ar, err := NewReader(r)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}

	h1, herr := ar.Next()
	if herr != nil {
		t.Fatalf("Header: %v", herr)
	}
	ut.AssertEqual(t, time.Unix(1447140471, 0), h1.ModTime())
	ut.AssertEqual(t, 1000, h1.UserId())
	ut.AssertEqual(t, 1000, h1.GroupId())
	ut.AssertEqual(t, "file1", h1.Name())
	ut.AssertEqual(t, int64(8), h1.Size())

	// Skipping the body of file 1

	h2, herr := ar.Next()
	if herr != nil {
		t.Fatalf("Header: %v", herr)
	}
	ut.AssertEqual(t, time.Unix(1447140471, 0), h2.ModTime())
	ut.AssertEqual(t, 1000, h2.UserId())
	ut.AssertEqual(t, 1000, h2.GroupId())
	ut.AssertEqual(t, "fileabc", h2.Name())
	ut.AssertEqual(t, int64(3), h2.Size())

	// Read only some of the body
	data := make([]byte, 2)
	n, berr := ar.Body().Read(data)
	if berr != nil {
		t.Fatalf("Data: %v", berr)
	}
	ut.AssertEqual(t, 2, n)
	ut.AssertEqual(t, []byte("12"), data)

	h3, herr := ar.Next()
	if herr != nil {
		t.Fatalf("Header: %v", herr)
	}
	ut.AssertEqual(t, time.Unix(1447140471, 0), h3.ModTime())
	ut.AssertEqual(t, 1000, h3.UserId())
	ut.AssertEqual(t, 1000, h3.GroupId())
	ut.AssertEqual(t, "dir1/file1", h3.Name())
	ut.AssertEqual(t, int64(6), h3.Size())

	// Read the full file
	data = make([]byte, 6)
	n, berr = ar.Body().Read(data)
	if berr != nil {
		t.Fatalf("Data: %v", berr)
	}
	ut.AssertEqual(t, 6, n)
	ut.AssertEqual(t, []byte("123abc"), data)

	if err := ar.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestWithSystemArCommandList(t *testing.T) {
	if _, err := exec.LookPath("ar"); err != nil {
		t.Skipf("ar command not found: %v", err)
	}

	// Write out to an archive file
	tmpfile, err := ioutil.TempFile("", "go-ar-test.")
	if err != nil {
		t.Fatalf("unable to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name()) // clean up
	ar, err := NewWriter(tmpfile)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	ar.Add("file1.txt", []byte("file1 contents"))
	ar.Add("file2.txt", []byte("file2 contents"))
	ar.Add("dir1/file3.txt", []byte("file3 contents"))
	ar.Close()

	// Use the ar command to list the file
	cmd_list := exec.Command("ar", "t", tmpfile.Name())
	var cmd_list_out_buf bytes.Buffer
	cmd_list.Stdout = &cmd_list_out_buf
	if err := cmd_list.Run(); err != nil {
		t.Fatalf("ar command failed: %v\n%s", err, cmd_list_out_buf.String())
	}

	cmd_list_actual_out := cmd_list_out_buf.String()
	cmd_list_expect_out := `file1.txt
file2.txt
dir1/file3.txt
`
	ut.AssertEqual(t, cmd_list_expect_out, cmd_list_actual_out)
}

func TestWithSystemArCommandExtract(t *testing.T) {
	arpath, err := exec.LookPath("ar")
	if err != nil {
		t.Skipf("ar command not found: %v", err)
	}

	// Write out to an archive file
	tmpfile, err := ioutil.TempFile("", "go-ar-test.")
	if err != nil {
		t.Fatalf("unable to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name()) // clean up
	ar, err := NewWriter(tmpfile)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
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
	var cmd_extract_out_buf bytes.Buffer
	cmd_extract.Stdout = &cmd_extract_out_buf
	if err := cmd_extract.Run(); err != nil {
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
