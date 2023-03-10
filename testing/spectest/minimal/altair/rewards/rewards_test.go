package rewards

import (
	"testing"

	"github.com/prysmaticlabs/prysm/v3/testing/spectest/shared/altair/rewards"
)

func TestMinimal_Altair_Rewards(t *testing.T) {
	rewards.RunPrecomputeRewardsAndPenaltiesTests(t, "minimal")
}
