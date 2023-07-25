package params

// UseOasisNetworkConfig uses the Oasis specific
// network config.
func UseOasisNetworkConfig() {
	cfg := BeaconNetworkConfig().Copy()
	cfg.ContractDeploymentBlock = 502405
	cfg.BootstrapNodes = []string{
		// FastexChain Oasis bootnode.
		"enr:-LG4QI1-Qu7pmmPwxCiaYCxtbuIQzS0GGZAtVpRcobNMS_fzDq4GTQokN721cZGFsVR70HKswJmKS1m1sZbB5AVCF9eGAYl4tg01h2F0dG5ldHOIAAAAAAAAAACEZXRoMpCJnEJeAAAoWv__________gmlkgnY0gmlwhCPDPg2Jc2VjcDI1NmsxoQP15vbtLz7X3iiyLtt2FmV1IWMtrvglrSW8VpIEdabk14N1ZHCCIyg",
	}
	OverrideBeaconNetworkConfig(cfg)
}

// OasisConfig defines the config for the
// Oasis testnet.
func OasisConfig() *BeaconChainConfig {
	cfg := MainnetConfig().Copy()
	cfg.MinGenesisTime = 1688976000
	cfg.GenesisDelay = 176400 // 49 hours
	cfg.ConfigName = OasisName
	cfg.SecondsPerETH1Block = 12
	cfg.DepositChainID = 4090
	cfg.DepositNetworkID = 4090

	cfg.GenesisForkVersion = []byte{0, 0, 40, 90}
	cfg.AltairForkEpoch = 1
	cfg.AltairForkVersion = []byte{1, 0, 40, 90}
	cfg.BellatrixForkEpoch = 2
	cfg.BellatrixForkVersion = []byte{2, 0, 40, 90}
	cfg.CapellaForkEpoch = cfg.FarFutureEpoch
	cfg.CapellaForkVersion = []byte{3, 0, 40, 90}
	cfg.TerminalTotalDifficulty = "1205400" // ~ 2023-07-25 08:03:10 UTC
	cfg.DepositContractAddress = "0x84e09dbd827b66bfa66f42fc50dc732c01790122"
	cfg.InitializeForkSchedule()
	return cfg
}
