package entity

import "github.com/gagliardetto/solana-go"

type PubKeyPair struct {
	Key   solana.PublicKey
	Value solana.PublicKey
}
