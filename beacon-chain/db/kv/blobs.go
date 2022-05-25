package kv

import (
	"context"

	"github.com/pkg/errors"
	types "github.com/prysmaticlabs/prysm/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/encoding/bytesutil"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	bolt "go.etcd.io/bbolt"
	"go.opencensus.io/trace"
)

// SaveBlobsSidecar saves the blobs for a given epoch in the sidecar bucket.
func (s *Store) SaveBlobsSidecar(ctx context.Context, blob *ethpb.BlobsSidecar) error {
	_, span := trace.StartSpan(ctx, "BeaconDB.SaveBlobsSidecar")
	defer span.End()
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(blobsBucket)
		enc, err := encode(ctx, blob)
		if err != nil {
			return err
		}
		return bkt.Put(blob.BeaconBlockRoot, enc)
	})
}

func (s *Store) BlobsSidecar(ctx context.Context, blockRoot [32]byte) (*ethpb.BlobsSidecar, error) {
	ctx, span := trace.StartSpan(ctx, "BeaconDB.BlobsSidecar")
	defer span.End()

	var enc []byte
	if err := s.db.View(func(tx *bolt.Tx) error {
		enc = tx.Bucket(blobsBucket).Get(blockRoot[:])
		return nil
	}); err != nil {
		return nil, err
	}
	if len(enc) == 0 {
		return nil, nil
	}
	blob := &ethpb.BlobsSidecar{}
	if err := decode(ctx, enc, blob); err != nil {
		return nil, err
	}
	return blob, nil
}

func (s *Store) BlobsSidecarsBySlot(ctx context.Context, slot types.Slot) (bool, []*ethpb.BlobsSidecar, error) {
	ctx, span := trace.StartSpan(ctx, "BeaconDB.BlobsSidecarsBySlot")
	defer span.End()

	var blobsSidecars []*ethpb.BlobsSidecar
	err := s.db.View(func(tx *bolt.Tx) error {
		keys := blockRootsBySlot(ctx, tx, slot)
		blockRoots := make([][32]byte, 0, len(keys))
		for i := 0; i < len(keys); i++ {
			blockRoots[i] = bytesutil.ToBytes32(keys[i])
		}

		for _, blockRoot := range blockRoots {
			enc := tx.Bucket(blobsBucket).Get(blockRoot[:])
			if len(enc) == 0 {
				return nil
			}
			blobs := &ethpb.BlobsSidecar{}
			if err := decode(ctx, enc, blobs); err != nil {
				return err
			}
			blobsSidecars = append(blobsSidecars, blobs)
		}

		return nil
	})
	if err != nil {
		return false, nil, errors.Wrap(err, "could not retrieve blobs")
	}

	return true, blobsSidecars, nil
}
