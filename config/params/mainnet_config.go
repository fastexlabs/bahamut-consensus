package params

import (
	"math"
	"time"

	fieldparams "github.com/prysmaticlabs/prysm/v4/config/fieldparams"
	"github.com/prysmaticlabs/prysm/v4/encoding/bytesutil"
)

// MainnetConfig returns the configuration to be used in the main network.
func MainnetConfig() *BeaconChainConfig {
	if mainnetBeaconConfig.ForkVersionSchedule == nil {
		mainnetBeaconConfig.InitializeForkSchedule()
	}
	return mainnetBeaconConfig
}

const (
	// Genesis Fork Epoch for the mainnet config.
	genesisForkEpoch = 0
	// Altair Fork Epoch for mainnet config.
	mainnetAltairForkEpoch = 1
	// Bellatrix Fork Epoch for mainnet config.
	mainnetBellatrixForkEpoch = 2
	// Capella Fork Epoch for mainnet config
	mainnetCapellaForkEpoch = 130731
)

var mainnetNetworkConfig = &NetworkConfig{
	GossipMaxSize:                    1 << 20,      // 1 MiB
	GossipMaxSizeBellatrix:           10 * 1 << 20, // 10 MiB
	MaxChunkSize:                     1 << 20,      // 1 MiB
	MaxChunkSizeBellatrix:            10 * 1 << 20, // 10 MiB
	AttestationSubnetCount:           64,
	AttestationPropagationSlotRange:  32,
	MaxRequestBlocks:                 1 << 10, // 1024
	TtfbTimeout:                      5 * time.Second,
	RespTimeout:                      10 * time.Second,
	MaximumGossipClockDisparity:      500 * time.Millisecond,
	MessageDomainInvalidSnappy:       [4]byte{0o0, 0o0, 0o0, 0o0},
	MessageDomainValidSnappy:         [4]byte{0o1, 0o0, 0o0, 0o0},
	ETH2Key:                          "eth2",
	AttSubnetKey:                     "attnets",
	SyncCommsSubnetKey:               "syncnets",
	MinimumPeersInSubnetSearch:       20,
	ContractDeploymentBlock:          646081,
	MinEpochsForBlobsSidecarsRequest: 4096,
	MaxRequestBlobSidecars:           768,
	MaxRequestBlocksDeneb:            128,
	// TODO(fastex-chain): Set fastex-chain bootnodes.
	BootstrapNodes: []string{
		"enr:-MK4QM_tgHqJXM7U3bIImq9OLLKgc6ytdUwgW5BgSdu05dOcYGy_vgDLAqI6K1ld-ojodeU5qlOVwSlICamMbZs_S5aGAY27q_b1h2F0dG5ldHOIAAAAAAAAAACEZXRoMpBy8gcXAwAzQf__________gmlkgnY0gmlwhCPDK-6Jc2VjcDI1NmsxoQO4tL0Fd1b2_Lt8XdeuOQ0JqDhcLZY4OrM7mMhi1HL114hzeW5jbmV0cwCDdGNwgjLIg3VkcIIu4A",
		"enr:-MK4QPDEFngFojaTqTIMZalEA1KLyn78OtSWy4JtN8jtZcsgc3_c46RN-y21otZPoLJX7MSvBIwDAJA5qMMkG3LJt0CGAY0hRsd9h2F0dG5ldHOIAAAAAAAAAACEZXRoMpBy8gcXAwAzQf__________gmlkgnY0gmlwhCJgr_CJc2VjcDI1NmsxoQNrRvtg8beyk9PPbTj29ytu_noBeIxNbHndKH2mlIjsDYhzeW5jbmV0cwCDdGNwgjLIg3VkcIIu4A",
		"enr:-MK4QAvVbpe6-dmbUHPkGFZMTV6mlQyp6OFssLmnQnds81HXbjZtG6dUoHURdxpARMyEyVsSS69nBYV-tJIPEHp7rNSGAYuzLsmch2F0dG5ldHOIAAAAAAAAAACEZXRoMpBy8gcXAwAzQf__________gmlkgnY0gmlwhCPuB62Jc2VjcDI1NmsxoQLsT2iocGu1dTs-kkbF2v67tyHnJmCMb0NWEKXUs1aWEIhzeW5jbmV0cwCDdGNwgjLIg3VkcIIu4A",
	},
}

