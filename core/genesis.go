// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/simplechain-org/simplechain/core/vm"
	"math/big"
	"strings"

	"github.com/simplechain-org/simplechain/common"
	"github.com/simplechain-org/simplechain/common/hexutil"
	"github.com/simplechain-org/simplechain/common/math"
	"github.com/simplechain-org/simplechain/core/rawdb"
	"github.com/simplechain-org/simplechain/core/state"
	"github.com/simplechain-org/simplechain/core/types"
	"github.com/simplechain-org/simplechain/ethdb"
	"github.com/simplechain-org/simplechain/log"
	"github.com/simplechain-org/simplechain/params"
	"github.com/simplechain-org/simplechain/rlp"
)

//go:generate gencodec -type Genesis -field-override genesisSpecMarshaling -out gen_genesis.go
//go:generate gencodec -type GenesisAccount -field-override genesisAccountMarshaling -out gen_genesis_account.go

var errGenesisNoConfig = errors.New("genesis has no chain configuration")
var (
	cesiContractAddr = common.HexToAddress("0xC0018822caC60FE4f223b758C83636546cF0D535")
	cesiOwner        = common.HexToAddress("0x51236b7c02956b698185b755b1f3a6a0d4bf20e5")
	cesiBin          = `608060405234801561001057600080fd5b50336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550610da1806100606000396000f3fe60806040526004361061007b5760003560e01c8063b89d212b1161004e578063b89d212b14610254578063c46a93d21461027f578063dd5ddfe814610347578063f3ae2415146103bf5761007b565b80632d06177a1461007d57806383c1cd8a146100ce5780639679976014610139578063a717b8481461018a575b005b34801561008957600080fd5b506100cc600480360360208110156100a057600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610428565b005b3480156100da57600080fd5b5061011d600480360360208110156100f157600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610667565b604051808260ff1660ff16815260200191505060405180910390f35b34801561014557600080fd5b506101886004803603602081101561015c57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506106bd565b005b34801561019657600080fd5b506101d9600480360360208110156101ad57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506108f3565b6040518080602001828103825283818151815260200191508051906020019080838360005b838110156102195780820151818401526020810190506101fe565b50505050905090810190601f1680156102465780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561026057600080fd5b506102696109d4565b6040518082815260200191505060405180910390f35b34801561028b57600080fd5b50610345600480360360208110156102a257600080fd5b81019080803590602001906401000000008111156102bf57600080fd5b8201836020820111156102d157600080fd5b803590602001918460018302840111640100000000831117156102f357600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050509192919290505050610a38565b005b34801561035357600080fd5b506103a36004803603604081101561036a57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803560ff169060200190929190505050610bc2565b604051808260ff1660ff16815260200191505060405180910390f35b3480156103cb57600080fd5b5061040e600480360360208110156103e257600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610c71565b604051808215151515815260200191505060405180910390f35b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146104ea576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252600a8152602001807f6f6e6c79204f776e65720000000000000000000000000000000000000000000081525060200191505060405180910390fd5b600160008273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff16156105aa576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260128152602001807f616c7265616479206973206d616e61676572000000000000000000000000000081525060200191505060405180910390fd5b60018060008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff0219169083151502179055507f8c2044b49cc7aa6c563c4ceb51d36108781595caa09e14748661afae62089db781604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390a150565b6000600260008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff169050919050565b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461077f576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252600a8152602001807f6f6e6c79204f776e65720000000000000000000000000000000000000000000081525060200191505060405180910390fd5b600160008273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff1661083e576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252600b8152602001807f6e6f74206d616e6167657200000000000000000000000000000000000000000081525060200191505060405180910390fd5b600160008273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81549060ff02191690557f71bb3750cf9c208f72f3507b50e3130adcb47f97a44ffdae1c4057809efcd64e81604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390a150565b6060600360008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000208054600181600116156101000203166002900480601f0160208091040260200160405190810160405280929190818152602001828054600181600116156101000203166002900480156109c85780601f1061099d576101008083540402835291602001916109c8565b820191906000526020600020905b8154815290600101906020018083116109ab57829003601f168201915b50505050509050919050565b60006003600260003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff1660ff1611610a3257600080fd5b43905090565b80816040516024018080602001828103825283818151815260200191508051906020019080838360005b83811015610a7d578082015181840152602081019050610a62565b50505050905090810190601f168015610aaa5780820380516001836020036101000a031916815260200191505b5092505050604051602081830303815290604052906040518082805190602001908083835b60208310610af25780518252602082019150602081019050602083039250610acf565b6001836020036101000a03801982511681845116808217855250505050505090500191505060405180910390207bffffffffffffffffffffffffffffffffffffffffffffffffffffffff19166020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff8381831617835250505050600360003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000209080519060200190610bbe929190610cc7565b5050565b600081600260008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff021916908360ff160217905550600260008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff16905092915050565b6000600160008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff169050919050565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f10610d0857805160ff1916838001178555610d36565b82800160010185558215610d36579182015b82811115610d35578251825591602001919060010190610d1a565b5b509050610d439190610d47565b5090565b610d6991905b80821115610d65576000816000905550600101610d4d565b5090565b9056fea265627a7a72315820a074ec20d6df6a13f04c7a8c5b736dabe76625459f2f72d7ee576a93df6bb0f764736f6c634300050b0032`
)

