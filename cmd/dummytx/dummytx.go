package main

import (
	"context"
	"crypto/ecdsa"
	"flag"
	"log"
	"math/big"
	"math/rand"
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

var (
	sourceKey = []string{
		"e97f894d3862f82acc6981eaf91f680861cb3bf55b7401e85f4a2dfda9f7d322",
		"5aedb85503128685e4f92b0cc95e9e1185db99339f9b85125c1e2ddc0f7c4c48",
		"6543d61166268b929166e7626b9eeb277feea8bc13bff6bd5f2d01fcb5543a3e",
		"3969a3c690cc238d337561d973c4d107ef7db28a63d23d10ca579742fe14450d",
		"53a04ee9a47dadf7ca2b13c61eae1f05205f183e029ebce95f52f68e12dded5f",
		"5fc3281bd2894b8418a895897fbc7fe204099ae2f99f930da195d393976d1bbc",
		"ac817b6310c8ce7a2fd6783a663e4d47f0ab1fb3ae2fba5f85f07bdca5a36e90",
		"94dbb9085d79578046c4b10a0a7baefead738371ac763ee6ed7d727d348a2509",
		"bfcca7c164821bc6162657da43e816a78b5993ecaa8174a56a90788a83707749",
		"31f6f1e0b357c6c4b626e96b84a09289ccb0dd565a4d87831f7b5e49384fbe14",
		"50c5633909248612d848d4c8d353b57eb0802b8ae1bcea5d89b5cfe0134efa52",
		"1c502ae118d54d9b9f4c6bb47e9aff53e705c2eb2464c3403be445dc975d5f32",
		"b1c6d3a853d30624a771d0c29161ad32c7179d141d7d4b5e6534fc8b2514d8f6",
		"9284053f3a224697240b67a2450b8fe6bcaa07c61ab5f134f2dba41d76b506f4",
		"024bce5b3cd68652749937e9028e57ecac49ee3342fef15c76887d9bf144e51a",
		"d30976fcd0846a346c8de3a327170a2c96fb8a00972b03a84eedbfbc18b646d8",
		"2f06cf29ef2f787ac1a7b18cc89db27387dbb9c2a868914c5c506bee33e506f0",
		"6e5d936f2578d7805d1aa70f3e6d82492dfc5af838c2bbbe50fa2a343aab43e8",
	}
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main() {
	sendTx := flag.Bool("sendtx", true, "enable only send tx")
	senderCount := flag.Int("c", 4, "the number of sender")
	MonitorTx := flag.Bool("monitor", false, "enable monitor txs in block")
	startBlock := flag.Int("startBlock", 10, "calculate tps start from this block")

	url := flag.String("url", "ws://localhost:8546", "websocket url")
	flag.Parse()
	//fmt.Println(*sendTx, *startBlock, *url, *MonitorTx)

	ctx, cancel := context.WithCancel(context.Background())
	client, err := ethclient.Dial(*url)
	if err != nil {
		log.Fatalf(errPrefix+" connect %s: %v", *url, err)
	}

	block, err := client.BlockByNumber(ctx, nil)
	if err != nil {
		log.Println(errPrefix, err)
	}
	log.Println("block number: ", block.Number().String())

	if *sendTx {
		log.Printf("start send tx %d accounts", *senderCount)
		for i := 0; i < *senderCount; i++ {
			privateKey, err := crypto.HexToECDSA(sourceKey[i])
			if err != nil {
				log.Fatalf(errPrefix+" parse private key: %v", err)
			}
			publicKey := privateKey.Public()
			publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
			if !ok {
				log.Fatalf(errPrefix + " cannot assert type: publicKey is not of type *ecdsa.PublicKey")
			}
			fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
			log.Printf("sender:%d %s\n", i, fromAddress.String())

			go dummyTx(ctx, client, i, fromAddress)
		}
	}
	if *MonitorTx {
		log.Println("start monitor txs in blockChain")
		go calcTotalCount(ctx, client, *startBlock)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(interrupt)
	<-interrupt
	cancel()

	time.Sleep(time.Second)
	log.Println("dummy transaction exit")
}

func dummyTx(ctx context.Context, client *ethclient.Client, index int, fromAddr common.Address) {
	nonce, err := client.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		log.Fatalf(errPrefix+" get new nonce: %v", err)
	}
	value := big.NewInt(0)
	gasLimit := uint64(21000 + (20+64)*68) // in units
	gasPrice := big.NewInt(0)
	//gasPrice, err := client.SuggestGasPrice(ctx)
	//if err != nil {
	//	log.Fatalf(errPrefix+" get gas price: %v", err)
	//}
	toAddress := common.HexToAddress("0xffd79941b7085805f48ded97298694c6bb950e2c")

	var (
		data          [20 + 64]byte
		meterCount    = 0
		sendTxCal     = time.NewTimer(0)
		minuteTxCount = 0
	)

	<-sendTxCal.C
	sendTxCal.Reset(1 * time.Minute)
	copy(data[:], fromAddr.Bytes())

	start := time.Now()
	for {
		select {
		case <-ctx.Done():
			seconds := time.Since(start).Seconds()
			log.Printf("dummyTx:%v return (total %v in %v s, %v txs/s)", index, meterCount, seconds, float64(meterCount)/seconds)
			return
		case <-sendTxCal.C:
			sendTxCal.Reset(1 * time.Minute)
			log.Printf("%d th account, 1 minute send tx,total %d, send tps: %d txs/s  ", index, minuteTxCount, minuteTxCount/60)
			minuteTxCount = 0
			nonce, err = client.PendingNonceAt(ctx, fromAddr)
			if err != nil {
				log.Fatalf(errPrefix+" get new nonce: %v", err)
			}
			return
		default:
			//build,sign,send transaction
			_, _ = rand.Read(data[20:]) //read random 64 bytes
			dummy(ctx, nonce, toAddress, value, gasLimit, gasPrice, data[:], client, fromAddr)

			switch {
			case nonce%20000 == 0:
				nonce, err = client.PendingNonceAt(ctx, fromAddr)
				if err != nil {
					log.Fatalf(errPrefix+" get new nonce: %v", err)
				}
			default:
				nonce++
			}
			minuteTxCount++
			meterCount++
		}
	}
}

func dummy(ctx context.Context, nonce uint64, toAddress common.Address, value *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, client *ethclient.Client, fromAddress common.Address) {
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)
	err := client.SendTransaction(ctx, tx)
	if err != nil {
		log.Printf(warnPrefix+" send tx: %v", err)
	}

}

func calcTotalCount(ctx context.Context, client *ethclient.Client, startBlock int) {
	heads := make(chan *types.Header, 1)
	sub, err := client.SubscribeNewHead(context.Background(), heads)
	if err != nil {
		log.Fatal(errPrefix+"failed to subscribe to head events", "err", err)
	}
	defer sub.Unsubscribe()

	var (
		txsCount       uint
		minuteTxsCount uint
		finalCount     uint64
		timer          = time.NewTimer(0)
		start          = time.Now()
		minuteCount    = 0
	)

	<-timer.C
	timer.Reset(1 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			calcTotalCountExit(finalCount, time.Since(start).Seconds())
			return
		case <-timer.C:
			timer.Reset(1 * time.Minute)
			minuteCount++
			log.Printf("%d, 1min finalize %v txs, %v txs/s", minuteCount, minuteTxsCount, minuteTxsCount/60)

			if minuteCount == 10 {
				calcTotalCountExit(finalCount, time.Since(start).Seconds())
				//return
				finalCount = 0
				minuteCount = 0
			}
			minuteTxsCount = 0
		case head := <-heads:
			txsCount, err = client.TransactionCount(ctx, head.Hash())
			if err != nil {
				log.Printf(warnPrefix+"get txCount of block %v: %v", head.Hash(), err)
			}
			log.Printf("block Number: %s, txCount: %d", head.Number.String(), txsCount)
			minuteTxsCount += txsCount
			finalCount += uint64(txsCount)
		default:

		}
	}
}

func calcTotalCountExit(txsCount uint64, seconds float64) {
	//log.Println("calcTotalCount return")
	log.Printf("total finalize %v txs in %v seconds, %v txs/s", txsCount, seconds, float64(txsCount)/seconds)
}

func claimFunds(ctx context.Context, client *ethclient.Client, toAddress common.Address) {
	//0xffd79941b7085805f48ded97298694c6bb950e2c
	privateKey, err := crypto.HexToECDSA("04c070620a899a470a669fdbe0c6e1b663fd5bc953d9411eb36faa382005b3ad")
	if err != nil {
		log.Fatalf(errPrefix+" parse private key: %v", err)
	}
	fromAddress := common.HexToAddress("0xffd79941b7085805f48ded97298694c6bb950e2c")

	value := new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(10000)) // in wei (10000 eth)
	gasLimit := uint64(21000)                                                     // in units
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatalf(errPrefix+" get gas price: %v", err)
	}

	doClaim(ctx, client, fromAddress, toAddress, value, gasLimit, gasPrice, privateKey)
}

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
