package yubi

import (
	"fmt"
	"context"
	"crypto/ecdsa"
"crypto/elliptic"

	"github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/ecadlabs/signatory/pkg/vault/yubi"
	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"

)



// NewPrivKeySecp256k1Unsafe will attach to a key and store the public key for later use.
//
// This function is marked as unsafe as it will retrieve a pubkey without user verification.
// It can only be used to verify a pubkey but never to create new accounts/keys. In that case,
// please refer to NewPrivKeySecp256k1
func NewPrivKeySecp256k1Unsafe() (types.YubiPrivKey, error) {
	config := &yubi.Config{
		Address: "127.0.0.1:12345",
		Password: "",
		AuthKeyID: 1,
		KeyImportDomains: 1,
	}

	hsm, err := yubi.New(context.Background(), config)
	if err != nil {
		return nil, err
	}

	pubkey, err := getPubKeyUnsafe(hsm)
	if err != nil {
		return nil, err
	}

	return &PrivKeyYubiSecp256k1{
		CachedPubKey: pubkey,
	}, nil
}

// getPubKeyUnsafe reads the pubkey from a ledger device
//
// This function is marked as unsafe as it will retrieve a pubkey without user verification
// It can only be used to verify a pubkey but never to create new accounts/keys. In that case,
// please refer to getPubKeyAddrSafe
//
// since this involves IO, it may return an error, which is not exposed
// in the PubKey interface, so this function allows better error handling
func getPubKeyUnsafe(hsm *yubi.HSM) (types.PubKey, error) {
	publicKey, err := hsm.GetPublicKey(context.Background(), "1729") // Tezos, for now
	if err != nil {
		return nil, fmt.Errorf("Could not connect to yubi", err)
	}

	// publicKey = storedKey.PublicKey()
	pubKey, ok := publicKey.PublicKey().(ecdsa.PublicKey)
	if !ok {
		fmt.Errorf("Could not assert the public key to ed25519 public key")
	}
	pubKeyBytes := elliptic.Marshal(pubKey, pubKey.X, pubKey.Y)

	// re-serialize in the 33-byte compressed format
	cmp, err := btcec.ParsePubKey(pubKeyBytes, btcec.S256())
	if err != nil {
		return nil, fmt.Errorf("error parsing public key: %v", err)
	}

	compressedPublicKey := make([]byte, secp256k1.PubKeySize)
	copy(compressedPublicKey, cmp.SerializeCompressed())

	return &secp256k1.PubKey{Key: compressedPublicKey}, nil
}

type PrivKeyYubiSecp256k1 struct {
	// CachedPubKey should be private, but we want to encode it via
	// go-amino so we can view the address later, even without having the
	// ledger attached.
	CachedPubKey types.PubKey
}

// PubKey returns the cached public key.
func (pkl PrivKeyYubiSecp256k1) PubKey() types.PubKey {
	return pkl.CachedPubKey
}
