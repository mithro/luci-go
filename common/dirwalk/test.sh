#!/bin/bash
# Copyright 2016 The LUCI Authors. All rights reserved.
# Use of this source code is governed under the Apache License, Version 2.0
# that can be found in the LICENSE file.

set -e
set -x

TESTS=common/dirwalk/tests/*.json

go install github.com/luci/luci-go/common/dirwalk/tests/tools/gendir
go install github.com/luci/luci-go/common/dirwalk/tests/tools/walkdir

echo "Generating the test directories"
echo "----------------------------------"
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
echo "All test directories done."
echo
echo

echo "Verifying the walks"
echo "----------------------------------"
for METHOD in basic nostat parallel; do
	for TESTFILE in $TESTS; do
		TESTNAME="$(basename $TESTFILE .json)"
		TESTDIR="$TMPDIR/$TESTNAME"
		echo "Verifying $METHOD on $TESTNAME"
		walkdir --dir $TESTDIR --method $METHOD --do verify || exit 1
		echo
	done
	echo
	echo
	echo
done
echo "All test directories done."
echo
echo

echo "Running the performance tests"
echo "----------------------------------"
for METHOD in basic nostat parallel; do
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
echo "All perf tests done."
echo
echo