// Genesis specifies the header fields, state of a genesis block. It also defines hard
// fork switch-over blocks through the chain configuration.
type Genesis struct {
	Config     *params.ChainConfig `json:"config"`
	Nonce      uint64              `json:"nonce"`
	Timestamp  uint64              `json:"timestamp"`
	ExtraData  []byte              `json:"extraData"`
	GasLimit   uint64              `json:"gasLimit"   gencodec:"required"`
	Difficulty *big.Int            `json:"difficulty" gencodec:"required"`
	Mixhash    common.Hash         `json:"mixHash"`
	Coinbase   common.Address      `json:"coinbase"`
	Alloc      GenesisAlloc        `json:"alloc"      gencodec:"required"`

	// These fields are used for consensus tests. Please don't use them
	// in actual genesis blocks.
	Number     uint64      `json:"number"`
	GasUsed    uint64      `json:"gasUsed"`
	ParentHash common.Hash `json:"parentHash"`
}

// GenesisAlloc specifies the initial state that is part of the genesis block.
type GenesisAlloc map[common.Address]GenesisAccount

func (ga *GenesisAlloc) UnmarshalJSON(data []byte) error {
	m := make(map[common.UnprefixedAddress]GenesisAccount)
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	*ga = make(GenesisAlloc)
	for addr, a := range m {
		(*ga)[common.Address(addr)] = a
	}
	return nil
}

// GenesisAccount is an account in the state of the genesis block.
type GenesisAccount struct {
	Code       []byte                      `json:"code,omitempty"`
	Storage    map[common.Hash]common.Hash `json:"storage,omitempty"`
	Balance    *big.Int                    `json:"balance" gencodec:"required"`
	Nonce      uint64                      `json:"nonce,omitempty"`
	PrivateKey []byte                      `json:"secretKey,omitempty"` // for tests
}

// field type overrides for gencodec
type genesisSpecMarshaling struct {
	Nonce      math.HexOrDecimal64
	Timestamp  math.HexOrDecimal64
	ExtraData  hexutil.Bytes
	GasLimit   math.HexOrDecimal64
	GasUsed    math.HexOrDecimal64
	Number     math.HexOrDecimal64
	Difficulty *math.HexOrDecimal256
	Alloc      map[common.UnprefixedAddress]GenesisAccount
}

type genesisAccountMarshaling struct {
	Code       hexutil.Bytes
	Balance    *math.HexOrDecimal256
	Nonce      math.HexOrDecimal64
	Storage    map[storageJSON]storageJSON
	PrivateKey hexutil.Bytes
}

// storageJSON represents a 256 bit byte array, but allows less than 256 bits when
// unmarshaling from hex.
type storageJSON common.Hash

func (h *storageJSON) UnmarshalText(text []byte) error {
	text = bytes.TrimPrefix(text, []byte("0x"))
	if len(text) > 64 {
		return fmt.Errorf("too many hex characters in storage key/value %q", text)
	}
	offset := len(h) - len(text)/2 // pad on the left
	if _, err := hex.Decode(h[offset:], text); err != nil {
		fmt.Println(err)
		return fmt.Errorf("invalid hex storage key/value %q", text)
	}
	return nil
}

