package version

const (
	Phase0 = iota
	Altair
	Bellatrix
	FastexPhase1
	Capella
)

func String(version int) string {
	switch version {
	case Phase0:
		return "phase0"
	case Altair:
		return "altair"
	case Bellatrix:
		return "bellatrix"
	case FastexPhase1:
		return "fastex-phase1"
	case Capella:
		return "capella"
	default:
		return "unknown version"
	}
}

// All returns a list of all known fork versions.
func All() []int {
	return []int{Phase0, Altair, Bellatrix, FastexPhase1, Capella}
}
