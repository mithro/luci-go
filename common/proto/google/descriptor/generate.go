// Copyright 2015 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

//go:generate cproto -discovery=false -desc util_test.desc

package descriptor

import (
	"github.com/golang/protobuf/proto"
)

var _ = proto.Marshal