func (h storageJSON) MarshalText() ([]byte, error) {
	return hexutil.Bytes(h[:]).MarshalText()
}

// GenesisMismatchError is raised when trying to overwrite an existing
// genesis block with an incompatible one.
type GenesisMismatchError struct {
	Stored, New common.Hash
}

func (e *GenesisMismatchError) Error() string {
	return fmt.Sprintf("database already contains an incompatible genesis block (have %x, new %x)", e.Stored[:8], e.New[:8])
}

// SetupGenesisBlock writes or updates the genesis block in db.
// The block that will be used is:
//
//                          genesis == nil       genesis != nil
//                       +------------------------------------------
//     db has no genesis |  main-net default  |  genesis
//     db has genesis    |  from DB           |  genesis (if compatible)
//
// The stored chain configuration will be updated if it is compatible (i.e. does not
// specify a fork block below the local head block). In case of a conflict, the
// error is a *params.ConfigCompatError and the new, unwritten config is returned.
//
// The returned chain configuration is never nil.
func SetupGenesisBlock(db ethdb.Database, genesis *Genesis) (*params.ChainConfig, common.Hash, error) {
	return SetupGenesisBlockWithOverride(db, genesis, nil)
}
func SetupGenesisBlockWithOverride(db ethdb.Database, genesis *Genesis, constantinopleOverride *big.Int) (*params.ChainConfig, common.Hash, error) {
	if genesis != nil && genesis.Config == nil {
		return params.AllScryptProtocolChanges, common.Hash{}, errGenesisNoConfig
	}

	// Just commit the new block if there is no stored genesis block.
	stored := rawdb.ReadCanonicalHash(db, 0)
	if (stored == common.Hash{}) {
		if genesis == nil {
			log.Info("Writing default main-net genesis block")
			genesis = DefaultGenesisBlock()
		} else {
			log.Info("Writing custom genesis block")
		}
		block, err := genesis.Commit(db)
		return genesis.Config, block.Hash(), err
	}

	// Check whether the genesis block is already written.
	if genesis != nil {
		hash := genesis.ToBlock(nil).Hash()
		if hash != stored {
			return genesis.Config, hash, &GenesisMismatchError{stored, hash}
		}
	}

	// Get the existing chain configuration.
	newcfg := genesis.configOrDefault(stored)
	if constantinopleOverride != nil {
		newcfg.ConstantinopleBlock = constantinopleOverride
		newcfg.PetersburgBlock = constantinopleOverride
	}
	storedcfg := rawdb.ReadChainConfig(db, stored)
	if storedcfg == nil {
		log.Warn("Found genesis block without chain config")
		rawdb.WriteChainConfig(db, stored, newcfg)
		return newcfg, stored, nil
	}
	// Special case: don't change the existing config of a non-mainnet chain if no new
	// config is supplied. These chains would get AllProtocolChanges (and a compat error)
	// if we just continued here.
	if genesis == nil && stored != params.MainnetGenesisHash {
		return storedcfg, stored, nil
	}

	// Check config compatibility and write the config. Compatibility errors
	// are returned to the caller unless we're already at block zero.
	height := rawdb.ReadHeaderNumber(db, rawdb.ReadHeadHeaderHash(db))
	if height == nil {
		return newcfg, stored, fmt.Errorf("missing block number for head header hash")
	}
	compatErr := storedcfg.CheckCompatible(newcfg, *height)
	if compatErr != nil && *height != 0 && compatErr.RewindTo != 0 {
		return newcfg, stored, compatErr
	}
	rawdb.WriteChainConfig(db, stored, newcfg)
	return newcfg, stored, nil
}

func (g *Genesis) configOrDefault(ghash common.Hash) *params.ChainConfig {
	switch {
	case g != nil:
		return g.Config
	case ghash == params.MainnetGenesisHash:
		return params.MainnetChainConfig
	case ghash == params.TestnetGenesisHash:
		return params.TestnetChainConfig
	default:
		return params.AllScryptProtocolChanges
	}
}

