package blocks

import (
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v3/encoding/bytesutil"
	enginev1 "github.com/prysmaticlabs/prysm/v3/proto/engine/v1"
	eth "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v3/runtime/version"
	"google.golang.org/protobuf/proto"
)

// Proto converts the signed beacon block to a protobuf object.
func (b *SignedBeaconBlock) Proto() (proto.Message, error) {
	if b == nil {
		return nil, errNilBlock
	}

	blockMessage, err := b.block.Proto()
	if err != nil {
		return nil, err
	}

	switch b.version {
	case version.Phase0:
		var block *eth.BeaconBlock
		if blockMessage != nil {
			var ok bool
			block, ok = blockMessage.(*eth.BeaconBlock)
			if !ok {
				return nil, errIncorrectBlockVersion
			}
		}
		return &eth.SignedBeaconBlock{
			Block:     block,
			Signature: b.signature[:],
		}, nil
	case version.Altair:
		var block *eth.BeaconBlockAltair
		if blockMessage != nil {
			var ok bool
			block, ok = blockMessage.(*eth.BeaconBlockAltair)
			if !ok {
				return nil, errIncorrectBlockVersion
			}
		}
		return &eth.SignedBeaconBlockAltair{
			Block:     block,
			Signature: b.signature[:],
		}, nil
	case version.Bellatrix:
		if b.IsBlinded() {
			var block *eth.BlindedBeaconBlockBellatrix
			if blockMessage != nil {
				var ok bool
				block, ok = blockMessage.(*eth.BlindedBeaconBlockBellatrix)
				if !ok {
					return nil, errIncorrectBlockVersion
				}
			}
			return &eth.SignedBlindedBeaconBlockBellatrix{
				Block:     block,
				Signature: b.signature[:],
			}, nil
		}
		var block *eth.BeaconBlockBellatrix
		if blockMessage != nil {
			var ok bool
			block, ok = blockMessage.(*eth.BeaconBlockBellatrix)
			if !ok {
				return nil, errIncorrectBlockVersion
			}
		}
		return &eth.SignedBeaconBlockBellatrix{
			Block:     block,
			Signature: b.signature[:],
		}, nil
	case version.EIP4844:
		var block *eth.BeaconBlockWithBlobKZGs
		if blockMessage != nil {
			var ok bool
			block, ok = blockMessage.(*eth.BeaconBlockWithBlobKZGs)
			if !ok {
				return nil, errIncorrectBlockVersion
			}
		}
		return &eth.SignedBeaconBlockWithBlobKZGs{Block: block, Signature: b.signature[:]}, nil
	default:
		return nil, errors.New("unsupported signed beacon block version")
	}
}

