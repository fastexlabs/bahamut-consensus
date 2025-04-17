package params

// UseHorizonNetworkConfig uses the Horizon specific
// network config.
func UseHorizonNetworkConfig() {
	cfg := BeaconNetworkConfig().Copy()
	cfg.ContractDeploymentBlock = 27102
	cfg.BootstrapNodes = []string{
		// FastexChain Horizon bootnode.
		"enr:-MK4QIJzui9U2bJuDlP5ElU1fKtFyVwQOEssjhRPhxUcnsVdQ93wg1BMDVADeQeM-ib9pCZ0SQ0kvPKKzhY__7qBoiSGAZLhwdXah2F0dG5ldHOIAAAAAAAAAACEZXRoMpC2vsBrAwAZNB4FAAAAAAAAgmlkgnY0gmlwhCKMzwKJc2VjcDI1NmsxoQJFVUyprT2S-I7oQz1-D1527qsOhKdQeQ11hW1Zjv-bq4hzeW5jbmV0cwCDdGNwgjLIg3VkcIIu4A",
		"enr:-MK4QAMSjfDKuthUHJmpp564p3WohbnHmPvcFdiTQWrdOR2JSQ4z3UCSc5BAX3lOzAqRhflRxqx9y1fm9vPrBuNswdSGAZLiHzQ7h2F0dG5ldHOIAAAAAAAAAACEZXRoMpC2vsBrAwAZNB4FAAAAAAAAgmlkgnY0gmlwhCO7RCGJc2VjcDI1NmsxoQL5T7ELcZjc0-abeyEKGVXpHweLvJZA7Et4Umw1BmudQohzeW5jbmV0cwCDdGNwgjLIg3VkcIIu4A",
		"enr:-MK4QGRYcKyxsceEtvDox51vCbsrhhXJvH1R1WTAIRWAA02IF2XtPwu16TT8ae4K1yfLp_N4JL8uKTZdZr6AWlaNQheGAZLiGbpTh2F0dG5ldHOIAAAAAAAAAACEZXRoMpC2vsBrAwAZNP__________gmlkgnY0gmlwhCPDPg2Jc2VjcDI1NmsxoQJn_Bra04fNEfHVfsyhpWTU_U7Koi3FxmposEcKeh56O4hzeW5jbmV0cwCDdGNwgjLIg3VkcIIu4A",
	}
	OverrideBeaconNetworkConfig(cfg)
}

// HorizonConfig defines the config for the
// Horizon testnet.
func HorizonConfig() *BeaconChainConfig {
	cfg := MainnetConfig().Copy()
	cfg.MinGenesisTime = 1730104778
	cfg.GenesisDelay = 172800 // 24 hours
	cfg.ConfigName = HorizonName
	cfg.SecondsPerETH1Block = 12
	cfg.DepositChainID = 2552
	cfg.DepositNetworkID = 2552
	cfg.MinGenesisActiveValidatorCount = 4104

	cfg.GenesisForkVersion = []byte{0, 0, 25, 52}
	cfg.AltairForkEpoch = 1
	cfg.AltairForkVersion = []byte{1, 0, 25, 52}
	cfg.BellatrixForkEpoch = 2
	cfg.BellatrixForkVersion = []byte{2, 0, 25, 52}
	cfg.CapellaForkEpoch = 1310
	cfg.CapellaForkVersion = []byte{3, 0, 25, 52}
	cfg.DenebForkEpoch = cfg.FarFutureEpoch
	cfg.DenebForkVersion = []byte{4, 0, 25, 52}
	cfg.TerminalTotalDifficulty = "167079"
	cfg.DepositContractAddress = "0xc235be402811955ab4e9a9753e13d103781740f7"
	cfg.InitializeForkSchedule()
	cfg.FirstFixFork = false

	return cfg
}
