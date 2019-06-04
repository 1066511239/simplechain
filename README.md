## Go SimpleChain (Work In Process)

[![Build Status](https://travis-ci.org/simplechain-org/simplechain.svg?branch=dev)](https://travis-ci.org/simplechain-org/simplechain)
[![GoDoc](https://godoc.org/github.com/simplechain-org/simplechain?status.svg)](https://godoc.org/github.com/simplechain-org/simplechain)
[![Go Report Card](https://goreportcard.com/badge/github.com/simplechain-org/simplechain)](https://goreportcard.com/report/github.com/simplechain-org/simplechain)

New implementation of SimpleChain base on geth v1.8.27. SimpleChain use scrypt algorithm instead of
ethash for POW mining.

## Building the source (just for developer)

Building sipe which is the simplechain client requires both a Go (version 1.9 or later) and a C compiler.
You can install them using your favourite package manager.
Once the dependencies are installed, run
```bash
$ make sipe
```

or, to build the full suite of utilities:

```bash
$ make all
```

for developers, after your code completed , please run lint tools and fix lint errors.:

```bash
$ make lint
```

#### Development flow:
checkout one branch -> complete your code -> make lint and make test -> git commit and git push -> make a pull request -> code rewiew and approve it -> merge dev -> release master 

#### Defining the private genesis state (develop mode)

First, you'll need to create the genesis state of your networks, which all nodes need to be aware of
and agree upon. This consists of a small JSON file (e.g. call it `genesis.json`):

```json
{
  "config": {
    "chainId": 29169,
    "homesteadBlock": 1,
    "eip150Block": 2,
    "eip150Hash": "0x0000000000000000000000000000000000000000000000000000000000000000",
    "eip155Block": 3,
    "eip158Block": 3,
    "byzantiumBlock": 4,
    "constantinopleBlock": 5,
    "scrypt": {}
  },
  "nonce": "0x0",
  "timestamp": "0x5cd907f1",
  "extraData": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "gasLimit": "0x47b760",
  "difficulty": "0x80",
  "mixHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "coinbase": "0x0000000000000000000000000000000000000000",
  "alloc": {
    "0000000000000000000000000000000000000000": {
      "balance": "0x1"
    }
  },
  "number": "0x0",
  "gasUsed": "0x0",
  "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000"
}
```

The above fields should be fine for most purposes, although we'd recommend changing the `nonce` to
some random value so you prevent unknown remote nodes from being able to connect to you. If you'd
like to pre-fund some accounts for easier testing, you can populate the `alloc` field with account
configs:

```json
"alloc": {
  "0x0000000000000000000000000000000000000001": {"balance": "111111111"},
  "0xae1c5cce0d1b444785860122f6d7c66f44440c76": {"balance": "222222222"}
}
```

With the genesis state defined in the above JSON file, you'll need to initialize **every** Sipe node
with it prior to starting it up to ensure all blockchain parameters are correctly set:

you can run cmd/puppeth to generate the genesis file

with your genesis file, you should run `sipe init`
```
$ ./build/bin/sipe init path/to/genesis.json --datadir path/for/your/dir
```

#### Running a private miner

```
$ ./build/bin/sipe --datadir path/for/your/dir --etherbase 0xae1c5cce0d1b444785860122f6d7c66f44440c76 --mine --minerthreads 1
```

Which will start mining blocks and transactions on a single CPU thread, crediting all proceedings to
the account specified by `--etherbase`. You can further tune the mining by changing the default gas
limit blocks converge to (`--targetgaslimit`) and the price transactions are accepted at (`--gasprice`).

## Contribution
Please make sure your contributions adhere to our coding guidelines:

 * Code must adhere to the official Go [formatting](https://golang.org/doc/effective_go.html#formatting) guidelines (i.e. uses [gofmt](https://golang.org/cmd/gofmt/)).
 * Code must be documented adhering to the official Go [commentary](https://golang.org/doc/effective_go.html#commentary) guidelines.
 * Pull requests need to be based on and opened against the `master` branch.
 * Commit messages should be prefixed with the package(s) they modify.
   * E.g. "eth, rpc: make trace configs optional"

Please see the [Developers' Guide](https://github.com/simplechain-org/simplechain/wiki/Developers'-Guide)
for more details on configuring your environment, managing project dependencies and testing procedures.