// Proto converts the beacon block to a protobuf object.
func (b *BeaconBlock) Proto() (proto.Message, error) {
	if b == nil {
		return nil, nil
	}

	bodyMessage, err := b.body.Proto()
	if err != nil {
		return nil, err
	}

	switch b.version {
	case version.Phase0:
		var body *eth.BeaconBlockBody
		if bodyMessage != nil {
			var ok bool
			body, ok = bodyMessage.(*eth.BeaconBlockBody)
			if !ok {
				return nil, errIncorrectBodyVersion
			}
		}
		return &eth.BeaconBlock{
			Slot:          b.slot,
			ProposerIndex: b.proposerIndex,
			ParentRoot:    b.parentRoot[:],
			StateRoot:     b.stateRoot[:],
			Body:          body,
		}, nil
	case version.Altair:
		var body *eth.BeaconBlockBodyAltair
		if bodyMessage != nil {
			var ok bool
			body, ok = bodyMessage.(*eth.BeaconBlockBodyAltair)
			if !ok {
				return nil, errIncorrectBodyVersion
			}
		}
		return &eth.BeaconBlockAltair{
			Slot:          b.slot,
			ProposerIndex: b.proposerIndex,
			ParentRoot:    b.parentRoot[:],
			StateRoot:     b.stateRoot[:],
			Body:          body,
		}, nil
	case version.Bellatrix:
		if b.IsBlinded() {
			var body *eth.BlindedBeaconBlockBodyBellatrix
			if bodyMessage != nil {
				var ok bool
				body, ok = bodyMessage.(*eth.BlindedBeaconBlockBodyBellatrix)
				if !ok {
					return nil, errIncorrectBodyVersion
				}
			}
			return &eth.BlindedBeaconBlockBellatrix{
				Slot:          b.slot,
				ProposerIndex: b.proposerIndex,
				ParentRoot:    b.parentRoot[:],
				StateRoot:     b.stateRoot[:],
				Body:          body,
			}, nil
		}
		var body *eth.BeaconBlockBodyBellatrix
		if bodyMessage != nil {
			var ok bool
			body, ok = bodyMessage.(*eth.BeaconBlockBodyBellatrix)
			if !ok {
				return nil, errIncorrectBodyVersion
			}
		}
		return &eth.BeaconBlockBellatrix{
			Slot:          b.slot,
			ProposerIndex: b.proposerIndex,
			ParentRoot:    b.parentRoot[:],
			StateRoot:     b.stateRoot[:],
			Body:          body,
		}, nil
	case version.EIP4844:
		var body *eth.BeaconBlockBodyWithBlobKZGs
		if bodyMessage != nil {
			var ok bool
			body, ok = bodyMessage.(*eth.BeaconBlockBodyWithBlobKZGs)
			if !ok {
				return nil, errIncorrectBodyVersion
			}
		}
		return &eth.BeaconBlockWithBlobKZGs{
			Slot:          b.slot,
			ProposerIndex: b.proposerIndex,
			ParentRoot:    b.parentRoot[:],
			StateRoot:     b.stateRoot[:],
			Body:          body,
		}, nil
	default:
		return nil, errors.New("unsupported beacon block version")
	}
}

// Proto converts the beacon block body to a protobuf object.
func (b *BeaconBlockBody) Proto() (proto.Message, error) {
	if b == nil {
		return nil, nil
	}

	switch b.version {
	case version.Phase0:
		return &eth.BeaconBlockBody{
			RandaoReveal:      b.randaoReveal[:],
			Eth1Data:          b.eth1Data,
			Graffiti:          b.graffiti[:],
			ProposerSlashings: b.proposerSlashings,
			AttesterSlashings: b.attesterSlashings,
			Attestations:      b.attestations,
			Deposits:          b.deposits,
			VoluntaryExits:    b.voluntaryExits,
		}, nil
	case version.Altair:
		return &eth.BeaconBlockBodyAltair{
			RandaoReveal:      b.randaoReveal[:],
			Eth1Data:          b.eth1Data,
			Graffiti:          b.graffiti[:],
			ProposerSlashings: b.proposerSlashings,
			AttesterSlashings: b.attesterSlashings,
			Attestations:      b.attestations,
			Deposits:          b.deposits,
			VoluntaryExits:    b.voluntaryExits,
			SyncAggregate:     b.syncAggregate,
		}, nil
	case version.Bellatrix:
		if b.isBlinded {
			payloadHeader, err := b.executionDataHeader.PbGenericPayloadHeader()
			if err != nil {
				return nil, err
			}
			return &eth.BlindedBeaconBlockBodyBellatrix{
				RandaoReveal:           b.randaoReveal[:],
				Eth1Data:               b.eth1Data,
				Graffiti:               b.graffiti[:],
				ProposerSlashings:      b.proposerSlashings,
				AttesterSlashings:      b.attesterSlashings,
				Attestations:           b.attestations,
				Deposits:               b.deposits,
				VoluntaryExits:         b.voluntaryExits,
				SyncAggregate:          b.syncAggregate,
				ExecutionPayloadHeader: payloadHeader,
			}, nil
		}
		payload, err := b.executionData.PbGenericPayload()
		if err != nil {
			return nil, err
		}
		return &eth.BeaconBlockBodyBellatrix{
			RandaoReveal:      b.randaoReveal[:],
			Eth1Data:          b.eth1Data,
			Graffiti:          b.graffiti[:],
			ProposerSlashings: b.proposerSlashings,
			AttesterSlashings: b.attesterSlashings,
			Attestations:      b.attestations,
			Deposits:          b.deposits,
			VoluntaryExits:    b.voluntaryExits,
			SyncAggregate:     b.syncAggregate,
			ExecutionPayload:  payload,
		}, nil
	case version.EIP4844:
		// TODO(EIP-4844): Blinded blocks
		payload, err := b.executionData.PbEip4844Payload()
		if err != nil {
			return nil, err
		}
		return &eth.BeaconBlockBodyWithBlobKZGs{
			RandaoReveal:      b.randaoReveal[:],
			Eth1Data:          b.eth1Data,
			Graffiti:          b.graffiti[:],
			ProposerSlashings: b.proposerSlashings,
			AttesterSlashings: b.attesterSlashings,
			Attestations:      b.attestations,
			Deposits:          b.deposits,
			VoluntaryExits:    b.voluntaryExits,
			SyncAggregate:     b.syncAggregate,
			ExecutionPayload:  payload,
			BlobKzgs:          b.blobKzgs,
		}, nil
	default:
		return nil, errors.New("unsupported beacon block body version")
	}
}