var mainnetBeaconConfig = &BeaconChainConfig{
	// Constants (Non-configurable)
	FarFutureEpoch:           math.MaxUint64,
	FarFutureSlot:            math.MaxUint64,
	BaseRewardsPerEpoch:      4,
	DepositContractTreeDepth: 32,
	GenesisDelay:             345600,

	// Misc constant.
	TargetCommitteeSize:            128,
	MaxValidatorsPerCommittee:      2048,
	MaxCommitteesPerSlot:           64,
	MinPerEpochChurnLimit:          4,
	ChurnLimitQuotient:             1 << 16,
	ShuffleRoundCount:              90,
	MinGenesisActiveValidatorCount: 4096,
	MinGenesisTime:                 1690876800, // Jul 12, 2023, 8am UTC.
	TargetAggregatorsPerCommittee:  16,
	HysteresisQuotient:             4,
	HysteresisDownwardMultiplier:   1,
	HysteresisUpwardMultiplier:     5,

	// Gwei value constants.
	MinDepositAmount:          256 * 1e9,
	MaxEffectiveBalance:       8192 * 1e9,
	EjectionBalance:           4096 * 1e9,
	EffectiveBalanceIncrement: 1 * 1e9,

	// Initial value constants.
	BLSWithdrawalPrefixByte:         byte(0),
	ETH1AddressWithdrawalPrefixByte: byte(1),
	ZeroHash:                        [32]byte{},

	// Time parameter constants.
	MinAttestationInclusionDelay:     1,
	SecondsPerSlot:                   12,
	SlotsPerEpoch:                    32,
	SqrRootSlotsPerEpoch:             5,
	MinSeedLookahead:                 1,
	MaxSeedLookahead:                 4,
	EpochsPerEth1VotingPeriod:        64,
	SlotsPerHistoricalRoot:           8192,
	MinValidatorWithdrawabilityDelay: 256,
	ShardCommitteePeriod:             256,
	MinEpochsToInactivityPenalty:     4,
	Eth1FollowDistance:               2048,

	// Fork choice algorithm constants.
	ProposerScoreBoost:              40,
	ReorgWeightThreshold:            20,
	ReorgParentWeightThreshold:      160,
	ReorgMaxEpochsSinceFinalization: 2,
	IntervalsPerSlot:                3,

	// Ethereum PoW parameters.
	DepositChainID:         5165,                                         // Chain ID of eth1 mainnet.
	DepositNetworkID:       5165,                                         // Network ID of eth1 mainnet.
	DepositContractAddress: "0x385C32d00cD8FF896eCd7Ca3335bb30f391a3057", // TODO(fastex): set Fasttoken deposit contract address

	// Validator params.
	RandomSubnetsPerValidator:         1 << 0,
	EpochsPerRandomSubnetSubscription: 1 << 8,

	// While eth1 mainnet block times are closer to 13s, we must conform with other clients in
	// order to vote on the correct eth1 blocks.
	//
	// Additional context: https://github.com/ethereum/consensus-specs/issues/2132
	// Bug prompting this change: https://github.com/prysmaticlabs/prysm/issues/7856
	// Future optimization: https://github.com/prysmaticlabs/prysm/issues/7739
	SecondsPerETH1Block: 12,

	// State list length constants.
	EpochsPerHistoricalVector: 65536,
	EpochsPerSlashingsVector:  8192,
	HistoricalRootsLimit:      16777216,
	ValidatorRegistryLimit:    1099511627776,

	// Reward and penalty quotients constants.
	BaseRewardFactor:               156,
	WhistleBlowerRewardQuotient:    512,
	ProposerRewardQuotient:         8,
	InactivityPenaltyQuotient:      67108864,
	MinSlashingPenaltyQuotient:     128,
	ProportionalSlashingMultiplier: 1,

	// Max operations per block constants.
	MaxProposerSlashings:             16,
	MaxAttesterSlashings:             2,
	MaxAttestations:                  128,
	MaxDeposits:                      16,
	MaxVoluntaryExits:                16,
	MaxWithdrawalsPerPayload:         16,
	MaxBlsToExecutionChanges:         16,
	MaxValidatorsPerWithdrawalsSweep: 16384,

	// BLS domain values.
	DomainBeaconProposer:              bytesutil.Uint32ToBytes4(0x00000000),
	DomainBeaconAttester:              bytesutil.Uint32ToBytes4(0x01000000),
	DomainRandao:                      bytesutil.Uint32ToBytes4(0x02000000),
	DomainDeposit:                     bytesutil.Uint32ToBytes4(0x03000000),
	DomainVoluntaryExit:               bytesutil.Uint32ToBytes4(0x04000000),
	DomainSelectionProof:              bytesutil.Uint32ToBytes4(0x05000000),
	DomainAggregateAndProof:           bytesutil.Uint32ToBytes4(0x06000000),
	DomainSyncCommittee:               bytesutil.Uint32ToBytes4(0x07000000),
	DomainSyncCommitteeSelectionProof: bytesutil.Uint32ToBytes4(0x08000000),
	DomainContributionAndProof:        bytesutil.Uint32ToBytes4(0x09000000),
	DomainApplicationMask:             bytesutil.Uint32ToBytes4(0x00000001),
	DomainApplicationBuilder:          bytesutil.Uint32ToBytes4(0x00000001),
	DomainBLSToExecutionChange:        bytesutil.Uint32ToBytes4(0x0A000000),
	DomainBlobSidecar:                 bytesutil.Uint32ToBytes4(0x0B000000),

	// FastexChain consensus constants.
	EpochsPerActivityPeriod: 1575, // One week (12s * 32 * 1575)

	// Prysm constants.
	GweiPerEth:                     1000000000,
	WeiPerGwei:                     1000000000,
	BaseTransactionCost:            21000,
	BLSSecretKeyLength:             32,
	BLSPubkeyLength:                48,
	DefaultBufferSize:              10000,
	WithdrawalPrivkeyFileName:      "/shardwithdrawalkey",
	ValidatorPrivkeyFileName:       "/validatorprivatekey",
	RPCSyncCheck:                   1,
	EmptySignature:                 [96]byte{},
	DefaultPageSize:                250,
	MaxPeersToSync:                 15,
	SlotsPerArchivedPoint:          2048,
	GenesisCountdownInterval:       time.Minute,
	ConfigName:                     MainnetName,
	PresetBase:                     "mainnet",
	BeaconStateFieldCount:          23,
	BeaconStateAltairFieldCount:    26,
	BeaconStateBellatrixFieldCount: 27,
	BeaconStateCapellaFieldCount:   30,
	BeaconStateDenebFieldCount:     30,

	// Slasher related values.
	WeakSubjectivityPeriod:          54000,
	PruneSlasherStoragePeriod:       10,
	SlashingProtectionPruningEpochs: 512,

	// Weak subjectivity values.
	SafetyDecay: 10,

	// Fork related values.
	GenesisEpoch:         genesisForkEpoch,
	GenesisForkVersion:   []byte{0, 0, 51, 65},
	AltairForkVersion:    []byte{1, 0, 51, 65},
	AltairForkEpoch:      mainnetAltairForkEpoch,
	BellatrixForkVersion: []byte{2, 0, 51, 65},
	BellatrixForkEpoch:   mainnetBellatrixForkEpoch,
	CapellaForkVersion:   []byte{3, 0, 51, 65},
	CapellaForkEpoch:     mainnetCapellaForkEpoch,
	DenebForkVersion:     []byte{4, 0, 51, 65},
	DenebForkEpoch:       math.MaxUint64,

	// New values introduced in Altair hard fork 1.
	// Participation flag indices.
	TimelySourceFlagIndex: 0,
	TimelyTargetFlagIndex: 1,
	TimelyHeadFlagIndex:   2,

	// Incentivization weight values.
	TimelySourceWeight: 14,
	TimelyTargetWeight: 26,
	TimelyHeadWeight:   14,
	SyncRewardWeight:   2,
	ProposerWeight:     0, // Is not used by Proof-Of-Stake-and-Activity algorithm.
	WeightDenominator:  56,

	// Validator related values.
	TargetAggregatorsPerSyncSubcommittee: 16,
	SyncCommitteeSubnetCount:             4,

	// Misc values.
	SyncCommitteeSize:            512,
	InactivityScoreBias:          4,
	InactivityScoreRecoveryRate:  16,
	EpochsPerSyncCommitteePeriod: 256,

	// Updated penalty values.
	InactivityPenaltyQuotientAltair:         3 * 1 << 24, //50331648
	MinSlashingPenaltyQuotientAltair:        64,
	ProportionalSlashingMultiplierAltair:    2,
	MinSlashingPenaltyQuotientBellatrix:     32,
	ProportionalSlashingMultiplierBellatrix: 3,
	InactivityPenaltyQuotientBellatrix:      1 << 24,

	// Light client
	MinSyncCommitteeParticipants: 1,

	// Bellatrix
	TerminalBlockHashActivationEpoch: 18446744073709551615,
	TerminalBlockHash:                [32]byte{},
	TerminalTotalDifficulty:          "10087646", // ~ at 2025-04-08 15:00:00 GTM+4 in block 5116659
	EthBurnAddressHex:                "0x0000000000000000000000000000000000000000",
	DefaultBuilderGasLimit:           uint64(30000000),

	// Mevboost circuit breaker
	MaxBuilderConsecutiveMissedSlots: 3,
	MaxBuilderEpochMissedSlots:       5,
	// Execution engine timeout value
	ExecutionEngineTimeoutValue: 8, // 8 seconds default based on: https://github.com/ethereum/execution-apis/blob/main/src/engine/specification.md#core

	// Subnet value
	BlobsidecarSubnetCount: 6,

	MaxPerEpochActivationChurnLimit: 8,
	//todo unit act
	FirstFixFork: true,
}

