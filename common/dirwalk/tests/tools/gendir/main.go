// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
)

type testSettings struct {
	Name    string
	Comment string
	Seed    int64
	Tree    treeSettings
}

func mainImpl() error {
	config := flag.String("config", "", "JSON config file for generating test file.")
	outdir := flag.String("outdir", "", "Where to write the output.")
	seed := flag.Int64("seed", 4, "Seed for random.")
	verbose := flag.Bool("v", false, "verbose mode")

	flag.Parse()

	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}

	var settings testSettings
	settings.Seed = *seed

	configdata, err := ioutil.ReadFile(*config)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(configdata, &settings); err != nil {
		return err
	}

	fmt.Printf("%s: %s\n", settings.Name, settings.Comment)
	fmt.Println("===============================================")
	fmt.Println()

	if len(*outdir) == 0 {
		return errors.New("no output directory supplied")
	}

	if _, err := os.Stat(*outdir); !os.IsNotExist(err) {
		return errors.New("directory exists")
	}

	if *seed != 4 && settings.Seed != *seed {
		return errors.New("seed supplied by test config")
	}

	r := rand.New(rand.NewSource(settings.Seed))

	// Create the root directory
	if err := os.MkdirAll(*outdir, 0700); err != nil {
		return err
	}

	generateTree(r, *outdir, &settings.Tree)
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "gendir: %s.\n", err)
		os.Exit(1)
	}
}
