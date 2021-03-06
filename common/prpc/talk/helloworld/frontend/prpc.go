// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package helloworld

import (
	"github.com/julienschmidt/httprouter"

	"github.com/luci/luci-go/server/discovery"
	"github.com/luci/luci-go/server/middleware"
	"github.com/luci/luci-go/server/prpc"

	"github.com/luci/luci-go/common/prpc/talk/helloworld/proto"
)

func InstallAPIRoutes(router *httprouter.Router, base middleware.Base) {
	server := &prpc.Server{}
	helloworld.RegisterGreeterServer(server, &greeterService{})
	discovery.Enable(server)
	server.InstallHandlers(router, base)
}
