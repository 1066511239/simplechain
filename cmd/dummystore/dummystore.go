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

	"github.com/simplechain-org/simplechain/accounts/abi/bind"
	"github.com/simplechain-org/simplechain/cmd/dummystore/contract"
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

func main() {
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatalf(errPrefix+" connect http://localhost:8545: %v", err)
	}

	var privkeys = []string{
		"e97f894d3862f82acc6981eaf91f680861cb3bf55b7401e85f4a2dfda9f7d322",
		"5aedb85503128685e4f92b0cc95e9e1185db99339f9b85125c1e2ddc0f7c4c48",
		"6543d61166268b929166e7626b9eeb277feea8bc13bff6bd5f2d01fcb5543a3e",
		"3969a3c690cc238d337561d973c4d107ef7db28a63d23d10ca579742fe14450d",
		"53a04ee9a47dadf7ca2b13c61eae1f05205f183e029ebce95f52f68e12dded5f",
		"5fc3281bd2894b8418a895897fbc7fe204099ae2f99f930da195d393976d1bbc",
		"ac817b6310c8ce7a2fd6783a663e4d47f0ab1fb3ae2fba5f85f07bdca5a36e90",
		"94dbb9085d79578046c4b10a0a7baefead738371ac763ee6ed7d727d348a2509",
		"f3a131e6001f2d0e51b9e2a7fb83953d911fe242f0ccec0c1e87d3a113527245",
		"52fa69273ecd3d889326c575215e7dcf95abf7fd1446d794b6e945d4d9ed2775",
		"aab4e132211a24255992c13ac3c6d09270496e141fe0ad2c79ecf19ca31aa8be",
		"8041b982caa012337dbfa21183ffb88361f4ffdd65b168d5a6aa8f0949a123b0",
	}

	countChans := make([]chan int, len(privkeys))

	ctx, cancel := context.WithCancel(context.Background())

	instance := deployStore(client)

	for i, privKey := range privkeys {
		newCountChan := make(chan int)
		countChans[i] = newCountChan
		go dummyTx(ctx, client, instance, i, privKey, newCountChan)
	}

	go calcTotalCount(ctx, countChans)

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

func dummyTx(ctx context.Context, client *ethclient.Client, instance *contract.Store, index int, privKey string, count chan<- int) {
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

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf(errPrefix+" get new nonce: %v", err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf(errPrefix+" get gas price: %v", err)
	}
	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	var (
		genTimer   = time.NewTicker(dummyInterval)
		meterCount = 0
		data       [32]byte
	)

	copy(data[:], common.Hex2Bytes("0xd962b109b0bfdef7d6568cff8e6fe24d55e80d5749f6d80ddea66c0647dbb03a"))

	for {
		select {
		case <-genTimer.C:
			count <- meterCount / 5
			meterCount = 0
		case <-ctx.Done():
			log.Printf("dummyTx:%v return", index)
			close(count)
			return
		default:
			_, err = instance.Set(auth, data, data)
			if err != nil {
				log.Printf(warnPrefix+" call Set: %v", err)
				if strings.Contains(err.Error(), "insufficient funds for gas * price + value") {
					claimFunds(client, fromAddress)
					//waiting transfer tx
					time.Sleep(dummyInterval)
				}
			}

			nonce, err = client.PendingNonceAt(context.Background(), fromAddress)
			if err != nil {
				log.Fatalf(errPrefix+" get new nonce: %v", err)
			}

			auth.Nonce = big.NewInt(int64(nonce))
			meterCount++
		}
	}
}

func calcTotalCount(ctx context.Context, countChans []chan int) {
	totalCount := 0
	for {
		select {
		case <-ctx.Done():
			log.Println("calcTotalCount return")
			return
		default:
			for i := range countChans {
				totalCount += <-countChans[i]
			}
			log.Printf("average per second dummy txs: %v", totalCount)
			totalCount = 0
		}
	}
}

func claimFunds(client *ethclient.Client, toAddress common.Address) {
	//0xffd79941b7085805f48ded97298694c6bb950e2c
	privateKey, err := crypto.HexToECDSA("04c070620a899a470a669fdbe0c6e1b663fd5bc953d9411eb36faa382005b3ad")
	if err != nil {
		log.Fatalf(errPrefix+" parse private key: %v", err)
	}
	fromAddress := common.HexToAddress("0xffd79941b7085805f48ded97298694c6bb950e2c")

	value := new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(10000)) // in wei (10000 eth)
	gasLimit := uint64(21000)                                                     // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf(errPrefix+"get gas price: %v", err)
	}

TryAgain:
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf(errPrefix+"get new nonce: %v", err)
	}

	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(new(big.Int).SetInt64(110)), privateKey)
	if err != nil {
		log.Fatalf(errPrefix+"sign tx: %v", err)
	}
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Printf(warnPrefix+"send tx: %v", err)
		if strings.Contains(err.Error(), "replacement transaction underpriced") {
			log.Println("waiting 5s and try again")
			time.Sleep(dummyInterval)
			goto TryAgain
		}
	}
}

func deployStore(client *ethclient.Client) *contract.Store {
	privateKey, err := crypto.HexToECDSA("04c070620a899a470a669fdbe0c6e1b663fd5bc953d9411eb36faa382005b3ad")
	if err != nil {
		log.Fatalf(errPrefix+" parse private key: %v", err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatalf(errPrefix + "cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf(errPrefix+"get new nonce: %v", err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf(errPrefix+"get gas price: %v", err)
	}

	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	address, tx, instance, err := contract.DeployStore(auth, client)
	if err != nil {
		log.Fatalf(errPrefix+" deploy contract", err)
	}

	log.Printf("new contract address  = %v", address.Hex())
	log.Printf("this transaction hash = %v", tx.Hash().String())

	return instance
}