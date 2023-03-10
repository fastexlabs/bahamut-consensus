package params

import "math"

func UseFastexChainTestnetConfig() {
	cfg := BeaconNetworkConfig().Copy()
	cfg.ContractDeploymentBlock = 0
	cfg.BootstrapNodes = []string{
		"enr:-MK4QNy2nO0u5AC6mbteqkJDwsJMK-HCvZVYNF8CPGljq699YmJaW0bgAabGREHay8UaKwUMFzHNZ6KF7bz14WYXwVOGAYYp_Nmqh2F0dG5ldHOI___9______-EZXRoMpDOhQ4BAgAAKv__________gmlkgnY0gmlwhKwehzyJc2VjcDI1NmsxoQNpwmCiCDAoW8vPa3rYUoOVDhEhfmykhAMeJUhJbA_dTYhzeW5jbmV0cw-DdGNwgjLIg3VkcIIu4A",
		"enr:-MK4QKO3YYGupjaQOAZyjXdNHHgZ3o3VBsIBwPz7zGrcrXpJAnuOD69fvnE0JpfkyuZ3heHhcF5htkCqvAPdW1lTxOSGAYYp_Pexh2F0dG5ldHOI__________-EZXRoMpDOhQ4BAgAAKv__________gmlkgnY0gmlwhKwQT2GJc2VjcDI1NmsxoQIobcN13Okg6WgSn0mLhT0KUDJ4vVvqYbs97hSJdIe_DIhzeW5jbmV0cw-DdGNwgjLIg3VkcIIu4A",
		"enr:-MK4QA2hYODoz20ZFesfPKQMKx3wkDt6owIJHayO6X4AdZxaBF1ZwiDCC7TqD0T91sd81HN72lKItX9zMxFdq9rojRGGAYYp_NRdh2F0dG5ldHOI_3________-EZXRoMpDOhQ4BAgAAKv__________gmlkgnY0gmlwhDEMK-uJc2VjcDI1NmsxoQLuDi9eJ-dqsHqHtQpIIy-OgHhLJ_u5b4-3LA4OuutrkYhzeW5jbmV0cw-DdGNwgjLIg3VkcIIu4A",
	}
	OverrideBeaconNetworkConfig(cfg)
}

func FastexChainTestnetConfig() *BeaconChainConfig {
	cfg := MainnetConfig().Copy()
	cfg.MinGenesisActiveValidatorCount = 768 // ONLY FOR TESTING
	cfg.Eth1FollowDistance = 2048
	cfg.SecondsPerETH1Block = 12
	cfg.MinGenesisTime = 1673948262
	cfg.GenesisDelay = 30
	cfg.ConfigName = FastexChainTestnet
	cfg.PresetBase = "fastex-chain-testnet"
	cfg.DepositChainID = 424242
	cfg.DepositNetworkID = 424242
	cfg.SecondsPerETH1Block = 12
	cfg.EpochsPerEth1VotingPeriod = 64
	cfg.MaxEffectiveBalance = 8192 * 1e9
	cfg.EjectionBalance = 4096 * 1e9
	cfg.SigmoidExpCoefficient = -1.5
	cfg.SigmoidLimit = 0.62
	cfg.GenesisForkVersion = []byte{0x00, 0x00, 0x00, 0x2A}
	cfg.AltairForkEpoch = 1
	cfg.AltairForkVersion = []byte{0x01, 0x00, 0x00, 0x2A}
	cfg.BellatrixForkEpoch = 2
	cfg.BellatrixForkVersion = []byte{0x02, 0x00, 0x00, 0x2A}
	cfg.CapellaForkEpoch = math.MaxUint64
	cfg.CapellaForkVersion = []byte{0x03, 0x00, 0x00, 0x2A}
	cfg.TerminalTotalDifficulty = "113400"
	cfg.GenesisDelay = 44000
	cfg.DepositContractAddress = "0x0000000006011920201511051404051615190920"
	cfg.InitializeForkSchedule()
	return cfg
}
