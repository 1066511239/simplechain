.PHONY: build cluster test

default: build

build:
	go build -o bin/swarm    github.com/simplechain-org/simplechain/cmd/swarm
	go build -o bin/geth     github.com/simplechain-org/simplechain/cmd/geth
	go build -o bin/bootnode github.com/simplechain-org/simplechain/cmd/bootnode

cluster: build
	scripts/boot-cluster.sh

test:
	go test -v github.com/simplechain-org/simplechain/swarm/...