// MainnetTestConfig provides a version of the mainnet config that has a different name
// and a different fork choice schedule. This can be used in cases where we want to use config values
// that are consistent with mainnet, but won't conflict or cause the hard-coded genesis to be loaded.
func MainnetTestConfig() *BeaconChainConfig {
	mn := MainnetConfig().Copy()
	mn.ConfigName = MainnetTestName
	FillTestVersions(mn, 128)
	return mn
}

// FillTestVersions replaces the fork schedule in the given BeaconChainConfig with test values, using the given
// byte argument as the high byte (common across forks).
func FillTestVersions(c *BeaconChainConfig, b byte) {
	c.GenesisForkVersion = make([]byte, fieldparams.VersionLength)
	c.AltairForkVersion = make([]byte, fieldparams.VersionLength)
	c.BellatrixForkVersion = make([]byte, fieldparams.VersionLength)
	c.CapellaForkVersion = make([]byte, fieldparams.VersionLength)
	c.DenebForkVersion = make([]byte, fieldparams.VersionLength)

	c.GenesisForkVersion[fieldparams.VersionLength-1] = b
	c.AltairForkVersion[fieldparams.VersionLength-1] = b
	c.BellatrixForkVersion[fieldparams.VersionLength-1] = b
	c.CapellaForkVersion[fieldparams.VersionLength-1] = b
	c.DenebForkVersion[fieldparams.VersionLength-1] = b

	c.GenesisForkVersion[0] = 0
	c.AltairForkVersion[0] = 1
	c.BellatrixForkVersion[0] = 2
	c.CapellaForkVersion[0] = 3
	c.DenebForkVersion[0] = 4
}
