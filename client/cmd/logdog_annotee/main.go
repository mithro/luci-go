// Copyright 2015 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/luci/luci-go/client/logdog/annotee/executor"
	"github.com/luci/luci-go/client/logdog/butlerlib/bootstrap"
	"github.com/luci/luci-go/client/logdog/butlerlib/streamclient"
	log "github.com/luci/luci-go/common/logging"
	"github.com/luci/luci-go/common/logging/gologger"
	"golang.org/x/net/context"
)

const (
	// configErrorReturnCode is returned when there is an error with Annotee's
	// command-line configuration.
	configErrorReturnCode = 2

	// runtimeErrorReturnCode is returned when the execution fails due to an
	// Annotee runtime error. This is intended to help differentiate Annotee
	// errors from passthrough bootstrapped subprocess errors.
	//
	// This will only be returned for runtime errors. If there is a flag error
	// or a configuration error, standard Annotee return codes (likely to overlap
	// with standard process return codes) will be used.
	//
	// This value has been chosen so as not to conflict with LogDog Butler runtime
	// error return code, allowing users to differentiate between Butler and
	// Annotee errors.
	runtimeErrorReturnCode = 251
)

type application struct {
	context.Context

	annotate           annotationMode
	jsonArgsPath       string
	butlerStreamServer string
	tee                bool
	printSummary       bool
	testingDir         string
}

func (a *application) addToFlagSet(fs *flag.FlagSet) {
	fs.StringVar(&a.jsonArgsPath, "json-args-path", "",
		"If specified, this is a JSON file containing the full command to run as an "+
			"array of strings.")
	fs.Var(&a.annotate, "annotate",
		"Annotation handling mode. Options are: "+annotationFlagEnum.Choices())
	fs.StringVar(&a.butlerStreamServer, "butler-stream-server", "",
		"The Butler stream server location. If empty, Annotee will check for Butler "+
			"bootstrapping and extract the stream server from that.")
	fs.BoolVar(&a.tee, "tee", true,
		"Tee the bootstrapped process' STDOUT/STDERR streams.")
	fs.BoolVar(&a.printSummary, "print-summary", true,
		"Print the annotation protobufs that were emitted at the end.")
	fs.StringVar(&a.testingDir, "testing-dir", "",
		"Rather than coupling to a Butler instance, output generated annotations "+
			"and streams to this directory.")
}

func (a *application) loadJSONArgs() ([]string, error) {
	fd, err := os.Open(a.jsonArgsPath)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	dec := json.NewDecoder(fd)
	args := []string(nil)
	if err := dec.Decode(&args); err != nil {
		return nil, err
	}
	return args, nil
}

func (a *application) getStreamClient() (streamclient.Client, error) {
	if a.testingDir != "" {
		return newFilesystemClient(a.testingDir)
	}

	if a.butlerStreamServer != "" {
		return streamclient.New(a.butlerStreamServer)
	}

	// Assume we're a bootstrapped application.
	bs, err := bootstrap.Get()
	if err != nil {
		log.Fields{
			log.ErrorKey: err,
		}.Errorf(a, "Could not get LogDog Butler bootstrap information.")
		return nil, err
	}
	if bs.Client == nil {
		return nil, errors.New("bootstrapped Butler did not export a stream server")
	}
	return bs.Client, nil
}

func mainImpl(args []string) int {
	ctx := gologger.Use(context.Background())

	logFlags := log.Config{
		Level: log.Warning,
	}
	a := &application{
		Context: ctx,
	}

	fs := &flag.FlagSet{}
	logFlags.AddFlags(fs)
	a.addToFlagSet(fs)
	if err := fs.Parse(args); err != nil {
		log.Fields{
			log.ErrorKey: err,
		}.Errorf(a, "Failed to parse flags.")
		return configErrorReturnCode
	}
	a.Context = logFlags.Set(a.Context)

	client, err := a.getStreamClient()
	if err != nil {
		log.Fields{
			log.ErrorKey: err,
		}.Errorf(a, "Failed to get stream client instance.")
		return configErrorReturnCode
	}

	// Determine bootstrapped process arguments.
	args = fs.Args()
	if a.jsonArgsPath != "" {
		if len(args) > 0 {
			log.Fields{
				"commandLineArgs": args,
				"jsonArgsPath":    a.jsonArgsPath,
			}.Errorf(a, "Cannot specify both JSON and command-line arguments.")
			return configErrorReturnCode
		}

		args, err = a.loadJSONArgs()
		if err != nil {
			log.Fields{
				log.ErrorKey:   err,
				"jsonArgsPath": a.jsonArgsPath,
			}.Errorf(a, "Failed to load JSON arguments.")
			return configErrorReturnCode
		}
	}
	if len(args) == 0 {
		log.Errorf(a, "No command-line arguments were supplied.")
		return configErrorReturnCode
	}

	e := executor.Executor{
		Annotate: executor.AnnotationMode(a.annotate),
		Stdin:    os.Stdin,
		Command:  args,
		Client:   client,
	}
	if a.tee {
		e.TeeStdout = os.Stdout
		e.TeeStderr = os.Stderr
	}
	if err := e.Run(a); err != nil {
		log.Fields{
			log.ErrorKey: err,
		}.Errorf(a, "Failed during execution.")
		return runtimeErrorReturnCode
	}

	// Display a summary!
	if a.printSummary {
		for _, s := range e.Steps() {
			fmt.Printf("=== Annotee: %q ===\n", s.StepComponent.Name)
			fmt.Println(proto.MarshalTextString(s))
		}
	}

	return e.ReturnCode()
}

func main() {
	os.Exit(mainImpl(os.Args[1:]))
}
