package scorers

import (
	"flag"
	"os"
	"testing"

	"github.com/prysmaticlabs/prysm/v4/config/features"
	"github.com/prysmaticlabs/prysm/v4/testing/require"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func TestLoger(t *testing.T) {
	t.Run("logger testing", func(t *testing.T) {
		app := cli.App{}
		set := flag.NewFlagSet("test", 0)
		set.Bool(features.EnableScorerLogging.Name, true, "test")
		require.NoError(t, features.ConfigureBeaconChain(cli.NewContext(&app, set, nil)))
		logger().WithFields(logrus.Fields{
			"Field1": "Filed1 msg",
		}).Debug("Some message")
		require.NoError(t, os.Remove("scorers.log"))
	})
}
