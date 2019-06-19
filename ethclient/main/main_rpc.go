package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/simplechain-org/simplechain/common/hexutil"

	"github.com/simplechain-org/simplechain/ethclient"
	"gopkg.in/gin-gonic/gin.v1/json"
	"log"
	"math/big"
)

func main() {
	serverCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := ethclient.DialContext(serverCtx, "http://192.168.3.203:8545")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	block, err := client.BlockByNumber(serverCtx, big.NewInt(0))
	if err != nil {
		log.Fatal(err)
	}

	re, _ := json.Marshal(block.Header())
	fmt.Println(block.Transactions().Len()," blockHeader : ",string(re))
	trx, _ := json.Marshal(block.Transactions())
	fmt.Println(block.Transactions().Len()," trx : ",string(trx))

	uncles, _ := json.Marshal(block.Uncles())
	fmt.Println(block.Transactions().Len()," uncles : ",string(uncles))

	w := new(bytes.Buffer)
	err = block.EncodeRLP(w)
	if err != nil {
		log.Fatal(err)
	}
	ff :=hexutil.Encode(w.Bytes())
	fmt.Println("rlp:  ",ff)

}

//func main() {
//
//}