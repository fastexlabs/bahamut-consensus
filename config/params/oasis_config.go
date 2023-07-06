package params

// UseOasisNetworkConfig uses the Oasis specific
// network config.
func UseOasisNetworkConfig() {
	cfg := BeaconNetworkConfig().Copy()
	// TODO(fastex): set deposit contract deployment block and bootnodes.
	cfg.ContractDeploymentBlock = 4367322
	cfg.BootstrapNodes = []string{
		// Prysm's bootnode
		"enr:-Ku4QFmUkNp0g9bsLX2PfVeIyT-9WO-PZlrqZBNtEyofOOfLMScDjaTzGxIb1Ns9Wo5Pm_8nlq-SZwcQfTH2cgO-s88Bh2F0dG5ldHOIAAAAAAAAAACEZXRoMpDkvpOTAAAQIP__________gmlkgnY0gmlwhBLf22SJc2VjcDI1NmsxoQLV_jMOIxKbjHFKgrkFvwDvpexo6Nd58TK5k7ss4Vt0IoN1ZHCCG1g",
		// Lighthouse's bootnode by Afri
		"enr:-LK4QH1xnjotgXwg25IDPjrqRGFnH1ScgNHA3dv1Z8xHCp4uP3N3Jjl_aYv_WIxQRdwZvSukzbwspXZ7JjpldyeVDzMCh2F0dG5ldHOIAAAAAAAAAACEZXRoMpB53wQoAAAQIP__________gmlkgnY0gmlwhIe1te-Jc2VjcDI1NmsxoQOkcGXqbCJYbcClZ3z5f6NWhX_1YPFRYRRWQpJjwSHpVIN0Y3CCIyiDdWRwgiMo",
		// Lighthouse's bootnode by Sigp
		"enr:-Ly4QFPk-cTMxZ3jWTafiNblEZkQIXGF2aVzCIGW0uHp6KaEAvBMoctE8S7YU0qZtuS7By0AA4YMfKoN9ls_GJRccVpFh2F0dG5ldHOI__________-EZXRoMpCC9KcrAgAQIIS2AQAAAAAAgmlkgnY0gmlwhKh3joWJc2VjcDI1NmsxoQKrxz8M1IHwJqRIpDqdVW_U1PeixMW5SfnBD-8idYIQrIhzeW5jbmV0cw-DdGNwgiMog3VkcIIjKA",
		"enr:-L64QJmwSDtaHVgGiqIxJWUtxWg6uLCipsms6j-8BdsOJfTWAs7CLF9HJnVqFE728O-JYUDCxzKvRdeMqBSauHVCMdaCAVWHYXR0bmV0c4j__________4RldGgykIL0pysCABAghLYBAAAAAACCaWSCdjSCaXCEQWxOdolzZWNwMjU2azGhA7Qmod9fK86WidPOzLsn5_8QyzL7ZcJ1Reca7RnD54vuiHN5bmNuZXRzD4N0Y3CCIyiDdWRwgiMo",
		// Teku's bootnode By Afri
		"enr:-KG4QCIzJZTY_fs_2vqWEatJL9RrtnPwDCv-jRBuO5FQ2qBrfJubWOWazri6s9HsyZdu-fRUfEzkebhf1nvO42_FVzwDhGV0aDKQed8EKAAAECD__________4JpZIJ2NIJpcISHtbYziXNlY3AyNTZrMaED4m9AqVs6F32rSCGsjtYcsyfQE2K8nDiGmocUY_iq-TSDdGNwgiMog3VkcIIjKA",
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

	// TODO(fastex): set Oasis fork schedule, TTD and Deposint contract address.
	cfg.GenesisForkVersion = []byte{0, 0, 40, 90}
	cfg.AltairForkEpoch = 36660
	cfg.AltairForkVersion = []byte{1, 0, 40, 90}
	cfg.BellatrixForkEpoch = 112260
	cfg.BellatrixForkVersion = []byte{2, 0, 40, 90}
	cfg.CapellaForkEpoch = 162304
	cfg.CapellaForkVersion = []byte{3, 0, 40, 90}
	cfg.TerminalTotalDifficulty = "10790000"
	cfg.DepositContractAddress = "0xff50ed3d0ec03aC01D4C79aAd74928BFF48a7b2b"
	cfg.InitializeForkSchedule()
	return cfg
}