// ToBlock creates the genesis block and writes state of a genesis specification
// to the given database (or discards it if nil).
func (g *Genesis) ToBlock(db ethdb.Database) *types.Block {
	if db == nil {
		db = ethdb.NewMemDatabase()
	}
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(db))
	for addr, account := range g.Alloc {
		if len(account.Code) > 4 && bytes.Equal(account.Code[:4], []byte{0x60, 0x80, 0x60, 0x40}) {
			//create contract
			context := vm.Context{
				CanTransfer: CanTransfer,
				Transfer:    Transfer,
				GetHash:     func(uint64) common.Hash { return common.Hash{} },
				Origin:      addr,
				Coinbase:    g.Coinbase,
				BlockNumber: big.NewInt(0),
				Time:        big.NewInt(int64(g.Timestamp)),
				Difficulty:  g.Difficulty,
				GasLimit:    g.GasLimit,
				GasPrice:    big.NewInt(0),
			}
			evm := vm.NewEVM(context, statedb, g.Config, vm.Config{})
			res, contractAddr, returnGas, suberr := evm.Create(vm.AccountRef(addr), account.Code, g.GasLimit, big.NewInt(0))
			if suberr != nil {
				log.Error("contract create failed", "error info", suberr.Error())
				return nil
			}
			if account.Balance != nil {
				statedb.AddBalance(contractAddr, account.Balance)
			}
			log.Info("contract info", "address", contractAddr, "return gas", returnGas, "contract code", fmt.Sprintf("%x", res))
		} else {
			if account.Balance != nil {
				statedb.AddBalance(addr, account.Balance)
			}
			statedb.SetCode(addr, account.Code)
			statedb.SetNonce(addr, account.Nonce)
		}
		for key, value := range account.Storage {
			statedb.SetState(addr, key, value)
		}
	}
	root := statedb.IntermediateRoot(false)
	head := &types.Header{
		Number:     new(big.Int).SetUint64(g.Number),
		Nonce:      types.EncodeNonce(g.Nonce),
		Time:       g.Timestamp,
		ParentHash: g.ParentHash,
		Extra:      g.ExtraData,
		GasLimit:   g.GasLimit,
		GasUsed:    g.GasUsed,
		Difficulty: g.Difficulty,
		MixDigest:  g.Mixhash,
		Coinbase:   g.Coinbase,
		Root:       root,
	}
	if g.GasLimit == 0 {
		head.GasLimit = params.GenesisGasLimit
	}
	if g.Difficulty == nil {
		head.Difficulty = params.GenesisDifficulty
	}
	statedb.Commit(false)
	statedb.Database().TrieDB().Commit(root, true)

	return types.NewBlock(head, nil, nil, nil)
}

// Commit writes the block and state of a genesis specification to the database.
// The block is committed as the canonical head block.
func (g *Genesis) Commit(db ethdb.Database) (*types.Block, error) {
	block := g.ToBlock(db)
	if block.Number().Sign() != 0 {
		return nil, fmt.Errorf("can't commit genesis block with number > 0")
	}
	rawdb.WriteTd(db, block.Hash(), block.NumberU64(), g.Difficulty)
	rawdb.WriteBlock(db, block)
	rawdb.WriteReceipts(db, block.Hash(), block.NumberU64(), nil)
	rawdb.WriteCanonicalHash(db, block.Hash(), block.NumberU64())
	rawdb.WriteHeadBlockHash(db, block.Hash())
	rawdb.WriteHeadHeaderHash(db, block.Hash())

	config := g.Config
	if config == nil {
		config = params.AllScryptProtocolChanges
	}
	rawdb.WriteChainConfig(db, block.Hash(), config)
	return block, nil
}

// MustCommit writes the genesis block and state to db, panicking on error.
// The block is committed as the canonical head block.
func (g *Genesis) MustCommit(db ethdb.Database) *types.Block {
	block, err := g.Commit(db)
	if err != nil {
		panic(err)
	}
	return block
}

// GenesisBlockForTesting creates and writes a block in which addr has the given wei balance.
func GenesisBlockForTesting(db ethdb.Database, addr common.Address, balance *big.Int) *types.Block {
	g := Genesis{Alloc: GenesisAlloc{addr: {Balance: balance}}}
	return g.MustCommit(db)
}

