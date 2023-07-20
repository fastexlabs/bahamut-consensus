package params

// UseOasisNetworkConfig uses the Oasis specific
// network config.
func UseOasisNetworkConfig() {
	cfg := BeaconNetworkConfig().Copy()
	cfg.ContractDeploymentBlock = 502405
	cfg.BootstrapNodes = []string{
		// FastexChain Oasis bootnode.
		"enr:-LG4QKpj0Ue_SnjxStp-60GaXdkbzU2rtZPwR_gMehWBqrGND3JRofqK7eNg99p7AXlQbP1dCYt7CtduvMwvn4Ee8cmGAYlEg3M5h2F0dG5ldHOIAAAAAAAAAACEZXRoMpCJnEJeAAAoWv__________gmlkgnY0gmlwhCPDPg2Jc2VjcDI1NmsxoQP15vbtLz7X3iiyLtt2FmV1IWMtrvglrSW8VpIEdabk14N1ZHCCIyg",
	}
	OverrideBeaconNetworkConfig(cfg)
}

// OasisConfig defines the config for the
// Oasis testnet.
func OasisConfig() *BeaconChainConfig {
	cfg := MainnetConfig().Copy()
	cfg.MinGenesisTime = 1688976000
	cfg.GenesisDelay = 86400
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
	cfg.TerminalTotalDifficulty = "1191000" // ~ 2023-07-24 08:03:10 UTC
	cfg.DepositContractAddress = "0x84e09dbd827b66bfa66f42fc50dc732c01790122"
	cfg.InitializeForkSchedule()
	return cfg
}
