// Copyright 2025 SCION Association, Anapaya Systems
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

package connect

import (
	"context"
	"net"

	"connectrpc.com/connect"
	"github.com/quic-go/quic-go/http3"

	"github.com/scionproto/scion/control/beaconing"
	"github.com/scionproto/scion/control/onehop"
	"github.com/scionproto/scion/pkg/addr"
	libconnect "github.com/scionproto/scion/pkg/connect"
	control_plane "github.com/scionproto/scion/pkg/proto/control_plane"
	"github.com/scionproto/scion/pkg/proto/control_plane/v1/control_planeconnect"
	seg "github.com/scionproto/scion/pkg/segment"
	"github.com/scionproto/scion/pkg/snet/squic"
)

type BeaconSenderFactory struct {
	Dialer libconnect.Dialer
}

func (f *BeaconSenderFactory) NewSender(
	ctx context.Context,
	dstIA addr.IA,
	egIfID uint16,
	nextHop *net.UDPAddr,
) (beaconing.Sender, error) {
	addr := &onehop.Addr{
		IA:      dstIA,
		Egress:  egIfID,
		SVC:     addr.SvcCS,
		NextHop: nextHop,
	}
	dialer := f.Dialer(addr)
	return &BeaconSender{
		Addr: "https://" + addr.SVC.BaseString(),
		Client: &libconnect.HTTPClient{
			RoundTripper: &http3.Transport{
				Dial: dialer.DialEarly,
			},
		},
	}, nil

}

type BeaconSender struct {
	Addr   string
	Client *libconnect.HTTPClient
}

func (s BeaconSender) Send(ctx context.Context, b *seg.PathSegment) error {
	client := control_planeconnect.NewSegmentCreationServiceClient(s.Client, s.Addr)
	_, err := client.Beacon(ctx, connect.NewRequest(&control_plane.BeaconRequest{
		Segment: seg.PathSegmentToPB(b),
	}))
	return err
}

// Close closes the BeaconSender and releases all underlying resources.
func (s BeaconSender) Close() error {
	return s.Client.RoundTripper.Close()
}

// Registrar registers segments.
type Registrar struct {
	Dialer libconnect.Dialer
}

// RegisterSegment registers a segment with the remote.
func (r Registrar) RegisterSegment(ctx context.Context, meta seg.Meta, remote net.Addr) error {
	peer := make(chan net.Addr, 1)
	dialer := r.Dialer(remote, squic.WithPeerChannel(peer))
	client := control_planeconnect.NewSegmentRegistrationServiceClient(
		libconnect.HTTPClient{
			RoundTripper: &http3.Transport{
				Dial: dialer.DialEarly,
			},
		},
		libconnect.BaseUrl(remote),
	)
	_, err := client.SegmentsRegistration(ctx,
		connect.NewRequest(&control_plane.SegmentsRegistrationRequest{
			Segments: map[int32]*control_plane.SegmentsRegistrationRequest_Segments{
				int32(meta.Type): {
					Segments: []*control_plane.PathSegment{
						seg.PathSegmentToPB(meta.Segment),
					},
				},
			},
		}),
	)
	return err
}
