// todo unit act
package params

// UseOceanNetworkConfig uses the Ocean specific
// network config.
func UseOceanNetworkConfig() {
	cfg := BeaconNetworkConfig().Copy()
	cfg.ContractDeploymentBlock = 120777
	cfg.BootstrapNodes = []string{
		// FastexChain Ocean bootnode.
		"enr:-MK4QBX_5pdqifsjrq5cMqTtCeGdwy5vhi1KcrEGYHFUoMtVfV1o-fGoAK6AqOgEdejuE3Dhpcan9LH7v2x06SZPWqSGAY8qJh6Ah2F0dG5ldHOIAAAAAAAAAACEZXRoMpBQkMu0AwAoOmkKAAAAAAAAgmlkgnY0gmlwhCKMzwKJc2VjcDI1NmsxoQJebC3ha7lpBWhQuCENG-5oksf1wwmpu8m7D8Be4VGqQYhzeW5jbmV0cwCDdGNwgjLIg3VkcIIu4A",
		"enr:-MK4QPvs4Y6oLLMgyctu0PVeJLQruit-Hme4CLDndxBtBzgxfAGxq_VvZoeP2DqFVdsFLR0VbUjtI4ymp5J6vNLacEuGAY8qJKeUh2F0dG5ldHOIAAAAAAAAAACEZXRoMpBQkMu0AwAoOmkKAAAAAAAAgmlkgnY0gmlwhCO7RCGJc2VjcDI1NmsxoQMnkSDNngSNs-ZTlNIx1ezc1RLWU49KyyQXzN5tQL7DxYhzeW5jbmV0cwCDdGNwgjLIg3VkcIIu4A",
	}
	OverrideBeaconNetworkConfig(cfg)
}

// OceanConfig defines the config for the
// Ocean testnet.
func OceanConfig() *BeaconChainConfig {
	cfg := MainnetConfig().Copy()
	cfg.MinGenesisTime = 1688976000
	cfg.GenesisDelay = 86400 // 24 hours
	cfg.ConfigName = OceanName
	cfg.SecondsPerETH1Block = 12
	cfg.DepositChainID = 4058
	cfg.DepositNetworkID = 4058

	cfg.GenesisForkVersion = []byte{0, 0, 40, 58}
	cfg.AltairForkEpoch = 1
	cfg.AltairForkVersion = []byte{1, 0, 40, 58}
	cfg.BellatrixForkEpoch = 2
	cfg.BellatrixForkVersion = []byte{2, 0, 40, 58}
	cfg.CapellaForkEpoch = 2665
	cfg.CapellaForkVersion = []byte{3, 0, 40, 58}
	cfg.DenebForkEpoch = cfg.FarFutureEpoch
	cfg.DenebForkVersion = []byte{4, 0, 40, 58}
	cfg.TerminalTotalDifficulty = "514667"
	cfg.DepositContractAddress = "0xCC0F57056771a04fB56436F6A565De98dD6B1915"
	cfg.InitializeForkSchedule()
	cfg.FirstFixFork = false

	return cfg
}
