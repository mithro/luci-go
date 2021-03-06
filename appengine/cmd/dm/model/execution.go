// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package model

import (
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"

	"golang.org/x/net/context"

	"github.com/luci/gae/service/datastore"
	"github.com/luci/luci-go/common/api/dm/service/v1"
	"github.com/luci/luci-go/common/grpcutil"
	"github.com/luci/luci-go/common/logging"
	google_pb "github.com/luci/luci-go/common/proto/google"
)

const ek = logging.ErrorKey

// Execution represents either an ongoing execution on the Quest's specified
// distributor, or is a placeholder for an already-completed Execution.
type Execution struct {
	ID      uint32         `gae:"$id"`
	Attempt *datastore.Key `gae:"$parent"`

	State       dm.Execution_State
	StateReason string `gae:",noindex"`

	Created          time.Time
	DistributorToken string
	DistributorURL   string `gae:",noindex"`

	// Token is a randomized nonce that's used to verify that RPCs verify from the
	// expected client (the client that's currently running the Execution). The
	// Token has 2 modes.
	//
	// When the Execution is handed to the distributor, the Token is randomly
	// generated by DM and passed to the distributor. The State of the Execution
	// starts as Scheduled. This token may be used by the client to "activate" the
	// Execution with the ActivateExecution rpc. At that point, the client
	// provides a new random token, the Execution State moves from Scheduled to
	// Running, and Token assumes the new value. As long as the Execution State is
	// running, the client may continue to use that new Token value to
	// authenticate other rpc's like AddDeps and FinishAttempt.
	//
	// As soon as the Execution is no longer supposed to have access, Token will
	// be nil'd out.
	Token []byte `gae:",noindex"`
}

// Revoke will null-out the Token and Put this Execution to the datastore.
func (e *Execution) Revoke(c context.Context) error {
	e.Token = nil
	return datastore.Get(c).Put(e)
}

func loadExecution(c context.Context, eid *dm.Execution_ID) (a *Attempt, e *Execution, err error) {
	ds := datastore.Get(c)

	a = &Attempt{ID: *eid.AttemptID()}
	e = &Execution{ID: eid.Id, Attempt: ds.KeyForObj(a)}
	err = datastore.Get(c).Get(a, e)

	if err != nil {
		err = fmt.Errorf("couldn't get attempt %v or its execution %d: %s", a.ID, e.ID, err)
		return
	}

	if a.CurExecution != e.ID {
		err = fmt.Errorf("verifying incorrect execution %d, expected %d", a.CurExecution, e.ID)
		return
	}
	return
}

func verifyExecutionAndCheckExTok(c context.Context, auth *dm.Execution_Auth) (a *Attempt, e *Execution, err error) {
	a, e, err = loadExecution(c, auth.Id)
	if err != nil {
		return
	}

	if a.State != dm.Attempt_EXECUTING {
		err = errors.New("Attempt is not executing")
		return
	}

	if e.State != dm.Execution_RUNNING {
		err = errors.New("Execution is not running")
		return
	}

	if subtle.ConstantTimeCompare(e.Token, auth.Token) != 1 {
		err = fmt.Errorf("incorrect Token: %x", hex.EncodeToString(auth.Token))
	}
	return
}

// AuthenticateExecution verifies that the Attempt is executing, and that evkey
// matches the execution key of the current Execution for this Attempt.
//
// As a bonus, it will return the loaded Attempt and Execution.
func AuthenticateExecution(c context.Context, auth *dm.Execution_Auth) (a *Attempt, e *Execution, err error) {
	a, e, err = verifyExecutionAndCheckExTok(c, auth)
	if err != nil {
		logging.Fields{ek: err, "eid": auth.Id}.Errorf(c, "failed to verify execution")
		err = grpcutil.Errf(codes.Unauthenticated, "requires execution Auth")
	}
	return a, e, err
}

// InvalidateExecution verifies that the execution key is valid, and then
// revokes the execution key.
//
// As a bonus, it will return the loaded Attempt and Execution.
func InvalidateExecution(c context.Context, auth *dm.Execution_Auth) (a *Attempt, e *Execution, err error) {
	if a, e, err = verifyExecutionAndCheckExTok(c, auth); err != nil {
		logging.Fields{ek: err, "eid": auth.Id}.Errorf(c, "failed to verify execution")
		err = grpcutil.Errf(codes.Unauthenticated, "requires execution Auth")
		return
	}

	err = e.Revoke(c)
	if err != nil {
		logging.Fields{ek: err, "eid": auth.Id}.Errorf(c, "failed to revoke execution")
		err = grpcutil.Errf(codes.Internal, "unable to invalidate Auth")
	}
	return
}

func verifyExecutionAndActivate(c context.Context, auth *dm.Execution_Auth, actTok []byte) (a *Attempt, e *Execution, err error) {
	a, e, err = loadExecution(c, auth.Id)
	if err != nil {
		return
	}

	if a.State != dm.Attempt_EXECUTING {
		err = errors.New("Attempt is in wrong state")
		return
	}

	switch e.State {
	case dm.Execution_SCHEDULED:
		if subtle.ConstantTimeCompare(e.Token, auth.Token) != 1 {
			err = errors.New("incorrect ActivationToken")
			return
		}

		e.State.MustEvolve(dm.Execution_RUNNING)
		e.Token = actTok
		err = datastore.Get(c).Put(e)

	case dm.Execution_RUNNING:
		if subtle.ConstantTimeCompare(e.Token, actTok) != 1 {
			err = errors.New("incorrect Token")
		}
		// either the Token matched, in which case this is simply a retry
		// by the same client, so there's no error, or it's wrong which means it's
		// a retry by a different client.

	default:
		err = fmt.Errorf("Execution is in wrong state")
	}
	return
}

// ActivateExecution validates that the execution is unactivated and that
// the activation token matches and then sets the token to the new
// value.
//
// It's OK to retry this. Subsequent invocations with the same Token
// will recognize this case and not return an error.
func ActivateExecution(c context.Context, auth *dm.Execution_Auth, actToken []byte) (a *Attempt, e *Execution, err error) {
	a, e, err = verifyExecutionAndActivate(c, auth, actToken)
	if err != nil {
		logging.Fields{ek: err, "eid": auth.Id}.Errorf(c, "failed to activate execution")
		err = grpcutil.Errf(codes.Unauthenticated, "failed to activate execution Auth")
	}
	return a, e, err
}

// ToProto returns a dm proto version of this Execution.
func (e *Execution) ToProto(includeID bool) *dm.Execution {
	ret := &dm.Execution{
		Data: &dm.Execution_Data{
			State:              e.State,
			StateReason:        e.StateReason,
			Created:            google_pb.NewTimestamp(e.Created),
			DistributorToken:   e.DistributorToken,
			DistributorInfoUrl: e.DistributorURL,
		},
	}
	if includeID {
		aid := &dm.Attempt_ID{}
		if err := aid.SetDMEncoded(e.Attempt.StringID()); err != nil {
			panic(err)
		}
		ret.Id = dm.NewExecutionID(aid.Quest, aid.Id, e.ID)
	}
	return ret
}
