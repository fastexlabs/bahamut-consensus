package scorers

import (
	"os"
	"sync"

	"github.com/prysmaticlabs/prysm/v4/config/features"
	"github.com/sirupsen/logrus"
)

const (
	badResponsesScorerName  string = "bad-responses"
	blockProviderScorerName string = "block-provider"
	gossipScorerName        string = "gossip"
	peerStatusScorerName    string = "peer-status"
)

// singleton
var scorerLog entryT
var configureOnce sync.Once

type entryT interface {
	Debug(args ...interface{})
	WithFields(logrus.Fields) entryT
}

// fakeEntry no logs.
type fakeEntry struct{}

func (*fakeEntry) Debug(args ...interface{}) {}

func (fe *fakeEntry) WithFields(logrus.Fields) entryT {
	return fe
}

// configuredEntry for logging scoring process in configured OUT.
type configuredEntry struct {
	entry *logrus.Entry
}

func (cf *configuredEntry) Debug(args ...interface{}) {
	cf.entry.Debug(args...)
}

func (cf *configuredEntry) WithFields(logrus.Fields) entryT {
	return cf
}

// logger for single entry depends on EnableScorerLogging flag.
func logger() entryT {
	configureOnce.Do(func() {
		if features.Get().EnableScorerLogging {
			scorerLog = &configuredEntry{
				entry: configureScorerLogger("scorers.log", "scorers"),
			}
		} else {
			scorerLog = &fakeEntry{}
		}
	})
	return scorerLog
}

func configureScorerLogger(filePath, topic string) *logrus.Entry {
	log := logrus.NewEntry(logrus.New()).WithField("prefix", topic)
	log.Logger.SetLevel(logrus.DebugLevel)

	// Setting default logging stream as stdout.
	log.Logger.Out = os.Stdout

	if file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
		log.Logger.Out = file
	} else {
		log.Warnf("Failed open %s file for logging.", filePath)
	}

	return log
}
