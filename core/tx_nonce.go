package core

import (
	"github.com/simplechain-org/simplechain/core/types"
	"github.com/simplechain-org/simplechain/log"

	mapset "github.com/deckarep/golang-set"
)

type TxNonceCheck struct {
	cache mapset.Set
}

func (c *TxNonceCheck) InsertCache(transaction *types.Transaction) {
	c.cache.Add(transaction.Nonce())
}

func (c *TxNonceCheck) DeleteCache(nonce types.TxNonce) {
	c.cache.Remove(nonce)
}

func (c *TxNonceCheck) IsNonceOk(transaction *types.Transaction) bool {
	nonce := transaction.Nonce()
	if c.cache.Contains(nonce) {
		log.Trace("CommonTransactionNonceCheck: isNonceOk: duplicated nonce", "transHash", transaction.Hash())
		return false
	}

	c.cache.Add(nonce)
	return true
}
