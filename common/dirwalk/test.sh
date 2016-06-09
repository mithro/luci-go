#!/bin/bash
# Copyright 2016 The LUCI Authors. All rights reserved.
# Use of this source code is governed under the Apache License, Version 2.0
# that can be found in the LICENSE file.

TESTS=common/dirwalk/tests/*.json

go install github.com/luci/luci-go/common/dirwalk/tests/tools/gendir
go install github.com/luci/luci-go/common/dirwalk/tests/tools/walkdir

echo "Generating the test directories"
TMPDIR=/usr/local/google/tmp/luci-tests
mkdir -p $TMPDIR
for TESTFILE in $TESTS; do
	TESTNAME="$(basename $TESTFILE .json)"
	TESTDIR="$TMPDIR/$TESTNAME"
	if ! [ -d $TESTDIR ]; then
		echo "Generating test directory for $TESTNAME"
		gendir -config $TESTFILE -outdir $TESTDIR
		du -h $TESTDIR
	fi
done

for METHOD in simple nostat parallel; do
	echo "Running $METHOD"
	for TESTFILE in $TESTS; do
		TESTNAME="$(basename $TESTFILE .json)"
		TESTDIR="$TMPDIR/$TESTNAME"
		OUTPUT=output.$METHOD.$TESTNAME
		echo "Running $METHOD.$TESTNAME"
		rm $OUTPUT
		$(which time) --verbose --output=$OUTPUT --append walkdir --dir $TESTDIR --method $METHOD $@ 2> $OUTPUT
		tail -n 20 $OUTPUT
		echo
	done
	echo
	echo
	echo
done
