package p2p

import (
	"crypto/rand"
	"fmt"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/prysmaticlabs/prysm/v4/config/params"
	"github.com/prysmaticlabs/prysm/v4/consensus-types/wrapper"
	pb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v4/testing/assert"
	"github.com/prysmaticlabs/prysm/v4/testing/require"
	logTest "github.com/sirupsen/logrus/hooks/test"
)

// Test `verifyConnectivity` function by trying to connect to google.com (successfully)
// and then by connecting to an unreachable IP and ensuring that a log is emitted
func TestVerifyConnectivity(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	hook := logTest.NewGlobal()
	cases := []struct {
		address              string
		port                 uint
		expectedConnectivity bool
		name                 string
	}{
		{"142.250.68.46", 80, true, "Dialing a reachable IP: 142.250.68.46:80"}, // google.com
		{"123.123.123.123", 19000, false, "Dialing an unreachable IP: 123.123.123.123:19000"},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf(tc.name),
			func(t *testing.T) {
				verifyConnectivity(tc.address, tc.port, "tcp")
				logMessage := "IP address is not accessible"
				if tc.expectedConnectivity {
					require.LogsDoNotContain(t, hook, logMessage)
				} else {
					require.LogsContain(t, hook, logMessage)
				}
			})
	}
}

func TestSerializeENR(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	t.Run("Ok", func(t *testing.T) {
		key, err := crypto.GenerateKey()
		require.NoError(t, err)
		db, err := enode.OpenDB("")
		require.NoError(t, err)
		lNode := enode.NewLocalNode(db, key)
		record := lNode.Node().Record()
		s, err := SerializeENR(record)
		require.NoError(t, err)
		assert.NotEqual(t, "", s)
		s = "enr:" + s
		newRec, err := enode.Parse(enode.ValidSchemes, s)
		require.NoError(t, err)
		assert.Equal(t, s, newRec.String())
	})

	t.Run("Nil record", func(t *testing.T) {
		_, err := SerializeENR(nil)
		require.NotNil(t, err)
		assert.ErrorContains(t, "could not serialize nil record", err)
	})
}

func TestMetaDataFromConfig(t *testing.T) {
	metaDataPath := "./metaDataTest"
	os.Remove(metaDataPath)
	cfg := &Config{
		MetaDataDir: metaDataPath,
	}
	emptyMetaData := &pb.MetaDataV0{
		SeqNumber: 0,
		Attnets:   bitfield.NewBitvector64(),
	}
	// no metadata file
	t.Log("No metadata file")
	metaData, err := metaDataFromConfig(cfg)
	require.NoError(t, err)
	require.NotNil(t, metaData)
	require.DeepEqual(t, emptyMetaData, metaData.MetadataObjV0())
	require.NoError(t, os.Remove(metaDataPath))

	var (
		attnets  [8]byte
		syncnets [1]byte
	)
	rand.Read(attnets[:])
	rand.Read(syncnets[:])

	// existring metadata V0
	t.Log("existring metadata V0")
	metaDataV0 := &pb.MetaDataV0{
		SeqNumber: 10,
		Attnets:   bitfield.Bitvector64(attnets[:]),
	}
	require.NoError(t, writeMetaData(cfg, wrapper.WrappedMetadataV0(metaDataV0)))
	metaData, err = metaDataFromConfig(cfg)
	require.NoError(t, err)
	require.NotNil(t, metaData)
	require.DeepEqual(t, metaDataV0, metaData.MetadataObjV0())
	require.NoError(t, os.Remove(metaDataPath))

	// existring metadata V1
	t.Log("existring metadata V1")
	metaDataV1 := &pb.MetaDataV1{
		SeqNumber: 11,
		Attnets:   bitfield.Bitvector64(attnets[:]),
		Syncnets:  bitfield.Bitvector4(syncnets[:]),
	}
	require.NoError(t, writeMetaData(cfg, wrapper.WrappedMetadataV1(metaDataV1)))
	metaData, err = metaDataFromConfig(cfg)
	require.NoError(t, err)
	require.NotNil(t, metaData)
	require.DeepEqual(t, metaDataV1, metaData.MetadataObjV1())
	require.NoError(t, os.Remove(metaDataPath))

	// invalid metadata
	t.Log("Invalid metadata")
	var buf [32]byte
	rand.Read(buf[:])
	require.NoError(t, os.WriteFile(metaDataPath, buf[:], 0o644))
	metaData, err = metaDataFromConfig(cfg)
	require.ErrorIs(t, err, ErrInvalidMetaData)
	require.NoError(t, os.Remove(metaDataPath))
}
