// Copyright 2020 Anapaya Systems
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package control_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/scionproto/scion/gateway/control"
	"github.com/scionproto/scion/gateway/control/mock_control"
	"github.com/scionproto/scion/pkg/log/mock_log"
	"github.com/scionproto/scion/pkg/metrics"
	"github.com/scionproto/scion/pkg/private/xtest"
)

func TestRouterClose(t *testing.T) {
	router := control.Router{}
	assert.NoError(t, router.Close(context.Background()))
	assert.NoError(t, router.Run(context.Background()))
	assert.NoError(t, router.Close(context.Background()))
}

func TestRouterRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	rt := mock_control.NewMockRoutingTable(ctrl)
	logger := mock_log.NewMockLogger(ctrl)
	logger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	events := make(chan control.SessionEvent)
	router := control.Router{
		RoutingTable: rt,
		RoutingTableIndices: map[int][]uint8{
			1: {100, 101, 102, 103},
			2: {100, 102},
			3: {101, 103},
		},
		DataplaneSessions: map[uint8]control.PktWriter{
			100: testPktWriter{ID: 100},
			101: testPktWriter{ID: 101},
			102: testPktWriter{ID: 102},
			103: testPktWriter{ID: 103},
		},
		Events: events,
		Metrics: control.RouterMetrics{
			RoutingChainHealthy: func(routingChain int) metrics.Gauge { return nil },
			SessionsAlive:       func(routingChain int) metrics.Gauge { return nil },
			SessionChanges:      func(routingChain int) metrics.Counter { return nil },
			StateChanges:        func(routingChain int) metrics.Counter { return nil },
		},
	}
	errChan := make(chan error)
	go func() { errChan <- router.Run(context.Background()) }()

	t.Run("Calling Run twice fails", func(t *testing.T) {
		time.Sleep(20 * time.Millisecond)
		assert.Error(t, router.Run(context.Background()))
	})
	t.Run("Updating table works as expected", func(t *testing.T) {
		callChan := make(chan struct{})
		writeCallChan := func(_ int, _ control.PktWriter) error {
			callChan <- struct{}{}
			return nil
		}
		// When receiving the session 103 up event, the routing table should get
		// the session 103 information inserted.
		rt.EXPECT().SetSession(1, router.DataplaneSessions[103])
		rt.EXPECT().SetSession(3, router.DataplaneSessions[103]).Do(writeCallChan)

		events <- control.SessionEvent{SessionID: 103, Event: control.EventUp}
		xtest.AssertReadReturnsBefore(t, callChan, time.Second)

		// When receiving the session 102 up event, the routing table should get
		// the session 102 information inserted.
		rt.EXPECT().SetSession(1, router.DataplaneSessions[102])
		rt.EXPECT().SetSession(2, router.DataplaneSessions[102]).Do(writeCallChan)

		events <- control.SessionEvent{SessionID: 102, Event: control.EventUp}
		xtest.AssertReadReturnsBefore(t, callChan, time.Second)

		// When receiving the session 103 down event, the routing table should get
		// the session 103 information removed.
		rt.EXPECT().ClearSession(3).Do(func(int) error {
			callChan <- struct{}{}
			return nil
		})

		events <- control.SessionEvent{SessionID: 103, Event: control.EventDown}
		xtest.AssertReadReturnsBefore(t, callChan, time.Second)

		// When receiving the session 103 up event, the routing table should get
		// the session 103 information inserted.
		rt.EXPECT().SetSession(3, router.DataplaneSessions[103]).Do(writeCallChan)

		events <- control.SessionEvent{SessionID: 103, Event: control.EventUp}
		xtest.AssertReadReturnsBefore(t, callChan, time.Second)
	})

	err := router.Close(context.Background())
	assert.NoError(t, err)
	select {
	case err := <-errChan:
		assert.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatalf("Timeout waiting on run to complete")
	}
}