// DefaultGenesisBlock returns the Ethereum main net genesis block.
func DefaultGenesisBlock() *Genesis {
	defaultGenesisAlloc := map[common.Address]GenesisAccount{
		cesiOwner: {Code: common.FromHex(cesiBin)},
	}
	return &Genesis{
		Config:     params.MainnetChainConfig,
		Timestamp:  1546531200,
		Nonce:      312,
		ExtraData:  hexutil.MustDecode(SipeGenesisBlockExtraData),
		GasLimit:   800000,
		Difficulty: big.NewInt(1200),
		Alloc:      defaultGenesisAlloc,
	}
}

// DefaultTestnetGenesisBlock returns the SimpleChain network genesis block.
func DefaultTestnetGenesisBlock() *Genesis {
	return &Genesis{
		Config:     params.TestnetChainConfig,
		Nonce:      312,
		ExtraData:  hexutil.MustDecode("0x73696d706c65636861696e20746573746e6574"),
		GasLimit:   16777216,
		Difficulty: big.NewInt(120000),
		Alloc:      decodePrealloc(SipeTestnetAllocData),
	}
}

// DefaultRinkebyGenesisBlock returns the Rinkeby network genesis block, only for genesis template
func DefaultRinkebyGenesisBlock() *Genesis {
	return &Genesis{
		Config:     params.RinkebyChainConfig,
		Timestamp:  1492009146,
		ExtraData:  hexutil.MustDecode("0x52657370656374206d7920617574686f7269746168207e452e436172746d616e42eb768f2244c8811c63729a21a3569731535f067ffc57839b00206d1ad20c69a1981b489f772031b279182d99e65703f0076e4812653aab85fca0f00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
		GasLimit:   4700000,
		Difficulty: big.NewInt(1),
		Alloc:      decodePrealloc(""),
	}
}

// DefaultEthashGenesisBlock returns the ethereum testnet network genesis block.
func DefaultEthashGenesisBlock() *Genesis {
	return &Genesis{
		Config:     params.TestnetChainConfig,
		Nonce:      66,
		ExtraData:  hexutil.MustDecode("0x3535353535353535353535353535353535353535353535353535353535353535"),
		GasLimit:   16777216,
		Difficulty: big.NewInt(1048576),
		Alloc:      decodePrealloc(EthTestnetAllocData),
	}
}

// DeveloperGenesisBlock returns the 'sipe --dev' genesis block. Note, this must
// be seeded with the
func DeveloperGenesisBlock(period uint64, faucet common.Address) *Genesis {
	// Override the default period to the user requested one
	config := *params.AllCliqueProtocolChanges
	config.Clique.Period = period

	// Assemble and return the genesis with the precompiles and faucet pre-funded
	return &Genesis{
		Config:     &config,
		ExtraData:  append(append(make([]byte, 32), faucet[:]...), make([]byte, 65)...),
		GasLimit:   6283185,
		Difficulty: big.NewInt(1),
		Alloc: map[common.Address]GenesisAccount{
			common.BytesToAddress([]byte{1}): {Balance: big.NewInt(1)}, // ECRecover
			common.BytesToAddress([]byte{2}): {Balance: big.NewInt(1)}, // SHA256
			common.BytesToAddress([]byte{3}): {Balance: big.NewInt(1)}, // RIPEMD
			common.BytesToAddress([]byte{4}): {Balance: big.NewInt(1)}, // Identity
			common.BytesToAddress([]byte{5}): {Balance: big.NewInt(1)}, // ModExp
			common.BytesToAddress([]byte{6}): {Balance: big.NewInt(1)}, // ECAdd
			common.BytesToAddress([]byte{7}): {Balance: big.NewInt(1)}, // ECScalarMul
			common.BytesToAddress([]byte{8}): {Balance: big.NewInt(1)}, // ECPairing
			faucet:                           {Balance: new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))},
		},
	}
}

func decodePrealloc(data string) GenesisAlloc {
	var p []struct{ Addr, Balance *big.Int }
	if err := rlp.NewStream(strings.NewReader(data), 0).Decode(&p); err != nil {
		panic(err)
	}
	ga := make(GenesisAlloc, len(p))
	for _, account := range p {
		ga[common.BigToAddress(account.Addr)] = GenesisAccount{Balance: account.Balance}
	}
	return ga
}
