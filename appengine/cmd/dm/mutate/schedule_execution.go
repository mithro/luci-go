// Copyright 2015 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package mutate

import (
	"github.com/luci/gae/service/datastore"
	"github.com/luci/luci-go/appengine/cmd/dm/model"
	"github.com/luci/luci-go/appengine/cmd/dm/types"
	"github.com/luci/luci-go/appengine/tumble"
	"golang.org/x/net/context"
)

// ScheduleExecution is a placeholder mutation that will be an entry into the
// Distributor scheduling state-machine.
type ScheduleExecution struct {
	For *types.AttemptID
}

// Root implements tumble.Mutation
func (s *ScheduleExecution) Root(c context.Context) *datastore.Key {
	return datastore.Get(c).KeyForObj(&model.Attempt{AttemptID: *s.For})
}

// RollForward implements tumble.Mutation
func (s *ScheduleExecution) RollForward(c context.Context) (muts []tumble.Mutation, err error) {
	return
}

func init() {
	tumble.Register((*ScheduleExecution)(nil))
}