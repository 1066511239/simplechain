package main

import (
	"context"
	"crypto/ecdsa"
	"flag"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/simplechain-org/simplechain/cmd/dummytx/db"
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
		//
		//"bfcca7c164821bc6162657da43e816a78b5993ecaa8174a56a90788a83707749",
		//"31f6f1e0b357c6c4b626e96b84a09289ccb0dd565a4d87831f7b5e49384fbe14",
		//"50c5633909248612d848d4c8d353b57eb0802b8ae1bcea5d89b5cfe0134efa52",
		//"1c502ae118d54d9b9f4c6bb47e9aff53e705c2eb2464c3403be445dc975d5f32",
		//"b1c6d3a853d30624a771d0c29161ad32c7179d141d7d4b5e6534fc8b2514d8f6",
		//"9284053f3a224697240b67a2450b8fe6bcaa07c61ab5f134f2dba41d76b506f4",
		//"024bce5b3cd68652749937e9028e57ecac49ee3342fef15c76887d9bf144e51a",
		//"d30976fcd0846a346c8de3a327170a2c96fb8a00972b03a84eedbfbc18b646d8",
		//"2f06cf29ef2f787ac1a7b18cc89db27387dbb9c2a868914c5c506bee33e506f0",
		//"6e5d936f2578d7805d1aa70f3e6d82492dfc5af838c2bbbe50fa2a343aab43e8",
		//"a0a3d33afe86bd99827604caade4a2418b5e1c0f4213626763c8a4904ada5552",
		//"7c7f9280fa44a190f4bafa8f75f1a272f18a8766f9fab4e01c68f190db6ff943",
		//"50af8f1078fe4bd4f4a16fcddcc19a73fc15ef74edae4df7512fe72dc44e8636",
		//"c7faed5ef7fecd406ad9b362d09b0b7d5abccd4f4e308c572417c8340e159f6e",
		//"66e0318c674ccfe3b6ed7c9177e8a73a6d6e3377507bc85a31ae10539b2f760e",
		//"359137e052174a1729181798e4ffc53be5232153a82851ca0d4a67f891cf7cfd",
		//"58db0bb015886e544e742fdfca3d50c5347a8a79f98b150adea9f4ac6a89ce5d",
		//"9c277eed4578e41880c830363a07ab27df757b91b85a10a2446bc0ebcf8141ad",
		//"724bf5ebe6058b9be382c42d3c906987dbccd557b36a3b7f7416f25efdffe30e",
		//"81fabd0982b836ca7f853a44900d150593243c7105c6c87489f4436fbe047329",
		//"ce542dc71c9b0316a7644e4abf5987b8afa9ba49300931ab65d1d3fd60267cc9",
		//"f51eb310a91fd28d67f9866d616a9c4d8cefafbfde64afbc06e0505550b606e2",
		//"e5ac84fca4a2219e05f577084c1361cf25c2cf45ba07eee135bf4a7f61c3c9c5",
		//"c99d1c92b1e8a4111d7a0766eb2b05ce827d9a75481de3159ffc40fde585defb",
		//"3ce36df955dbab881213b4aac5412a83f75ca012f257f6caaf73566101974f36",
		//"ace0bbf5540b9074168d77ccece1779a73511b23c9864f1e0d000c4e18a69c42",
		//"a6ecd58fc6c1ffb03e54f21218c713b6aef9dbb61212e1d8b8de521846e99c71",
		//"ad36f5efd748dc7895366d028badc5bfcac3355f3d1151760a50e7e13b4502be",
		//"73ca1ecefd17f0aaccd665ab1d2765fe44f272eb99744cf8128fd21547d7a9a9",
		//"14e7b973f57a7707ac10a6e26892f5b3041cb8f735d35edf47f7d23d067baeb5",
		//"2e2968d7a8e6482654ace269e333b2d0dbabc4e594b8fc594ba3be636f264c5d",
		//"d5d5d6ec9c08a09e2d45d539e724bb2181c8c25586f0ab4fe08285b943cf061b",
		//"0a421bb214d61aa943328993d785ccc15ad049cc31fde87bc06232547046c33b",
		//"3b3948e44e46d66587df14cf101570242bc24a568063ec7b046ffc8ed93bc2cd",
		//"4f9f72b9c4a688e3a6c4268627b6188704071d4efc919b23a45ad13ce9a56b17",
		//"7111f9be4a79363bbc242ab83a9d73b3fe8f072443463acf62c882aeb6633aaf",
		//"4bd077b5d373a40deabe0beed7f7eaf78a43fb582117af273e367f857624cdf7",
		//"c299e149fc6445772f30e7a3e183dbff2369cbe10a5f7bc685aaf06ec7fab8b7",
		//"faf00b3ee6162e6f126e59ab823f9fb847c7a821ae9f850a296cde4067ba91f0",
		//"a29199156d5e404ce1b8468cd5fbc8859251a9cbaedf6877d343cf9916adfac0",
		//"60091525639cd74720f2095d72196b2e5b222c7b7930fe4d2b1225cbd097488e",
		//"23f0b7099b105abbb3ac694068752b4210cbadb89397b29a66d983ee4ffda9b5",
		//"e0a48bd0478379c420e1c3c6cab0cc224dc6cdaeb488a10d87852e2d0d33875e",
		//"9b0ca2cf966db34535d160392f8f8773f78b3a9f58b0c8b5ad376eff9db7b5e1",
		//"e04bc7cc3d04f00a5dc96a0843e4ad63d0bc5d25977768911b1e5bf990607d94",
		//"2f9e2da9d4fef8f79f9fafc5e173c7ab175152137eec8ec9d3390e152f04c6e8",
		//"a552515da13ea1f8810d4a38c34dbf057f227954d307b1ef0b84d548f755703e",
		//"a742b92fce506bb28a1c33a3dce2eadf529062e37ad27997492ab9dac7ec66fb",
		//"cb7443ba9adec3554aaa4b9de53bd679b7351ad54915409550fbd9f31f2da876",
		//"4a904d7d5dc78bcd7b582cf5ce7b03b6dce92772c353b2936564a6924bbe5a57",
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

	PrintDB := flag.Bool("printDB", false, "get hashID from DB")
	getTxByHashData := flag.String("hashID", "", "get tx ID by hash data")
	flag.Parse()

	hashDb, err := db.NewLDBDatabase("./IdHash", 0, 0)
	if err != nil {
		log.Fatal(err)
	}
	if *PrintDB {
		itr := hashDb.NewIterator()
		itr.First()

		for itr.Valid() {
			fmt.Println("hash data: ", common.BytesToHash(itr.Key()).String(), "tx ID: ", common.BytesToHash(itr.Value()).String())
			itr.Next()
		}
		itr.Release()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	client, err := ethclient.Dial(*url)
	if err != nil {
		log.Fatalf(errPrefix+" connect %s: %v", *url, err)
	}

	if len(*getTxByHashData) == 66 {
		txID, err := hashDb.GetHashId(common.HexToHash(*getTxByHashData))
		if err != nil {
			log.Printf("getHashId failed,hash:%s, err: %s", *getTxByHashData, err)
		}
		log.Println("txID: ", txID.String())
		tx, _, err := client.TransactionByHash(ctx, txID)
		log.Println("hash data: ", common.BytesToHash(tx.Data()[20:]).String())
		txJson, err := tx.MarshalJSON()
		log.Println("tx:  ", string(txJson))
		return
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

			go dummyTx(ctx, client, i, fromAddress, hashDb)
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

func dummyTx(ctx context.Context, client *ethclient.Client, index int, fromAddr common.Address, hashDb *db.LDBDatabase) {
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
			dummy(ctx, nonce, toAddress, value, gasLimit, gasPrice, data[:], client, fromAddr, hashDb)

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

func dummy(ctx context.Context, nonce uint64, toAddress common.Address, value *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, client *ethclient.Client, fromAddress common.Address, hashDb *db.LDBDatabase) {
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)
	err := client.SendTransaction(ctx, tx)
	if err != nil {
		log.Printf(warnPrefix+" send tx: %v", err)
	}

	err = hashDb.InsertHash(common.BytesToHash(tx.Data()[20:]), tx.Hash())
	if err != nil {
		fmt.Printf("InsertHash failed: hash: %s,txId:%s\n", common.BytesToHash(tx.Data()[20:]).String(), tx.Hash().String())
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
