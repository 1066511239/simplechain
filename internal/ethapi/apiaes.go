package ethapi

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"github.com/simplechain-org/simplechain/accounts"
	"github.com/simplechain-org/simplechain/common"
	"github.com/simplechain-org/simplechain/common/hexutil"
	"github.com/simplechain-org/simplechain/crypto"
	"github.com/simplechain-org/simplechain/log"
	"github.com/simplechain-org/simplechain/rpc"
	"math/big"
	"time"
)

func (s *PublicTransactionPoolAPI) SaveDataWithKey(ctx context.Context, data string, key string, args SendTxArgs) (common.Hash, error) {
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: args.From}

	wallet, err := s.b.AccountManager().Find(account)
	if err != nil {
		return common.Hash{}, err
	}

	args.To = &managerContractAddress

	if args.Input == nil {
		methodID := crypto.Keccak256([]byte("saveEncryptedData(bytes)"))
		if len(key) != 16 {
			return common.Hash{}, fmt.Errorf("the length of key must 16")
		}
		dataIn, err := AesEncrypt(data, key)
		if err != nil {
			return common.Hash{}, err
		}
		var input []byte
		input = append(input, methodID[:4]...)
		input = append(input, common.LeftPadBytes(new(big.Int).SetUint64(32).Bytes(), 32)...)
		input = append(input, common.LeftPadBytes(new(big.Int).SetUint64(uint64(len(dataIn))).Bytes(), 32)...)
		l := len(dataIn)
		if l%32 > 0 {
			l = (l/32 + 1) * 32
		}
		input = append(input, common.RightPadBytes([]byte(dataIn), l)...)
		j := hexutil.Bytes([]byte(input))
		args.Input = &j
	}

	if args.Nonce == nil {
		// Hold the addresse's mutex around signing to prevent concurrent assignment of
		// the same nonce to multiple accounts.
		s.nonceLock.LockAddr(args.From)
		defer s.nonceLock.UnlockAddr(args.From)
	}

	// Set some sanity defaults and terminate on failure
	if err := args.setDefaults(ctx, s.b); err != nil {
		return common.Hash{}, err
	}
	// Assemble the transaction and sign with the wallet
	tx := args.toTransaction()

	var chainID *big.Int
	if config := s.b.ChainConfig(); config.IsEIP155(s.b.CurrentBlock().Number()) {
		chainID = config.ChainID
	}
	signed, err := wallet.SignTx(account, tx, chainID)
	if err != nil {
		return common.Hash{}, err
	}
	return submitTransaction(ctx, s.b, signed)
}

func (s *PublicBlockChainAPI) GetDataWithKey(ctx context.Context, from common.Address, key string) (string, error) {
	data, err := getEncryptedData()
	if err != nil {
		return "", err
	}
	args := buildCallArgs(data)
	args.From = from
	result, gas, failed, err := s.doCall(ctx, *args, rpc.PendingBlockNumber, 5*time.Second, s.b.RPCGasCap())
	log.Info("getEncryptedData", "result", result, "gas", gas, "failed", failed, "error", err)
	if err != nil {
		return "", err
	}

	l := binary.BigEndian.Uint64(result[56:64])
	reData := result[64 : 64+l]
	if len(key) != 16 {
		return "", fmt.Errorf("the length of key must 16")
	}
	getString := AesDecrypt(string(reData), key)

	return getString, nil
}
func getEncryptedData() (hexutil.Bytes, error) {
	fnId, err := getFnId("getEncryptedData()")
	if err != nil {
		return nil, err
	}

	var data []byte
	data = append(data, fnId...)
	return data, nil
}

func AesEncrypt(orig string, key string) (string, error) {
	// 转成字节数组
	origData := []byte(orig)
	k := []byte(key)
	// 分组秘钥
	// NewCipher该函数限制了输入k的长度必须为16, 24或者32
	block, err := aes.NewCipher(k[:16])
	if err != nil {
		return "", err
	}

	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 补全码
	origData = PKCS7Padding(origData, blockSize)
	// 加密模式
	blockMode := cipher.NewCBCEncrypter(block, k[:blockSize])
	// 创建数组
	cryted := make([]byte, len(origData))
	// 加密
	blockMode.CryptBlocks(cryted, origData)
	return base64.StdEncoding.EncodeToString(cryted), nil
}
func AesDecrypt(cryted string, key string) string {
	// 转成字节数组
	crytedByte, err := base64.StdEncoding.DecodeString(cryted)
	if err != nil {
		return "failed"
	}
	k := []byte(key)
	// 分组秘钥
	block, err := aes.NewCipher(k[:16])
	if err != nil {
		return "failed"
	}
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 加密模式
	blockMode := cipher.NewCBCDecrypter(block, k[:blockSize])
	// 创建数组
	orig := make([]byte, len(crytedByte))
	// 解密
	blockMode.CryptBlocks(orig, crytedByte)
	// 去补全码
	orig = PKCS7UnPadding(orig)
	return string(orig)
}

//补码
//AES加密数据块分组长度必须为128bit(byte[16])，密钥长度可以是128bit(byte[16])、192bit(byte[24])、256bit(byte[32])中的任意一个。
func PKCS7Padding(ciphertext []byte, blocksize int) []byte {
	padding := blocksize - len(ciphertext)%blocksize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

//去码
func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	if unpadding > length {
		return []byte("failed")
	}
	return origData[:(length - unpadding)]
}
