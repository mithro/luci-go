// Copyright 2015 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package coordinator

import (
	"fmt"

	"github.com/luci/gae/service/info"
	luciConfig "github.com/luci/luci-go/common/config"
	log "github.com/luci/luci-go/common/logging"
	"golang.org/x/net/context"
)

// NamespaceAccessType specifies the type of namespace access that is being
// requested for WithProjectNamespace.
type NamespaceAccessType int

const (
	// NamespaceAccessNoAuth grants unconditional access to a project's namespace.
	// This bypasses all ACL checks, and must only be used by service endpoints
	// that explicitly apply ACLs elsewhere.
	NamespaceAccessNoAuth NamespaceAccessType = iota

	// NamespaceAccessREAD enforces READ permission access to a project's
	// namespace.
	NamespaceAccessREAD

	// NamespaceAccessWRITE enforces WRITE permission access to a project's
	// namespace.
	NamespaceAccessWRITE
)

type servicesKeyType int

// WithServices installs the supplied Services instance into a Context.
func WithServices(c context.Context, s Services) context.Context {
	return context.WithValue(c, servicesKeyType(0), s)
}

// GetServices gets the Services instance installed in the supplied Context.
//
// If no Services has been installed, it will panic.
func GetServices(c context.Context) Services {
	s, ok := c.Value(servicesKeyType(0)).(Services)
	if !ok {
		panic("no Services instance is installed")
	}
	return s
}

// WithProjectNamespace sets the current namespace to the project name.
//
// It will return an error if the project name or the project's namespace is
// invalid.
//
// If the current user does not have the requested permission for the project, a
// MembershipError will be returned.
func WithProjectNamespace(c *context.Context, project luciConfig.ProjectName, at NamespaceAccessType) error {
	ctx := *c

	if err := project.Validate(); err != nil {
		log.WithError(err).Errorf(ctx, "Project name is invalid.")
		return err
	}

	// Validate the current user has the requested access.
	switch at {
	case NamespaceAccessNoAuth:
		break

	case NamespaceAccessREAD:
		pcfg, err := GetServices(ctx).ProjectConfig(ctx, project)
		if err != nil {
			log.WithError(err).Errorf(ctx, "Failed to load project config.")
			return err
		}

		if err := IsProjectReader(*c, pcfg); err != nil {
			log.WithError(err).Errorf(*c, "User denied READ access to requested project.")
			return err
		}

	case NamespaceAccessWRITE:
		pcfg, err := GetServices(ctx).ProjectConfig(ctx, project)
		if err != nil {
			log.WithError(err).Errorf(ctx, "Failed to load project config.")
			return err
		}

		if err := IsProjectWriter(*c, pcfg); err != nil {
			log.WithError(err).Errorf(*c, "User denied WRITE access to requested project.")
			return err
		}

	default:
		return fmt.Errorf("unknown access type: %v", at)
	}

	pns := ProjectNamespace(project)
	nc, err := info.Get(ctx).Namespace(pns)
	if err != nil {
		log.Fields{
			log.ErrorKey: err,
			"project":    project,
			"namespace":  pns,
		}.Errorf(ctx, "Failed to set namespace.")
		return err
	}

	*c = nc
	return nil
}

// Project returns the current project installed in the supplied Context's
// namespace.
//
// This function is called with the expectation that the Context is in a
// namespace conforming to ProjectNamespace. If this is not the case, this
// method will panic.
func Project(c context.Context) luciConfig.ProjectName {
	ns, _ := info.Get(c).GetNamespace()
	project := ProjectFromNamespace(ns)
	if project != "" {
		return project
	}
	panic(fmt.Errorf("current namespace %q does not begin with project namespace prefix (%q)", ns, projectNamespacePrefix))
}
