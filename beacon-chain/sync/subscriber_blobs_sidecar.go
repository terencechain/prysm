package sync

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"google.golang.org/protobuf/proto"
)

// skipcq: SCC-U1000
func (s *Service) blobsSidecarSubscriber(ctx context.Context, msg proto.Message) error {
	// TODO(EIP-4844)
	m, ok := msg.(*ethpb.SignedBlobsSidecar)
	if !ok {
		return fmt.Errorf("message was not type *eth.SignedBlobsSidecar, type=%T", msg)
	}

	if m == nil {
		return errors.New("nil blobs sidecar message")
	}

	blob := m.Message
	return s.cfg.beaconDB.SaveBlobsSidecar(ctx, blob)
}