func initSignedBlockFromProtoPhase0(pb *eth.SignedBeaconBlock) (*SignedBeaconBlock, error) {
	if pb == nil {
		return nil, errNilBlock
	}

	block, err := initBlockFromProtoPhase0(pb.Block)
	if err != nil {
		return nil, err
	}
	b := &SignedBeaconBlock{
		version:   version.Phase0,
		block:     block,
		signature: bytesutil.ToBytes96(pb.Signature),
	}
	return b, nil
}

func initSignedBlockFromProtoAltair(pb *eth.SignedBeaconBlockAltair) (*SignedBeaconBlock, error) {
	if pb == nil {
		return nil, errNilBlock
	}

	block, err := initBlockFromProtoAltair(pb.Block)
	if err != nil {
		return nil, err
	}
	b := &SignedBeaconBlock{
		version:   version.Altair,
		block:     block,
		signature: bytesutil.ToBytes96(pb.Signature),
	}
	return b, nil
}

func initSignedBlockFromProtoBellatrix(pb *eth.SignedBeaconBlockBellatrix) (*SignedBeaconBlock, error) {
	if pb == nil {
		return nil, errNilBlock
	}

	block, err := initBlockFromProtoBellatrix(pb.Block)
	if err != nil {
		return nil, err
	}
	b := &SignedBeaconBlock{
		version:   version.Bellatrix,
		block:     block,
		signature: bytesutil.ToBytes96(pb.Signature),
	}
	return b, nil
}

func initSignedBlockFromProtoEip4844(pb *eth.SignedBeaconBlockWithBlobKZGs) (*SignedBeaconBlock, error) {
	if pb == nil {
		return nil, errNilBlock
	}

	block, err := initBlockFromProtoEip4844(pb.Block)
	if err != nil {
		return nil, err
	}
	b := &SignedBeaconBlock{
		version:   version.EIP4844,
		block:     block,
		signature: bytesutil.ToBytes96(pb.Signature),
	}
	return b, nil
}

func initBlindedSignedBlockFromProtoBellatrix(pb *eth.SignedBlindedBeaconBlockBellatrix) (*SignedBeaconBlock, error) {
	if pb == nil {
		return nil, errNilBlock
	}

	block, err := initBlindedBlockFromProtoBellatrix(pb.Block)
	if err != nil {
		return nil, err
	}
	b := &SignedBeaconBlock{
		version:   version.Bellatrix,
		block:     block,
		signature: bytesutil.ToBytes96(pb.Signature),
	}
	return b, nil
}

