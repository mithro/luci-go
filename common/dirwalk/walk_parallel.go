// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package dirwalk

import (
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/eapache/channels"
)

type fileQueue struct {
	queued   uint64
	finished uint64
	items    channels.Channel
	waiton   chan bool
}

func (q *fileQueue) add(s string) {
	atomic.AddUint64(&q.queued, 1)
	q.items.In() <- s
}

func (q *fileQueue) done() {
	atomic.AddUint64(&q.finished, 1)

	if q.queued == q.finished {
		q.items.Close()
		q.waiton <- true
	}
}

func (q *fileQueue) wait() {
	<-q.waiton
}

func examinePath(queue *fileQueue, callback WalkFunc) {
	for ipath := range queue.items.Out() {
		path := ipath.(string)

		fi, err := os.Stat(path)
		if err != nil {
			callback(path, -1, nil, err)
			return
		}

		if fi.IsDir() {
			d, err := os.Open(path)
			if err != nil {
				callback(path, -1, nil, err)
			}

			dircontents, err := d.Readdirnames(-1)
			if err != nil {
				callback(path, -1, nil, err)
			}
			for _, name := range dircontents {
				fname := filepath.Join(path, name)
				queue.add(fname)
			}
		} else {
			f, err := os.Open(path)
			if err != nil {
				callback(path, -1, nil, err)
			} else {
				callback(path, fi.Size(), f, nil)
			}
		}
		queue.done()
	}
}

// WalkParallel is a directory walking function which uses multiple threads to
// walk a directory tree.
//
// On the majority of systems testing shows that this function is either
// slower than (or at best comparable) the non-parallel version while consuming
// many times the resources.
//
// Linux Kernel versions newer than >4.8 which disable locks in stat path can
// make this version faster.
//
// Use the performance tests to determine the correct walker for your platform
// and system!
func WalkParallel(root string, callback WalkFunc) {
	queue := fileQueue{queued: 0, finished: 0, items: channels.NewInfiniteChannel(), waiton: make(chan bool)}

	for w := 0; w <= 10; w++ {
		go examinePath(&queue, callback)
	}

	queue.add(root)
	queue.wait()
}
