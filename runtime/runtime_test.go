package main

import (
	"bytes"
	"testing"

	gossamertypes "github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/lib/common"
	runtimetypes "github.com/ChainSafe/gossamer/lib/runtime"
	"github.com/ChainSafe/gossamer/lib/runtime/wasmer"
	"github.com/ChainSafe/gossamer/lib/trie"
	"github.com/ChainSafe/gossamer/pkg/scale"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	keySystemHash, _         = common.Twox128Hash(constants.KeySystem)
	keyBlockHash, _          = common.Twox128Hash(constants.KeyBlockHash)
	keyDigestHash, _         = common.Twox128Hash(constants.KeyDigest)
	keyExecutionPhaseHash, _ = common.Twox128Hash(constants.KeyExecutionPhase)
	keyLastRuntime, _        = common.Twox128Hash(constants.KeyLastRuntimeUpgrade)
	keyNumberHash, _         = common.Twox128Hash(constants.KeyNumber)
	keyParentHash, _         = common.Twox128Hash(constants.KeyParentHash)
)

const WASM_RUNTIME = "../build/runtime.wasm"

func Test_CoreVersion(t *testing.T) {
	storage := trie.NewEmptyTrie()
	rt := wasmer.NewTestInstanceWithTrie(t, WASM_RUNTIME, storage)

	res, err := rt.Exec("Core_version", []byte{})
	assert.NoError(t, err)

	buffer := bytes.Buffer{}
	buffer.Write(res)
	dec := scale.NewDecoder(&buffer)
	runtimeVersion := runtimetypes.Version{}
	err = dec.Decode(&runtimeVersion)
	assert.NoError(t, err)
	assert.Equal(t, "node-template", string(runtimeVersion.SpecName))
	assert.Equal(t, "node-template", string(runtimeVersion.ImplName))
	assert.Equal(t, uint32(1), runtimeVersion.AuthoringVersion)
	assert.Equal(t, uint32(100), runtimeVersion.SpecVersion)
	assert.Equal(t, uint32(1), runtimeVersion.ImplVersion)
	assert.Equal(t, uint32(1), runtimeVersion.TransactionVersion)
	assert.Equal(t, uint32(1), runtimeVersion.StateVersion)
}

func Test_CoreInitializeBlock(t *testing.T) {
	parentHash := common.MustHexToHash("0x0f6d3477739f8a65886135f58c83ff7c2d4a8300a010dfc8b4c5d65ba37920bb")
	stateRoot := common.MustHexToHash("0x211fc45bbc8f57af1a5d01a689788024be5a1738b51e3fbae13494f1e9e318da")
	extrinsicsRoot := common.MustHexToHash("0x5e3ab240467545190bae81d181914f16a03cbfc23a809cc74764afc00b5a014f")
	blockNumber := uint(1)
	expectedStorageDigest := gossamertypes.NewDigest()

	digest := gossamertypes.NewDigest()

	sealDigest := gossamertypes.SealDigest{
		ConsensusEngineID: gossamertypes.BabeEngineID,
		// bytes for SealDigest that was created in setupHeaderFile function
		Data: []byte{158, 127, 40, 221, 220, 242, 124, 30, 107, 50, 141, 86, 148, 195, 104, 213, 178, 236, 93, 190,
			14, 65, 42, 225, 201, 143, 136, 213, 59, 228, 216, 80, 47, 172, 87, 31, 63, 25, 201, 202, 175, 40, 26,
			103, 51, 25, 36, 30, 12, 80, 149, 166, 131, 173, 52, 49, 98, 4, 8, 138, 54, 164, 189, 134},
	}

	preRuntimeDigest := gossamertypes.PreRuntimeDigest{
		ConsensusEngineID: gossamertypes.BabeEngineID,
		// bytes for PreRuntimeDigest that was created in setupHeaderFile function
		Data: []byte{1, 60, 0, 0, 0, 150, 89, 189, 15, 0, 0, 0, 0, 112, 237, 173, 28, 144, 100, 255,
			247, 140, 177, 132, 53, 34, 61, 138, 218, 245, 234, 4, 194, 75, 26, 135, 102, 227, 220, 1, 235, 3, 204,
			106, 12, 17, 183, 151, 147, 212, 227, 28, 192, 153, 8, 56, 34, 156, 68, 254, 209, 102, 154, 124, 124,
			121, 225, 230, 208, 169, 99, 116, 214, 73, 103, 40, 6, 157, 30, 247, 57, 226, 144, 73, 122, 14, 59, 114,
			143, 168, 143, 203, 221, 58, 85, 4, 224, 239, 222, 2, 66, 231, 168, 6, 221, 79, 169, 38, 12},
	}

	preRuntimeDigestItem := gossamertypes.NewDigestItem()
	assert.NoError(t, preRuntimeDigestItem.Set(preRuntimeDigest))

	sealDigestItem := gossamertypes.NewDigestItem()
	assert.NoError(t, sealDigestItem.Set(sealDigest))

	prdi, err := preRuntimeDigestItem.Value()
	assert.NoError(t, err)
	assert.NoError(t, digest.Add(prdi))

	sdi, err := sealDigestItem.Value()
	assert.NoError(t, err)
	assert.NoError(t, digest.Add(sdi))
	assert.NoError(t, expectedStorageDigest.Add(prdi))

	header := gossamertypes.NewHeader(parentHash, stateRoot, extrinsicsRoot, blockNumber, digest)
	encodedHeader, err := scale.Marshal(*header)
	assert.NoError(t, err)

	storage := trie.NewEmptyTrie()
	rt := wasmer.NewTestInstanceWithTrie(t, WASM_RUNTIME, storage)

	_, err = rt.Exec("Core_initialize_block", encodedHeader)
	assert.Nil(t, err)

	lrui := types.LastRuntimeUpgradeInfo{
		SpecVersion: sc.ToCompact(constants.SPEC_VERSION),
		SpecName:    constants.SPEC_NAME,
	}
	assert.Equal(t, lrui.Bytes(), storage.Get(append(keySystemHash, keyLastRuntime...)))

	encExtrinsicIndex0, _ := scale.Marshal(uint32(0))
	assert.Equal(t, encExtrinsicIndex0, storage.Get(constants.KeyExtrinsicIndex))

	encExecutionPhaseApplyExtrinsic, _ := scale.Marshal(uint32(0))
	assert.Equal(t, encExecutionPhaseApplyExtrinsic, storage.Get(append(keySystemHash, keyExecutionPhaseHash...)))

	encBlockNumber, _ := scale.Marshal(uint32(blockNumber))
	assert.Equal(t, encBlockNumber, storage.Get(append(keySystemHash, keyNumberHash...)))

	encExpectedDigest, err := scale.Marshal(expectedStorageDigest)
	assert.NoError(t, err)
	assert.Equal(t, encExpectedDigest, storage.Get(append(keySystemHash, keyDigestHash...)))
	assert.Equal(t, parentHash.ToBytes(), storage.Get(append(keySystemHash, keyParentHash...)))

	blockHashKey := append(keySystemHash, keyBlockHash...)
	encPrevBlock, _ := scale.Marshal(uint32(blockNumber - 1))
	numHash, err := common.Twox64(encPrevBlock)
	assert.NoError(t, err)

	blockHashKey = append(blockHashKey, numHash...)
	blockHashKey = append(blockHashKey, encPrevBlock...)
	assert.Equal(t, parentHash.ToBytes(), storage.Get(blockHashKey))
}
