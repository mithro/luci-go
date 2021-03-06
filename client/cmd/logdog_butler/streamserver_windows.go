// Copyright 2015 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"

	"github.com/luci/luci-go/client/internal/logdog/butler/streamserver"
	"golang.org/x/net/context"
)

const (
	// An example stream server URI.
	//
	// This expands to: //localhost/pipe/<name>
	exampleStreamServerURI = streamServerURI("net.pipe:logdog-butler")
)

type streamServerURI string

func (u streamServerURI) Parse() (string, error) {
	typ, value := parseStreamServer(string(u))
	if typ != "net.pipe" {
		return "", errors.New("Unsupported URI scheme.")
	}

	if value == "" {
		return "", errors.New("cannot have empty pipe name")
	}
	return value, nil
}

// Validates that the URI is correct for Windows.
func (u streamServerURI) Validate() error {
	_, err := u.Parse()
	return err
}

// Create a Windows stream server.
func createStreamServer(ctx context.Context, uri streamServerURI) streamserver.StreamServer {
	name, err := uri.Parse()
	if err != nil {
		panic("Failed to parse stream server URI.")
	}
	return streamserver.NewNamedPipeServer(ctx, fmt.Sprintf(`\\.\pipe\%s`, name))
}