func initBlockFromProtoPhase0(pb *eth.BeaconBlock) (*BeaconBlock, error) {
	if pb == nil {
		return nil, errNilBlock
	}

	body, err := initBlockBodyFromProtoPhase0(pb.Body)
	if err != nil {
		return nil, err
	}
	b := &BeaconBlock{
		version:       version.Phase0,
		slot:          pb.Slot,
		proposerIndex: pb.ProposerIndex,
		parentRoot:    bytesutil.ToBytes32(pb.ParentRoot),
		stateRoot:     bytesutil.ToBytes32(pb.StateRoot),
		body:          body,
	}
	return b, nil
}

func initBlockFromProtoAltair(pb *eth.BeaconBlockAltair) (*BeaconBlock, error) {
	if pb == nil {
		return nil, errNilBlock
	}

	body, err := initBlockBodyFromProtoAltair(pb.Body)
	if err != nil {
		return nil, err
	}
	b := &BeaconBlock{
		version:       version.Altair,
		slot:          pb.Slot,
		proposerIndex: pb.ProposerIndex,
		parentRoot:    bytesutil.ToBytes32(pb.ParentRoot),
		stateRoot:     bytesutil.ToBytes32(pb.StateRoot),
		body:          body,
	}
	return b, nil
}

func initBlockFromProtoBellatrix(pb *eth.BeaconBlockBellatrix) (*BeaconBlock, error) {
	if pb == nil {
		return nil, errNilBlock
	}

	body, err := initBlockBodyFromProtoBellatrix(pb.Body)
	if err != nil {
		return nil, err
	}
	b := &BeaconBlock{
		version:       version.Bellatrix,
		slot:          pb.Slot,
		proposerIndex: pb.ProposerIndex,
		parentRoot:    bytesutil.ToBytes32(pb.ParentRoot),
		stateRoot:     bytesutil.ToBytes32(pb.StateRoot),
		body:          body,
	}
	return b, nil
}

func initBlindedBlockFromProtoBellatrix(pb *eth.BlindedBeaconBlockBellatrix) (*BeaconBlock, error) {
	if pb == nil {
		return nil, errNilBlock
	}

	body, err := initBlindedBlockBodyFromProtoBellatrix(pb.Body)
	if err != nil {
		return nil, err
	}
	b := &BeaconBlock{
		version:       version.Bellatrix,
		slot:          pb.Slot,
		proposerIndex: pb.ProposerIndex,
		parentRoot:    bytesutil.ToBytes32(pb.ParentRoot),
		stateRoot:     bytesutil.ToBytes32(pb.StateRoot),
		body:          body,
	}
	return b, nil
}

func initBlockFromProtoEip4844(pb *eth.BeaconBlockWithBlobKZGs) (*BeaconBlock, error) {
	if pb == nil {
		return nil, errNilBlock
	}

	body, err := initBlockBodyFromProtoEip4844(pb.Body)
	if err != nil {
		return nil, err
	}
	b := &BeaconBlock{
		version:       version.EIP4844,
		slot:          pb.Slot,
		proposerIndex: pb.ProposerIndex,
		parentRoot:    bytesutil.ToBytes32(pb.ParentRoot),
		stateRoot:     bytesutil.ToBytes32(pb.StateRoot),
		body:          body,
	}
	return b, nil
}

func initBlockBodyFromProtoPhase0(pb *eth.BeaconBlockBody) (*BeaconBlockBody, error) {
	if pb == nil {
		return nil, errNilBlockBody
	}

	b := &BeaconBlockBody{
		version:           version.Phase0,
		isBlinded:         false,
		randaoReveal:      bytesutil.ToBytes96(pb.RandaoReveal),
		eth1Data:          pb.Eth1Data,
		graffiti:          bytesutil.ToBytes32(pb.Graffiti),
		proposerSlashings: pb.ProposerSlashings,
		attesterSlashings: pb.AttesterSlashings,
		attestations:      pb.Attestations,
		deposits:          pb.Deposits,
		voluntaryExits:    pb.VoluntaryExits,
	}
	return b, nil
}

