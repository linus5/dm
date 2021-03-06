// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package syncer

import (
	"time"

	"github.com/pingcap/dm/pkg/log"
	"github.com/pingcap/errors"
	"golang.org/x/net/context"

	"github.com/pingcap/dm/dm/pb"
	"github.com/pingcap/dm/dm/unit"
	sm "github.com/pingcap/dm/syncer/safe-mode"
)

func (s *Syncer) enableSafeModeInitializationPhase(ctx context.Context, safeMode *sm.SafeMode) {
	safeMode.Reset() // in initialization phase, reset first
	safeMode.Add(1)  // try to enable

	if s.cfg.SafeMode {
		safeMode.Add(1) // add 1 but should no corresponding -1
		log.Info("[syncer] enable safe-mode by config")
	}

	go func() {
		defer func() {
			err := safeMode.Add(-1) // try to disable after 5 minutes
			if err != nil {
				// send error to the fatal chan to interrupt the process
				s.runFatalChan <- unit.NewProcessError(pb.ErrorType_UnknownError, errors.ErrorStack(err))
			}
		}()

		select {
		case <-ctx.Done():
		case <-time.After(5 * time.Minute):
		}
	}()
}
