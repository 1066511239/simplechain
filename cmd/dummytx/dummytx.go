package main

import (
	"context"
	"crypto/ecdsa"
	"log"
	"math/big"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/simplechain-org/simplechain/common"
	"github.com/simplechain-org/simplechain/core/types"
	"github.com/simplechain-org/simplechain/crypto"
	"github.com/simplechain-org/simplechain/ethclient"
)

const (
	dummyInterval = 5 * time.Second
	warnPrefix    = "\x1b[93mwarn:\x1b[0m"
	errPrefix     = "\x1b[91merror:\x1b[0m"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main() {
	client, err := ethclient.Dial("ws://localhost:8546")
	if err != nil {
		log.Fatalf(errPrefix+" connect ws://localhost:8546: %v", err)
	}

	var sourceKey = []string{
		"e97f894d3862f82acc6981eaf91f680861cb3bf55b7401e85f4a2dfda9f7d322",
		"5aedb85503128685e4f92b0cc95e9e1185db99339f9b85125c1e2ddc0f7c4c48",
		"6543d61166268b929166e7626b9eeb277feea8bc13bff6bd5f2d01fcb5543a3e",
		"3969a3c690cc238d337561d973c4d107ef7db28a63d23d10ca579742fe14450d",
		"53a04ee9a47dadf7ca2b13c61eae1f05205f183e029ebce95f52f68e12dded5f",
		"5fc3281bd2894b8418a895897fbc7fe204099ae2f99f930da195d393976d1bbc",
		"ac817b6310c8ce7a2fd6783a663e4d47f0ab1fb3ae2fba5f85f07bdca5a36e90",
		"94dbb9085d79578046c4b10a0a7baefead738371ac763ee6ed7d727d348a2509",
	}

	var priKeys []string
	if len(os.Args) > 1 && os.Args[1] == "1" {
		priKeys = sourceKey[:1]
	} else if len(os.Args) > 1 && os.Args[1] == "2" {
		priKeys = sourceKey[:2]
	} else if len(os.Args) > 1 && os.Args[1] == "8" {
		priKeys = sourceKey
	} else {
		priKeys = sourceKey[:4]
	}

	ctx, cancel := context.WithCancel(context.Background())

	for i, priKey := range priKeys {
		go dummyTx(ctx, client, i, priKey)
	}

	//go calcTotalCount(ctx, client)

	go func() {
		http.ListenAndServe("127.0.0.1:6789", nil)
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(interrupt)
	<-interrupt
	cancel()

	time.Sleep(time.Second)
	log.Println("dummy transaction exit")
}

func dummyTx(ctx context.Context, client *ethclient.Client, index int, privKey string) {
	privateKey, err := crypto.HexToECDSA(privKey)
	if err != nil {
		log.Fatalf(errPrefix+" parse private key: %v", err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatalf(errPrefix + " cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		log.Fatalf(errPrefix+" get new nonce: %v", err)
	}
	value := big.NewInt(0)            // in wei (0.0001 eth)
	gasLimit := uint64(21000 + 52*68) // in units
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatalf(errPrefix+" get gas price: %v", err)
	}
	toAddress := common.HexToAddress("0xffd79941b7085805f48ded97298694c6bb950e2c")

	var data [20 + 32]byte
	copy(data[:], fromAddress.Bytes())
	copy(data[20:], common.FromHex("0xd962b109b0bfdef7d6568cff8e6fe24d55e80d5749f6d80ddea66c0647dbb03a"))

	var (
		genTimer   = time.NewTimer(0)
		meterCount = 0
	)

	<-genTimer.C
	genTimer.Reset(dummyInterval)

	for {
		select {
		case <-genTimer.C:
			log.Printf("%v send %v txs in 5s", index, meterCount)
			meterCount = 0
			genTimer.Reset(dummyInterval)
		case <-ctx.Done():
			log.Printf("dummyTx:%v return", index)
			return
		default:
			//build,sign,send transaction
			dummy(ctx, nonce, toAddress, value, gasLimit, gasPrice, data[:], client, fromAddress)

			switch {
			case nonce%20000 == 0:
				nonce, err = client.PendingNonceAt(ctx, fromAddress)
				if err != nil {
					log.Fatalf(errPrefix+" get new nonce: %v", err)
				}
			default:
				nonce++
			}

			meterCount++

		}
	}
}

func dummy(ctx context.Context, nonce uint64, toAddress common.Address, value *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, client *ethclient.Client, fromAddress common.Address) {
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)
	err := client.SendTransaction(ctx, tx)
	if err != nil {
		log.Printf(warnPrefix+" send tx: %v", err)
		if strings.Contains(err.Error(), "known transaction") {
			log.Printf("tx is %v:", tx)
		}
	}
}

//func calcTotalCount(ctx context.Context, client *ethclient.Client) {
//	timer := time.NewTimer(0)
//	<-timer.C
//	timer.Reset(dummyInterval)
//	for {
//		select {
//		case <-ctx.Done():
//			log.Println("calcTotalCount return")
//			return
//		case <-timer.C:
//			txCount, err := client.LatestTransactionCount(ctx)
//			if err != nil {
//				log.Printf(warnPrefix, "get latest txCount: %v", err)
//			}
//			log.Printf("average per second final txs persond: %v", txCount/5)
//			timer.Reset(dummyInterval)
//		default:
//
//		}
//	}
//}

//func claimFunds(ctx context.Context, client *ethclient.Client, toAddress common.Address) {
//	//0xffd79941b7085805f48ded97298694c6bb950e2c
//	privateKey, err := crypto.HexToECDSA("04c070620a899a470a669fdbe0c6e1b663fd5bc953d9411eb36faa382005b3ad")
//	if err != nil {
//		log.Fatalf(errPrefix+" parse private key: %v", err)
//	}
//	fromAddress := common.HexToAddress("0xffd79941b7085805f48ded97298694c6bb950e2c")
//
//	value := new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(10000)) // in wei (10000 eth)
//	gasLimit := uint64(21000)                                                     // in units
//	gasPrice, err := client.SuggestGasPrice(ctx)
//	if err != nil {
//		log.Fatalf(errPrefix+" get gas price: %v", err)
//	}
//
//	doClaim(ctx, client, fromAddress, toAddress, value, gasLimit, gasPrice, privateKey)
//}

func doClaim(ctx context.Context, client *ethclient.Client, fromAddress common.Address, toAddress common.Address, value *big.Int, gasLimit uint64, gasPrice *big.Int, privateKey *ecdsa.PrivateKey) {
	for {
		nonce, err := client.PendingNonceAt(ctx, fromAddress)
		if err != nil {
			log.Fatalf(errPrefix+" get new nonce: %v", err)
		}
		tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)
		signedTx, err := types.SignTx(tx, types.NewEIP155Signer(new(big.Int).SetInt64(110)), privateKey)
		if err != nil {
			log.Fatalf(errPrefix+" sign tx: %v", err)
		}

		err = client.SendTransaction(ctx, signedTx)
		switch {
		case err != nil:
			log.Printf(warnPrefix+" send tx: %v, sender=%v, receiver=%v", err, fromAddress.String(), toAddress.String())
			if !strings.Contains(err.Error(), "replacement transaction underpriced") {
				return
			}
			log.Println("waiting 5s and try again")
			time.Sleep(dummyInterval)
		default:
			return
		}
	}
}
