package sync

import (
	"context"
	"time"

	libp2pcore "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
	p2ptypes "github.com/prysmaticlabs/prysm/beacon-chain/p2p/types"
	"github.com/prysmaticlabs/prysm/config/params"
	"github.com/prysmaticlabs/prysm/monitoring/tracing"
	pb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"go.opencensus.io/trace"
)

// We assume a cost of 1 MiB per sidecar responded to a range request.
const avgSidecarBlobsTransferBytes = 1 << 10

// blobsSidecarsByRangeRPCHandler looks up the request blobs from the database from a given start slot index
func (s *Service) blobsSidecarsByRangeRPCHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) error {
	ctx, span := trace.StartSpan(ctx, "sync.BlobsSidecarsByRangeHandler")
	defer span.End()
	ctx, cancel := context.WithTimeout(ctx, respTimeout)
	defer cancel()
	SetRPCStreamDeadlines(stream)

	// Ticker to stagger out large requests.
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	r, ok := msg.(*pb.BlobsSidecarsByRangeRequest)
	if !ok {
		return errors.New("message is not type *pb.BlobsSidecarsByRangeRequest")
	}

	startSlot := r.StartSlot
	count := r.Count
	endSlot := startSlot.Add(count)

	var numBlobs uint64
	maxRequestBlobsSidecars := params.BeaconNetworkConfig().MaxRequestBlobsSidecars
	for slot := startSlot; slot < endSlot && numBlobs < maxRequestBlobsSidecars; slot.Add(1) {
		if err := s.rateLimiter.validateRequest(stream, uint64(avgSidecarBlobsTransferBytes)); err != nil {
			return err
		}

		exists, sidecars, err := s.cfg.beaconDB.BlobsSidecarsBySlot(ctx, slot)
		if err != nil {
			s.writeErrorResponseToStream(responseCodeServerError, p2ptypes.ErrGeneric.Error(), stream)
			tracing.AnnotateError(span, err)
			return err
		}
		if !exists {
			continue
		}

		for _, blobs := range sidecars {
			SetStreamWriteDeadline(stream, defaultWriteDuration)
			if chunkErr := WriteBlobsSidecarChunk(stream, s.cfg.chain, s.cfg.p2p.Encoding(), blobs); chunkErr != nil {
				log.WithError(chunkErr).Debug("Could not send a chunked response")
				s.writeErrorResponseToStream(responseCodeServerError, p2ptypes.ErrGeneric.Error(), stream)
				tracing.AnnotateError(span, chunkErr)
				return chunkErr
			}
		}

		numBlobs++
		s.rateLimiter.add(stream, int64(avgSidecarBlobsTransferBytes))

		// TODO(inphi): adjust ticker based on bucket capacity
		<-ticker.C
	}

	closeStream(stream, log)
	return nil
}

func (s *Service) sendBlobsSidecarsByRangeRequest(ctx context.Context, id peer.ID) error {
	// TODO(EIP-4844)
	return nil
}

func blobSize()