func initBlockBodyFromProtoAltair(pb *eth.BeaconBlockBodyAltair) (*BeaconBlockBody, error) {
	if pb == nil {
		return nil, errNilBlockBody
	}

	b := &BeaconBlockBody{
		version:           version.Altair,
		isBlinded:         false,
		randaoReveal:      bytesutil.ToBytes96(pb.RandaoReveal),
		eth1Data:          pb.Eth1Data,
		graffiti:          bytesutil.ToBytes32(pb.Graffiti),
		proposerSlashings: pb.ProposerSlashings,
		attesterSlashings: pb.AttesterSlashings,
		attestations:      pb.Attestations,
		deposits:          pb.Deposits,
		voluntaryExits:    pb.VoluntaryExits,
		syncAggregate:     pb.SyncAggregate,
	}
	return b, nil
}

func initBlockBodyFromProtoBellatrix(pb *eth.BeaconBlockBodyBellatrix) (*BeaconBlockBody, error) {
	if pb == nil {
		return nil, errNilBlockBody
	}

	executionData, err := NewExecutionData(pb.ExecutionPayload)
	if err != nil {
		return nil, err
	}

	b := &BeaconBlockBody{
		version:           version.Bellatrix,
		isBlinded:         false,
		randaoReveal:      bytesutil.ToBytes96(pb.RandaoReveal),
		eth1Data:          pb.Eth1Data,
		graffiti:          bytesutil.ToBytes32(pb.Graffiti),
		proposerSlashings: pb.ProposerSlashings,
		attesterSlashings: pb.AttesterSlashings,
		attestations:      pb.Attestations,
		deposits:          pb.Deposits,
		voluntaryExits:    pb.VoluntaryExits,
		syncAggregate:     pb.SyncAggregate,
		executionData:     executionData,
	}
	return b, nil
}

func initBlindedBlockBodyFromProtoBellatrix(pb *eth.BlindedBeaconBlockBodyBellatrix) (*BeaconBlockBody, error) {
	if pb == nil {
		return nil, errNilBlockBody
	}

	execHeader, err := NewExecutionDataHeader(pb.ExecutionPayloadHeader)
	if err != nil {
		return nil, err
	}

	b := &BeaconBlockBody{
		version:             version.Bellatrix,
		isBlinded:           true,
		randaoReveal:        bytesutil.ToBytes96(pb.RandaoReveal),
		eth1Data:            pb.Eth1Data,
		graffiti:            bytesutil.ToBytes32(pb.Graffiti),
		proposerSlashings:   pb.ProposerSlashings,
		attesterSlashings:   pb.AttesterSlashings,
		attestations:        pb.Attestations,
		deposits:            pb.Deposits,
		voluntaryExits:      pb.VoluntaryExits,
		syncAggregate:       pb.SyncAggregate,
		executionDataHeader: execHeader,
	}
	return b, nil
}

func initBlockBodyFromProtoEip4844(pb *eth.BeaconBlockBodyWithBlobKZGs) (*BeaconBlockBody, error) {
	if pb == nil {
		return nil, errNilBlockBody
	}

	executionData, err := NewExecutionData(pb.ExecutionPayload)
	if err != nil {
		return nil, err
	}

	b := &BeaconBlockBody{
		version:           version.EIP4844,
		isBlinded:         false,
		randaoReveal:      bytesutil.ToBytes96(pb.RandaoReveal),
		eth1Data:          pb.Eth1Data,
		graffiti:          bytesutil.ToBytes32(pb.Graffiti),
		proposerSlashings: pb.ProposerSlashings,
		attesterSlashings: pb.AttesterSlashings,
		attestations:      pb.Attestations,
		deposits:          pb.Deposits,
		voluntaryExits:    pb.VoluntaryExits,
		syncAggregate:     pb.SyncAggregate,
		executionData:     executionData,
		blobKzgs:          pb.BlobKzgs,
	}
	return b, nil
}

