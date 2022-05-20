package sync

import (
	"context"

	libp2pcore "github.com/libp2p/go-libp2p-core"
	"github.com/pkg/errors"
	pb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"go.opencensus.io/trace"
)

func (s *Service) blobsSidecarsByRangeRPCHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) error {
	// TODO(EIP-4844)
	ctx, span := trace.StartSpan(ctx, "sync.BlobsSidecarsByRangeHandler")
	defer span.End()

	_, ok := msg.(*pb.BlobsSidecarsByRangeRequest)
	if !ok {
		return errors.New("message is not type *pb.BlobsSidecarsByRangeRequest")
	}

	return nil
}
