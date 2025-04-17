package p2p

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/prysmaticlabs/prysm/v4/consensus-types/wrapper"
	ecdsaprysm "github.com/prysmaticlabs/prysm/v4/crypto/ecdsa"
	"github.com/prysmaticlabs/prysm/v4/io/file"
	pb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1/metadata"
	"github.com/sirupsen/logrus"
)

const (
	keyPath      = "network-keys"
	metaDataPath = "metaData.ssz"
)

const dialTimeout = 1 * time.Second

var ErrInvalidMetaData = errors.New("invalid metaData type")

// SerializeENR takes the enr record in its key-value form and serializes it.
func SerializeENR(record *enr.Record) (string, error) {
	if record == nil {
		return "", errors.New("could not serialize nil record")
	}
	buf := bytes.NewBuffer([]byte{})
	if err := record.EncodeRLP(buf); err != nil {
		return "", errors.Wrap(err, "could not encode ENR record to bytes")
	}
	enrString := base64.RawURLEncoding.EncodeToString(buf.Bytes())
	return enrString, nil
}

// Determines a private key for p2p networking from the p2p service's
// configuration struct. If no key is found, it generates a new one.
func privKey(cfg *Config) (*ecdsa.PrivateKey, error) {
	defaultKeyPath := path.Join(cfg.DataDir, keyPath)
	privateKeyPath := cfg.PrivateKey

	// PrivateKey cli flag takes highest precedence.
	if privateKeyPath != "" {
		return privKeyFromFile(cfg.PrivateKey)
	}

	_, err := os.Stat(defaultKeyPath)
	defaultKeysExist := !os.IsNotExist(err)
	if err != nil && defaultKeysExist {
		return nil, err
	}
	// Default keys have the next highest precedence, if they exist.
	if defaultKeysExist {
		return privKeyFromFile(defaultKeyPath)
	}
	// There are no keys on the filesystem, so we need to generate one.
	priv, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
	if err != nil {
		return nil, err
	}
	// If the StaticPeerID flag is set, save the generated key as the default
	// key, so that it will be used by default on the next node start.
	if cfg.StaticPeerID {
		rawbytes, err := priv.Raw()
		if err != nil {
			return nil, err
		}
		dst := make([]byte, hex.EncodedLen(len(rawbytes)))
		hex.Encode(dst, rawbytes)
		if err := file.WriteFile(defaultKeyPath, dst); err != nil {
			return nil, err
		}
		log.Infof("Wrote network key to file")
		// Read the key from the defaultKeyPath file just written
		// for the strongest guarantee that the next start will be the same as this one.
		return privKeyFromFile(defaultKeyPath)
	}
	return ecdsaprysm.ConvertFromInterfacePrivKey(priv)
}

// Retrieves a p2p networking private key from a file path.
func privKeyFromFile(path string) (*ecdsa.PrivateKey, error) {
	src, err := os.ReadFile(path) // #nosec G304
	if err != nil {
		log.WithError(err).Error("Error reading private key from file")
		return nil, err
	}
	dst := make([]byte, hex.DecodedLen(len(src)))
	_, err = hex.Decode(dst, src)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode hex string")
	}
	unmarshalledKey, err := crypto.UnmarshalSecp256k1PrivateKey(dst)
	if err != nil {
		return nil, err
	}
	return ecdsaprysm.ConvertFromInterfacePrivKey(unmarshalledKey)
}

// Retrieves node p2p metadata from a set of configuration values
// from the p2p service.
func metaDataFromConfig(cfg *Config) (metadata.Metadata, error) {
	metaDataPath := path.Join(cfg.DataDir, metaDataPath)
	if cfg.MetaDataDir != "" {
		metaDataPath = cfg.MetaDataDir
	}

	_, err := os.Stat(metaDataPath)
	metadataExist := !os.IsNotExist(err)
	if err != nil && metadataExist {
		return nil, err
	}
	if !metadataExist {
		metaData := &pb.MetaDataV0{
			SeqNumber: 0,
			Attnets:   bitfield.NewBitvector64(),
		}
		dst, err := metaData.MarshalSSZ()
		if err != nil {
			return nil, err
		}
		if err := file.WriteFile(metaDataPath, dst); err != nil {
			return nil, err
		}
		return wrapper.WrappedMetadataV0(metaData), nil
	}

	src, err := os.ReadFile(metaDataPath) // #nosec G304
	if err != nil {
		log.WithError(err).Error("Error reading metadata from file")
		return nil, err
	}

	metaDataV0 := &pb.MetaDataV0{}
	if err := metaDataV0.UnmarshalSSZ(src); err == nil {
		return wrapper.WrappedMetadataV0(metaDataV0), nil
	}
	metaDataV1 := &pb.MetaDataV1{}
	if err := metaDataV1.UnmarshalSSZ(src); err == nil {
		return wrapper.WrappedMetadataV1(metaDataV1), nil
	}
	return nil, ErrInvalidMetaData
}

func writeMetaData(cfg *Config, metaData metadata.Metadata) (err error) {
	if metaData == nil || metaData.IsNil() {
		return errors.New("nil metaData provided")
	}

	metaDataPath := path.Join(cfg.DataDir, metaDataPath)
	if cfg.MetaDataDir != "" {
		metaDataPath = cfg.MetaDataDir
	}

	_, err = os.Stat(metaDataPath)
	metadataExist := !os.IsNotExist(err)
	if err != nil && metadataExist {
		return err
	}

	dst, err := metaData.MarshalSSZ()
	if err != nil {
		return err
	}

	return file.WriteFile(metaDataPath, dst)
}

// Attempt to dial an address to verify its connectivity
func verifyConnectivity(addr string, port uint, protocol string) {
	if addr != "" {
		a := net.JoinHostPort(addr, fmt.Sprintf("%d", port))
		fields := logrus.Fields{
			"protocol": protocol,
			"address":  a,
		}
		conn, err := net.DialTimeout(protocol, a, dialTimeout)
		if err != nil {
			log.WithError(err).WithFields(fields).Warn("IP address is not accessible")
			return
		}
		if err := conn.Close(); err != nil {
			log.WithError(err).Debug("Could not close connection")
		}
	}
}
