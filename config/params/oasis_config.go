// todo unit act
package params

// UseOasisNetworkConfig uses the Oasis specific
// network config.
func UseOasisNetworkConfig() {
	cfg := BeaconNetworkConfig().Copy()
	cfg.ContractDeploymentBlock = 502405
	cfg.BootstrapNodes = []string{
		// FastexChain Oasis bootnode.
		"enr:-LG4QI1-Qu7pmmPwxCiaYCxtbuIQzS0GGZAtVpRcobNMS_fzDq4GTQokN721cZGFsVR70HKswJmKS1m1sZbB5AVCF9eGAYl4tg01h2F0dG5ldHOIAAAAAAAAAACEZXRoMpCJnEJeAAAoWv__________gmlkgnY0gmlwhCPDPg2Jc2VjcDI1NmsxoQP15vbtLz7X3iiyLtt2FmV1IWMtrvglrSW8VpIEdabk14N1ZHCCIyg",
		"enr:-MK4QDW72KTjU2RcUd3SMUWz9vW9IHMCstEyiErHLnw5i-kIO0FHBYsSaV3xxYI8ynjGnRSutgfCqheaOVJ0dVlhFlKGAYoXK-y5h2F0dG5ldHOIAAAAAAAAAACEZXRoMpDMilfyAwAoWv__________gmlkgnY0gmlwhCPDPg2Jc2VjcDI1NmsxoQJINwhsm_QLHl4wLSJ2HptdrGc1DIFsNUrKxaFBADcZB4hzeW5jbmV0cwCDdGNwgjLIg3VkcIIu4A",
		"enr:-MK4QEAoUybSi3BMLP4Pig3-N02Um0ANaBOtFCGVWlvfeQb5GQrcIIFA9eQYUKtnkWHmDNf5Avqt0TNoTO-GHfZv4V-GAYoXN1lzh2F0dG5ldHOIAAAAAAAAAACEZXRoMpDMilfyAwAoWv__________gmlkgnY0gmlwhCJYh7mJc2VjcDI1NmsxoQMv3ezcnOCiGBm0PSnt7sgPVNDVZ1NAC1igDKWRlcBBe4hzeW5jbmV0cwCDdGNwgjLIg3VkcIIu4A",
		"enr:-MK4QPbEhAr_q9JW6U76VgZVdKyLhaVo82IDyuuTTQdHNzucPkdyWhVA15RSuH-SiPmUXjORtPS2f7msNbB-ynO_KzeGAYoXc4IFh2F0dG5ldHOIAAAAAAAAAACEZXRoMpDMilfyAwAoWv__________gmlkgnY0gmlwhCJYycmJc2VjcDI1NmsxoQJ4dFs_DPW7THGr7KfmuvdNUXLoFjZ1YfaPFUl89H8b0YhzeW5jbmV0cwCDdGNwgjLIg3VkcIIu4A",
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
	cfg.DenebForkEpoch = cfg.FarFutureEpoch
	cfg.DenebForkVersion = []byte{4, 0, 40, 90}
	cfg.TerminalTotalDifficulty = "1205400" // ~ 2023-07-25 08:03:10 UTC
	cfg.DepositContractAddress = "0x84e09dbd827b66bfa66f42fc50dc732c01790122"
	cfg.InitializeForkSchedule()

	cfg.FirstFixFork = false
	return cfg
}