func initPayloadFromProto(pb *enginev1.ExecutionPayload) (*executionPayload, error) {
	if pb == nil {
		return nil, errNilPayload
	}

	e := &executionPayload{
		version:       version.Bellatrix,
		parentHash:    pb.ParentHash,
		feeRecipient:  pb.FeeRecipient,
		stateRoot:     pb.StateRoot,
		receiptsRoot:  pb.ReceiptsRoot,
		logsBloom:     pb.LogsBloom,
		prevRandao:    pb.PrevRandao,
		blockNumber:   pb.BlockNumber,
		gasLimit:      pb.GasLimit,
		gasUsed:       pb.GasUsed,
		timestamp:     pb.Timestamp,
		extraData:     pb.ExtraData,
		baseFeePerGas: pb.BaseFeePerGas,
		blockHash:     pb.BlockHash,
		transactions:  pb.Transactions,
	}
	return e, nil
}

func initPayloadFromProto4844(pb *enginev1.ExecutionPayload4844) (*executionPayload, error) {
	if pb == nil {
		return nil, errNilPayload
	}

	e := &executionPayload{
		version:       version.EIP4844,
		parentHash:    pb.ParentHash,
		feeRecipient:  pb.FeeRecipient,
		stateRoot:     pb.StateRoot,
		receiptsRoot:  pb.ReceiptsRoot,
		logsBloom:     pb.LogsBloom,
		prevRandao:    pb.PrevRandao,
		blockNumber:   pb.BlockNumber,
		gasLimit:      pb.GasLimit,
		gasUsed:       pb.GasUsed,
		timestamp:     pb.Timestamp,
		extraData:     pb.ExtraData,
		baseFeePerGas: pb.BaseFeePerGas,
		blockHash:     pb.BlockHash,
		transactions:  pb.Transactions,
		excessDataGas: pb.ExcessDataGas,
	}
	return e, nil
}

func initPayloadHeaderFromProto(pb *enginev1.ExecutionPayloadHeader) (*executionPayloadHeader, error) {
	if pb == nil {
		return nil, errNilPayload
	}

	e := &executionPayloadHeader{
		version:          version.Bellatrix,
		parentHash:       pb.ParentHash,
		feeRecipient:     pb.FeeRecipient,
		stateRoot:        pb.StateRoot,
		receiptsRoot:     pb.ReceiptsRoot,
		logsBloom:        pb.LogsBloom,
		prevRandao:       pb.PrevRandao,
		blockNumber:      pb.BlockNumber,
		gasLimit:         pb.GasLimit,
		gasUsed:          pb.GasUsed,
		timestamp:        pb.Timestamp,
		extraData:        pb.ExtraData,
		baseFeePerGas:    pb.BaseFeePerGas,
		blockHash:        pb.BlockHash,
		transactionsRoot: pb.TransactionsRoot,
	}
	return e, nil
}

func initPayloadHeaderFromProto4844(pb *enginev1.ExecutionPayloadHeader4844) (*executionPayloadHeader, error) {
	if pb == nil {
		return nil, errNilPayload
	}

	e := &executionPayloadHeader{
		version:          version.EIP4844,
		parentHash:       pb.ParentHash,
		feeRecipient:     pb.FeeRecipient,
		stateRoot:        pb.StateRoot,
		receiptsRoot:     pb.ReceiptsRoot,
		logsBloom:        pb.LogsBloom,
		prevRandao:       pb.PrevRandao,
		blockNumber:      pb.BlockNumber,
		gasLimit:         pb.GasLimit,
		gasUsed:          pb.GasUsed,
		timestamp:        pb.Timestamp,
		extraData:        pb.ExtraData,
		baseFeePerGas:    pb.BaseFeePerGas,
		blockHash:        pb.BlockHash,
		transactionsRoot: pb.TransactionsRoot,
		excessDataGas:    pb.ExcessDataGas,
	}
	return e, nil
}
