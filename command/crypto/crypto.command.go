package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/sx-network/sx-reporter/helper/keystore"
	"github.com/sx-network/sx-reporter/helper/types"
	"golang.org/x/crypto/sha3"
)

// S256 is the secp256k1 elliptic curve
var S256 = btcec.S256()

// generateECDSAKeyAndMarshal generates a new ECDSA private key and serializes it to a byte array
func generateECDSAKeyAndMarshal() ([]byte, error) {
	key, err := GenerateECDSAKey()
	if err != nil {
		return nil, err
	}

	buf, err := MarshalECDSAPrivateKey(key)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// MarshalECDSAPrivateKey serializes the private key's D value to a []byte
func MarshalECDSAPrivateKey(priv *ecdsa.PrivateKey) ([]byte, error) {
	return (*btcec.PrivateKey)(priv).Serialize(), nil
}

// GenerateECDSAKey generates a new key based on the secp256k1 elliptic curve.
func GenerateECDSAKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(S256, rand.Reader)
}

// GenerateAndEncodeECDSAPrivateKey returns a newly generated private key and the Base64 encoding of that private key
func GenerateAndEncodeECDSAPrivateKey() (*ecdsa.PrivateKey, []byte, error) {
	keyBuff, err := keystore.CreatePrivateKey(generateECDSAKeyAndMarshal)
	if err != nil {
		return nil, nil, err
	}

	privateKey, err := BytesToECDSAPrivateKey(keyBuff)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to execute byte array -> private key conversion, %w", err)
	}

	return privateKey, keyBuff, nil
}

// BytesToECDSAPrivateKey reads the input byte array and constructs a private key if possible
func BytesToECDSAPrivateKey(input []byte) (*ecdsa.PrivateKey, error) {
	// The key file on disk should be encoded in Base64,
	// so it must be decoded before it can be parsed by ParsePrivateKey
	decoded, err := hex.DecodeString(string(input))
	if err != nil {
		return nil, err
	}

	// Make sure the key is properly formatted
	if len(decoded) != 32 {
		// Key must be exactly 64 chars (32B) long
		return nil, fmt.Errorf("invalid key length (%dB), should be 32B", len(decoded))
	}

	// Convert decoded bytes to a private key
	key, err := ParseECDSAPrivateKey(decoded)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func ParseECDSAPrivateKey(buf []byte) (*ecdsa.PrivateKey, error) {
	prv, _ := btcec.PrivKeyFromBytes(S256, buf)

	return prv.ToECDSA(), nil
}

// PubKeyToAddress returns the Ethereum address of a public key
func PubKeyToAddress(pub *ecdsa.PublicKey) types.Address {
	buf := Keccak256(MarshalPublicKey(pub)[1:])[12:]

	return types.BytesToAddress(buf)
}

// Keccak256 calculates the Keccak256
func Keccak256(v ...[]byte) []byte {
	h := sha3.NewLegacyKeccak256()
	for _, i := range v {
		h.Write(i)
	}

	return h.Sum(nil)
}

// MarshalPublicKey marshals a public key on the secp256k1 elliptic curve.
func MarshalPublicKey(pub *ecdsa.PublicKey) []byte {
	return elliptic.Marshal(S256, pub.X, pub.Y)
}

// func GenerateAndEncodeBLSSecretKey() (*bls_sig.SecretKey, []byte, error) {
// 	keyBuff, err := keystore.CreatePrivateKey(generateBLSKeyAndMarshal)
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	secretKey, err := BytesToBLSSecretKey(keyBuff)
// 	if err != nil {
// 		return nil, nil, fmt.Errorf("unable to execute byte array -> private key conversion, %w", err)
// 	}

// 	return secretKey, keyBuff, nil
// }

// BytesToECDSAPrivateKey reads the input byte array and constructs a private key if possible
// func BytesToBLSSecretKey(input []byte) (*bls_sig.SecretKey, error) {
// 	// The key file on disk should be encoded in Base64,
// 	// so it must be decoded before it can be parsed by ParsePrivateKey
// 	decoded, err := hex.DecodeString(string(input))
// 	if err != nil {
// 		return nil, err
// 	}

// 	sk := &bls_sig.SecretKey{}
// 	if err := sk.UnmarshalBinary(decoded); err != nil {
// 		return nil, err
// 	}

// 	return sk, nil
// }

// GenerateBLSKey generates a new BLS key
// func GenerateBLSKey() (*bls_sig.SecretKey, error) {
// 	blsPop := bls_sig.NewSigPop()

// 	_, sk, err := blsPop.Keygen()
// 	if err != nil {
// 		return nil, err
// 	}

// 	return sk, nil
// }

// generateBLSKeyAndMarshal generates a new BLS secret key and serializes it to a byte array
// func generateBLSKeyAndMarshal() ([]byte, error) {
// 	key, err := GenerateBLSKey()
// 	if err != nil {
// 		return nil, err
// 	}

// 	buf, err := key.MarshalBinary()
// 	if err != nil {
// 		return nil, err
// 	}

// 	return buf, nil
// }

// BLSSecretKeyToPubkeyBytes returns bytes of BLS Public Key corresponding to the given secret key
// func BLSSecretKeyToPubkeyBytes(key *bls_sig.SecretKey) ([]byte, error) {
// 	pubKey, err := key.GetPublicKey()
// 	if err != nil {
// 		return nil, err
// 	}

// 	marshalled, err := pubKey.MarshalBinary()
// 	if err != nil {
// 		return nil, err
// 	}

// 	return marshalled, nil
// }
