// Package slasher implements slashing detection for eth2, able to catch slashable attestations
// and proposals that it receives via two event feeds, respectively. Any found slashings
// are then submitted to the beacon node's slashing operations pool. See the design document
// here https://hackmd.io/@prysmaticlabs/slasher.
package slasher

import (
	"context"
	"fmt"
	"time"

	types "github.com/prysmaticlabs/eth2-types"
	"github.com/prysmaticlabs/prysm/beacon-chain/blockchain"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/feed"
	statefeed "github.com/prysmaticlabs/prysm/beacon-chain/core/feed/state"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/beacon-chain/db"
	"github.com/prysmaticlabs/prysm/beacon-chain/operations/slashings"
	"github.com/prysmaticlabs/prysm/beacon-chain/state/stategen"
	"github.com/prysmaticlabs/prysm/beacon-chain/sync"
	ethpb "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/shared/event"
	"github.com/prysmaticlabs/prysm/shared/params"
	"github.com/prysmaticlabs/prysm/shared/slotutil"
)

// ServiceConfig for the slasher service in the beacon node.
// This struct allows us to specify required dependencies and
// parameters for slasher to function as needed.
type ServiceConfig struct {
	IndexedAttestationsFeed *event.Feed
	BeaconBlockHeadersFeed  *event.Feed
	Database                db.SlasherDatabase
	StateNotifier           statefeed.Notifier
	AttestationStateFetcher blockchain.AttestationStateFetcher
	StateGen                stategen.StateManager
	SlashingPoolInserter    slashings.PoolInserter
	HeadStateFetcher        blockchain.HeadFetcher
	SyncChecker             sync.Checker
}

// SlashingChecker is an interface for defining services that the beacon node may interact with to provide slashing data.
type SlashingChecker interface {
	IsSlashableBlock(ctx context.Context, proposal *ethpb.SignedBeaconBlockHeader) (*ethpb.ProposerSlashing, error)
	IsSlashableAttestation(ctx context.Context, attestation *ethpb.IndexedAttestation) ([]*ethpb.AttesterSlashing, error)
}

// Service defining a slasher implementation as part of
// the beacon node, able to detect eth2 slashable offenses.
type Service struct {
	params                 *Parameters
	serviceCfg             *ServiceConfig
	indexedAttsChan        chan *ethpb.IndexedAttestation
	beaconBlockHeadersChan chan *ethpb.SignedBeaconBlockHeader
	attsQueue              *attestationsQueue
	blksQueue              *blocksQueue
	ctx                    context.Context
	cancel                 context.CancelFunc
	slotTicker             *slotutil.SlotTicker
	genesisTime            time.Time
}

// New instantiates a new slasher from configuration values.
func New(ctx context.Context, srvCfg *ServiceConfig) (*Service, error) {
	ctx, cancel := context.WithCancel(ctx)
	return &Service{
		params:                 DefaultParams(),
		serviceCfg:             srvCfg,
		indexedAttsChan:        make(chan *ethpb.IndexedAttestation, 1),
		beaconBlockHeadersChan: make(chan *ethpb.SignedBeaconBlockHeader, 1),
		attsQueue:              newAttestationsQueue(),
		blksQueue:              newBlocksQueue(),
		ctx:                    ctx,
		cancel:                 cancel,
	}, nil
}

// Start listening for received indexed attestations and blocks
// and perform slashing detection on them.
func (s *Service) Start() {
	stateChannel := make(chan *feed.Event, 1)
	stateSub := s.serviceCfg.StateNotifier.StateFeed().Subscribe(stateChannel)
	stateEvent := <-stateChannel

	// Wait for us to receive the genesis time via a chain started notification.
	if stateEvent.Type == statefeed.ChainStarted {
		data, ok := stateEvent.Data.(*statefeed.ChainStartedData)
		if !ok {
			log.Error("Could not receive chain start notification, want *statefeed.ChainStartedData")
			return
		}
		s.genesisTime = data.StartTime
		log.WithField("genesisTime", s.genesisTime).Info("Starting slasher, received chain start event")
	} else if stateEvent.Type == statefeed.Initialized {
		// Alternatively, if the chain has already started, we then read the genesis
		// time value from this data.
		data, ok := stateEvent.Data.(*statefeed.InitializedData)
		if !ok {
			log.Error("Could not receive chain start notification, want *statefeed.ChainStartedData")
			return
		}
		s.genesisTime = data.StartTime
		log.WithField("genesisTime", s.genesisTime).Info("Starting slasher, chain already initialized")
	} else {
		// This should not happen.
		log.Error("Could start slasher, could not receive chain start event")
		return
	}

	stateSub.Unsubscribe()
	secondsPerSlot := params.BeaconConfig().SecondsPerSlot
	s.slotTicker = slotutil.NewSlotTicker(s.genesisTime, secondsPerSlot)

	s.waitForSync(s.genesisTime)

	s.waitForBackfill()

	log.Info("Completed chain sync, starting slashing detection")
	go s.processQueuedAttestations(s.ctx, s.slotTicker.C())
	go s.processQueuedBlocks(s.ctx, s.slotTicker.C())
	go s.receiveAttestations(s.ctx)
	go s.receiveBlocks(s.ctx)
	go s.pruneSlasherData(s.ctx, s.slotTicker.C())
}

func (s *Service) waitForBackfill() {
	wssPeriod := types.Epoch(4)
	headSlot := s.serviceCfg.HeadStateFetcher.HeadSlot()
	headEpoch := helpers.SlotToEpoch(headSlot)
	lookbackPeriod := headEpoch
	if lookbackPeriod > 4 {
		lookbackPeriod = lookbackPeriod - wssPeriod
	}
	fmt.Printf("Head epoch %d, lookback %d\n", headEpoch, lookbackPeriod)
	for {
		currentEpoch := slotutil.EpochsSinceGenesis(s.genesisTime)

		diff := currentEpoch
		if diff > lookbackPeriod {
			diff = diff - lookbackPeriod
		}

		// If we have no difference between the max epoch we have detected for
		// slasher and the current epoch on the clock, then we can exiit the loop.
		if diff == 0 {
			break
		}

		// We set the max epoch for slasher to the current epoch on the clock for backfilling.
		maxEpoch := currentEpoch

		s.backfill(lookbackPeriod, maxEpoch)

		// After backfilling, we set the lowest epoch for backfilling to be the
		// max epoch we have completed backfill to.
		lookbackPeriod = maxEpoch
	}
}

func (s *Service) backfill(start, end types.Epoch) {
	fmt.Printf("Backfilling from epoch %d to %d\n", start, end)
	for i := start; i < end; i++ {
	}
	time.Sleep(time.Second)
}

// Stop the slasher service.
func (s *Service) Stop() error {
	s.cancel()
	if s.slotTicker != nil {
		s.slotTicker.Done()
	}
	return nil
}

// Status of the slasher service.
func (s *Service) Status() error {
	return nil
}