// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package dirwalk

import "io"

type multiReadCloser struct {
	io.Reader
	closers []io.Closer
}

func (mrc multiReadCloser) Close() error {
	for _, c := range mrc.closers {
		if err := c.Close(); err != nil {
			return err
		}
	}
	return nil
}

type stringStack struct {
	elements []string
}

func (s *stringStack) push(v string) {
	s.elements = append(s.elements, v)
}

func (s *stringStack) pop() string {
	var v string
	v, s.elements = s.elements[s.last()], s.elements[:s.last()]
	return v
}

func (s *stringStack) peek() string {
	return s.elements[s.last()]
}

func (s *stringStack) size() int {
	return len(s.elements)
}

func (s *stringStack) last() int {
	return len(s.elements) - 1
}

func newStringStack() *stringStack {
	return &stringStack{elements: make([]string, 0)}
}
