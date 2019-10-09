package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/simplechain-org/simplechain/cmd/hashDatabase/db"
	"github.com/simplechain-org/simplechain/common"
	"github.com/simplechain-org/simplechain/core/types"
	"github.com/simplechain-org/simplechain/ethclient"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	websocketUrl = "ws://127.0.0.1:8546"
)

var fromAddress = common.HexToAddress("0xe673B8351E8B039F9e8d6Aca575ff36C1782fb17")

func main() {
	client, err := ethclient.Dial(websocketUrl)
	if err != nil {
		log.Fatal(err)
	}

	hashDb, err := db.NewLDBDatabase("./IdHash", 0, 0)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	subScribeNewHead(client, hashDb, ctx)

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
			fmt.Printf("New Block,number: %s,hash: %s\n", header.Number.String(), header.Hash().String())

			block, err := client.BlockByNumber(ctx, header.Number)
			if err != nil {
				fmt.Printf("get block %s failed: %s\n", header.Number.String(), err)
				break
			}
			txs := block.Transactions()
			if len(txs) > 0 {
				fmt.Println("count of tx", len(txs))
				for _, tx := range txs {
					data := tx.Data()
					if bytes.HasPrefix(data, fromAddress.Bytes()) {
						fmt.Printf("id:%s, hash：%s\n", tx.Hash().String(), common.BytesToHash(data[20:]).String())
						if err := hashDb.InsertHash(common.BytesToHash(data[2:]), tx.Hash()); err != nil {
							fmt.Printf("InsertDb failed: id:%s,hash:%s\n", tx.Hash().String(), common.BytesToHash(data[2:]).String())
						}
					}
				}
			}

		}
	}

}
