package db

import (
	"github.com/simplechain-org/simplechain/common"
	"testing"
)

func TestInsetIdHash(t *testing.T) {
	hashDb, err := NewLDBDatabase("./IdHash", 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	TxId := common.BytesToHash(common.FromHex("0xd962b109b0bfdef7d6568cff8e6fe24d55e80d5749f6d80ddea66c0647dbb03a"))
	hashData := common.BytesToHash(common.FromHex("0xe267591b78ab7ffc97fab9e9ae55ae2db067225dde4e989b7ec071b125ca6b94"))

	if err := hashDb.InsertHash(hashData, TxId); err != nil {
		t.Fatal("InsertHash failed: ", err)
	}

	re, err := hashDb.GetHashId(hashData)
	if err != nil {
		t.Fatal(err)
	}
	if re != TxId {
		t.Fatal("the hash not equal ")
	}

	//fmt.Println(re.String(),err)
}
