package kv

import (
	"bytes"
	"context"
	"fmt"

	"github.com/prysmaticlabs/prysm/v3/config/params"
	types "github.com/prysmaticlabs/prysm/v3/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v3/encoding/bytesutil"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	bolt "go.etcd.io/bbolt"
	"go.opencensus.io/trace"
)

const blobSidecarKeyLength = 48 // slot_to_rotating_buffer(blob.slot) ++ blob.slot ++ blob.block_root

// SaveBlobsSidecar saves the blobs for a given epoch in the sidecar bucket. When we receive a blob:
// 1. Convert slot using a modulo operator to [0, maxSlots] where maxSlots = MAX_BLOB_EPOCHS*SLOTS_PER_EPOCH
// 2. Compute key for blob as bytes(slot_to_rotating_buffer(blob.slot)) ++ bytes(blob.slot) ++ blob.block_root
// 3. Begin the save algorithm:  If the incoming blob has a slot bigger than the saved slot at the spot
//    in the rotating keys buffer, we overwrite all elements for that slot.
//
//    firstElemKey = getFirstElement(bucket)
//    shouldOverwrite = blob.slot > bytes_to_slot(firstElemKey[8:16])
//    if shouldOverwrite:
// 	    for existingKey := seek prefix bytes(slot_to_rotating_buffer(blob.slot))
//        bucket.delete(existingKey)
//    bucket.put(key, blob)
func (s *Store) SaveBlobsSidecar(ctx context.Context, blobSidecar *ethpb.BlobsSidecar) error {
	ctx, span := trace.StartSpan(ctx, "BeaconDB.SaveBlobsSidecar")
	defer span.End()
	return s.db.Update(func(tx *bolt.Tx) error {
		encodedBlobSidecar, err := encode(ctx, blobSidecar)
		if err != nil {
			return err
		}
		bkt := tx.Bucket(blobsBucket)
		c := bkt.Cursor()
		key := blobSidecarKey(blobSidecar)
		rotatingBufferPrefix := key[0:8]
		var firstElementKey []byte
		for k, _ := c.Seek(rotatingBufferPrefix); bytes.HasPrefix(k, rotatingBufferPrefix); k, _ = c.Next() {
			if len(k) != 0 {
				firstElementKey = k
				break
			}
		}
		// If there is no element stored at blob.slot % MAX_SLOTS_TO_PERSIST_BLOBS, then we simply
		// store the blob by key and exit early.
		if len(firstElementKey) == 0 {
			return bkt.Put(key, encodedBlobSidecar)
		} else if len(firstElementKey) != len(key) {
			return fmt.Errorf(
				"key length %d (%#x) != existing key length %d (%#x)",
				len(key),
				key,
				len(firstElementKey),
				firstElementKey,
			)
		}
		slotOfFirstElement := firstElementKey[8:16]
		// If we should overwrite old blobs at the spot in the rotating buffer, we clear data at that spot.
		shouldOverwrite := blobSidecar.BeaconBlockSlot > bytesutil.BytesToSlotBigEndian(slotOfFirstElement)
		if shouldOverwrite {
			for k, _ := c.Seek(rotatingBufferPrefix); bytes.HasPrefix(k, rotatingBufferPrefix); k, _ = c.Next() {
				if err := bkt.Delete(k); err != nil {
					log.WithError(err).Warnf("Could not delete blob with key %#x", k)
				}
			}
		}
		return bkt.Put(key, encodedBlobSidecar)
	})
}

// BlobsSidecar retrieves the blobs given a beacon block root.
func (s *Store) BlobsSidecar(ctx context.Context, beaconBlockRoot [32]byte) (*ethpb.BlobsSidecar, error) {
	ctx, span := trace.StartSpan(ctx, "BeaconDB.BlobsSidecar")
	defer span.End()

	var enc []byte
	if err := s.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(blobsBucket).Cursor()
		// Bucket size is bounded and bolt cursors are fast. Moreover, a thin caching layer can be added.
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if bytes.HasSuffix(k, beaconBlockRoot[:]) {
				enc = v
				break
			}
		}
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

// BlobsSidecarsBySlot retrieves sidecars from a slot.
func (s *Store) BlobsSidecarsBySlot(ctx context.Context, slot types.Slot) ([]*ethpb.BlobsSidecar, error) {
	ctx, span := trace.StartSpan(ctx, "BeaconDB.BlobsSidecarsBySlot")
	defer span.End()
	encodedItems := make([][]byte, 0)
	if err := s.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(blobsBucket).Cursor()
		// Bucket size is bounded and bolt cursors are fast. Moreover, a thin caching layer can be added.
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if len(k) != blobSidecarKeyLength {
				continue
			}
			slotInKey := bytesutil.BytesToSlotBigEndian(k[8:16])
			if slotInKey == slot {
				encodedItems = append(encodedItems, v)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	sidecars := make([]*ethpb.BlobsSidecar, len(encodedItems))
	if len(encodedItems) == 0 {
		return sidecars, nil
	}
	for i, enc := range encodedItems {
		blob := &ethpb.BlobsSidecar{}
		if err := decode(ctx, enc, blob); err != nil {
			return nil, err
		}
		sidecars[i] = blob
	}
	return sidecars, nil
}

// HasBlobsSidecar returns true if the blobs are in the db.
func (s *Store) HasBlobsSidecar(ctx context.Context, beaconBlockRoot [32]byte) bool {
	ctx, span := trace.StartSpan(ctx, "BeaconDB.HasBlobsSidecar")
	defer span.End()
	blobSidecar, err := s.BlobsSidecar(ctx, beaconBlockRoot)
	if err != nil {
		return false
	}
	return blobSidecar != nil
}

// DeleteBlobsSidecar returns true if the blobs are in the db.
func (s *Store) DeleteBlobsSidecar(ctx context.Context, beaconBlockRoot [32]byte) error {
	ctx, span := trace.StartSpan(ctx, "BeaconDB.DeleteBlobsSidecar")
	defer span.End()
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(blobsBucket)
		c := bkt.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			if bytes.HasSuffix(k, beaconBlockRoot[:]) {
				if err := bkt.Delete(k); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// We define a blob sidecar key as: bytes(slot_to_rotating_buffer(blob.slot)) ++ bytes(blob.slot) ++ blob.block_root
// where slot_to_rotating_buffer(slot) = slot % MAX_SLOTS_TO_PERSIST_BLOBS.
func blobSidecarKey(blob *ethpb.BlobsSidecar) []byte {
	slotsPerEpoch := params.BeaconConfig().SlotsPerEpoch
	maxEpochsToPersistBlobs := params.BeaconNetworkConfig().MinEpochsForBlobsSidecarsRequest
	maxSlotsToPersistBlobs := types.Slot(maxEpochsToPersistBlobs.Mul(uint64(slotsPerEpoch)))
	slotInRotatingBuffer := blob.BeaconBlockSlot.ModSlot(maxSlotsToPersistBlobs)
	key := bytesutil.SlotToBytesBigEndian(slotInRotatingBuffer)
	key = append(key, bytesutil.SlotToBytesBigEndian(blob.BeaconBlockSlot)...)
	key = append(key, blob.BeaconBlockRoot...)
	return key
}
