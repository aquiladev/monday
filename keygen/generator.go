package keygen

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
)

type Key struct {
	PubKey           string
	CompressedPubKey string
	PrivKey          string
}

func generate(d *big.Int) *Key {
	curve := btcec.S256()
	x, y := curve.ScalarBaseMult(d.Bytes())

	priv := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
			X:     x,
			Y:     y,
		},
		D: d,
	}

	params := &chaincfg.Params{PrivateKeyID: 0x80}
	wif, _ := btcutil.NewWIF((*btcec.PrivateKey)(priv), params, false)

	pubKey := (*btcec.PublicKey)(&priv.PublicKey)
	address, _ := btcutil.NewAddressPubKey(pubKey.SerializeUncompressed(), params)
	compressedAddress, _ := btcutil.NewAddressPubKey(pubKey.SerializeCompressed(), params)

	return &Key{
		PubKey:           address.EncodeAddress(),
		CompressedPubKey: compressedAddress.EncodeAddress(),
		PrivKey:          wif.String(),
	}
}
