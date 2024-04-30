package keystore

import (
	"encoding/hex"
	"fmt"
	"os"

	crypto "github.com/libp2p/go-libp2p-crypto"
)

type createFn func() ([]byte, error)

// CreateIfNotExists generates a private key at the specified path,
// or reads the file on that path if it is present
func CreateIfNotExists(path string, create createFn) ([]byte, error) {
	_, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to stat (%s): %w", path, err)
	}

	var keyBuff []byte
	if !os.IsNotExist(err) {
		// Key exists
		keyBuff, err = os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("unable to read private key from disk (%s), %w", path, err)
		}

		return keyBuff, nil
	}

	// Key doesn't exist yet, generate it
	keyBuff, err = create()
	if err != nil {
		return nil, fmt.Errorf("unable to generate private key, %w", err)
	}

	// Encode it to a readable format (Base64) and write to disk
	keyBuff = []byte(hex.EncodeToString(keyBuff))
	if err = os.WriteFile(path, keyBuff, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to write private key to disk (%s), %w", path, err)
	}

	return keyBuff, nil
}

func CreatePrivateKey(create createFn) ([]byte, error) {
	keyBuff, err := create()
	if err != nil {
		return nil, fmt.Errorf("unable to generate private key, %w", err)
	}

	// Encode it to a readable format (Base64) and return
	return []byte(hex.EncodeToString(keyBuff)), nil
}

// GenerateAndEncodeLibp2pKey generates a new networking private key, and encodes it into hex
func GenerateAndEncodeLibp2pKey() (crypto.PrivKey, []byte, error) {
	priv, _, err := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
	if err != nil {
		return nil, nil, err
	}

	buf, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return nil, nil, err
	}

	return priv, []byte(hex.EncodeToString(buf)), nil
}

// ParseLibp2pKey converts a byte array to a private key
func ParseLibp2pKey(key []byte) (crypto.PrivKey, error) {
	buf, err := hex.DecodeString(string(key))
	if err != nil {
		return nil, err
	}

	libp2pKey, err := crypto.UnmarshalPrivateKey(buf)
	if err != nil {
		return nil, err
	}

	return libp2pKey, nil
}
