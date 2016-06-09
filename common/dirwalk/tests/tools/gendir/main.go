// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

// Generate directory tree following a given structure.

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
)

type TestSettings struct {
	Name    string
	Comment string
	Seed    int64
	Tree    TreeSettings
}

var config = flag.String("config", "", "JSON config file for generating test file.")
var outdir = flag.String("outdir", "", "Where to write the output.")
var seed = flag.Int64("seed", 4, "Seed for random.")

func main() {
	flag.Parse()

	var settings TestSettings
	settings.Seed = *seed

	configdata, err := ioutil.ReadFile(*config)
	if err != nil {
		log.Fatal(err)
	}

	if err := json.Unmarshal(configdata, &settings); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s: %s\n", settings.Name, settings.Comment)
	fmt.Println("===============================================")
	fmt.Println()

	if len(*outdir) == 0 {
		log.Fatal("No output directory supplied.")
	}

	if _, err := os.Stat(*outdir); !os.IsNotExist(err) {
		log.Fatal("directory exists!")
	}

	if *seed != 4 && settings.Seed != *seed {
		log.Fatal("Seed supplied by test config.")
	}

	r := rand.New(rand.NewSource(settings.Seed))

	// Create the root directory
	if err := os.MkdirAll(*outdir, 0755); err != nil {
		log.Fatal(err)
	}

	GenerateTree(r, *outdir, &settings.Tree)
}
