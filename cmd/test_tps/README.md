##### 5s POA genesis.json

- gaslimt=1198915584, init count= 0xffd79941b7085805f48ded97298694c6bb950e2c
```json
{
  "config": {
    "chainId": 110,
    "homesteadBlock": 1,
    "eip150Block": 2,
    "eip150Hash": "0x0000000000000000000000000000000000000000000000000000000000000000",
    "eip155Block": 3,
    "eip158Block": 3,
    "byzantiumBlock": 4,
    "constantinopleBlock": 5,
    "clique": {
      "period":5, 
      "epoch": 30000
    }
  },
  "nonce": "0x0",
  "timestamp": "0x5d8993f8",
  "extraData": "0x0000000000000000000000000000000000000000000000000000000000000000ffd79941b7085805f48ded97298694c6bb950e2c0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
  "gasLimit": "0x47b760000",
  "difficulty": "0x1",
  "mixHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "coinbase": "0x0000000000000000000000000000000000000000",
  "alloc": {
    "ffd79941b7085805f48ded97298694c6bb950e2c": {
      "balance": "0x200000000000000000000000000000000000000000000000000000000000000"
    }
  },
  "number": "0x0",
  "gasUsed": "0x0",
  "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000"
}

```

##### 导入0xffd79941b7085805f48ded97298694c6bb950e2c私钥

```bash
 $ mkdir data
 $ echo 04c070620a899a470a669fdbe0c6e1b663fd5bc953d9411eb36faa382005b3ad > privkey
 $ echo 111111 > password
 $ sipe account import ./privkey --datadir=kk --password ./password --datadir ./data/

```
##### sipe

- 初始化genesis block
```bash
 $ sipe init ./poa.json --datadir data/
```

- 启动sipe
```bash
 $ sipe --datadir ./data --rpc --rpcvhosts "*" --rpcaddr 0.0.0.0 --rpcport 8545 --rpccorsdomain "*" --rpcapi "db,eth,net,web3,personal,debug" --ws --wsaddr 0.0.0.0 --wsport 8546 --wsapi "db,eth,net,web3,personal,debug" --unlock 0xffd79941b7085805f48ded97298694c6bb950e2c --password <(echo 111111) --mine --txpool.globalslots=40960
```

##### dummytx

- `dummytx`: 普通转账交易,默认是12个账户同时转账给0xffd79941b7085805f48ded97298694c6bb950e2c
- `dummytx xx`:带有32字节hash值的转账交易,`dummytx 4`和`dummytx 8`是选4或8个账户同时转账给0xffd79941b7085805f48ded97298694c6bb950e2c

如果账户没足够的sipc, 会从0xffd79941b7085805f48ded97298694c6bb950e2c转出10000sipc给该账户