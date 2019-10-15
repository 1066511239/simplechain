package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/simplechain-org/simplechain/cmd/hashDatabase/db"
	"github.com/simplechain-org/simplechain/common"
	"github.com/simplechain-org/simplechain/core/types"
	"github.com/simplechain-org/simplechain/ethclient"
)

const (
//websocketUrl = "ws://192.168.4.192:8546"
//websocketUrl = "ws://localhost:8546"
)

var fromAddress = common.HexToAddress("0xe673B8351E8B039F9e8d6Aca575ff36C1782fb17")

func main() {
	url := flag.String("url", "ws://localhost:8546", "websocket url")
	flag.Parse()
	fmt.Println(*url)
	client, err := ethclient.Dial(*url)
	if err != nil {
		log.Fatal(err)
	}

	hashDb, err := db.NewLDBDatabase("./IdHash", 0, 0)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go subScribeNewHead(client, hashDb, ctx)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(interrupt)
	<-interrupt
	cancel()

	fmt.Println("Got interrupt, shutting down...")
}

func subScribeNewHead(client *ethclient.Client, hashDb *db.LDBDatabase, ctx context.Context) {
	headerCh := make(chan *types.Header, 1)
	sub, err := client.SubscribeNewHead(ctx, headerCh)
	if err != nil {
		log.Fatal("subscribe NewHead error: ", err)
		return
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("subscribe newHead closed")
			return
		case err := <-sub.Err():
			fmt.Printf("subscribe newHead error:%s\n", err)
			return
		case header := <-headerCh:
			fmt.Printf("New Block,number: %s,hash: %s,", header.Number.String(), header.Hash().String())

			txsCount, err := client.TransactionCount(ctx, header.Hash())
			if err != nil {
				fmt.Println("GetTransactionCount error: ", err)
			}
			fmt.Printf(" txs:%d\n", txsCount)
		}
	}

}
