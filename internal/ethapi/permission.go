package ethapi

import (
	"github.com/simplechain-org/simplechain/common"
	"github.com/simplechain-org/simplechain/common/hexutil"
	"golang.org/x/crypto/sha3"
	"math/big"
)

var managerContractAddress = common.HexToAddress("0xC0018822caC60FE4f223b758C83636546cF0D535")

func isManagerData(address common.Address) (hexutil.Bytes, error) {
	fnId, err := getFnId("isManager(address)")
	if err != nil {
		return nil, err
	}

	var data []byte
	data = append(data, fnId...)
	data = append(data, common.LeftPadBytes(address.Bytes(), 32)...)

	return data, nil
}

func addManagerData(address common.Address) (hexutil.Bytes, error) {
	fnId, err := getFnId("addManager(address)")
	if err != nil {
		return nil, err
	}

	var data []byte
	data = append(data, fnId...)
	data = append(data, common.LeftPadBytes(address.Bytes(), 32)...)

	return data, nil
}

func deleteManagerData(address common.Address) (hexutil.Bytes, error) {
	fnId, err := getFnId("deleteManager(address)")
	if err != nil {
		return nil, err
	}

	var data []byte
	data = append(data, fnId...)
	data = append(data, common.LeftPadBytes(address.Bytes(), 32)...)

	return data, nil
}

func setPermissionData(address common.Address, level uint8) (hexutil.Bytes, error) {
	fnId, err := getFnId("setPermission(address,uint8)")
	if err != nil {
		return nil, err
	}

	var data []byte
	data = append(data, fnId...)
	data = append(data, common.LeftPadBytes(address.Bytes(), 32)...)
	data = append(data, common.LeftPadBytes(new(big.Int).SetUint64(uint64(level)).Bytes(), 32)...)

	return data, nil
}

func getPermissionData(address common.Address) (hexutil.Bytes, error) {
	fnId, err := getFnId("getPermission(address)")
	if err != nil {
		return nil, err
	}

	var data []byte
	data = append(data, fnId...)
	data = append(data, common.LeftPadBytes(address.Bytes(), 32)...)

	return data, nil
}

func managePermissionTestData() (hexutil.Bytes, error) {
	fnId, err := getFnId("managePermissionTest()")
	if err != nil {
		return nil, err
	}

	var data []byte
	data = append(data, fnId...)
	return data, nil
}

func saveDataData(msg string) (hexutil.Bytes, error) {
	fnId, err := getFnId("saveCryptoData(string)")
	if err != nil {
		return nil, err
	}

	var data []byte
	data = append(data, fnId...)
	data = append(data, common.LeftPadBytes(new(big.Int).SetUint64(32).Bytes(), 32)...)
	data = append(data, common.LeftPadBytes(new(big.Int).SetUint64(uint64(len(msg))).Bytes(), 32)...)

	l := len(msg)
	if l%32 > 0 {
		l = (l/32 + 1) * 32
	}
	data = append(data, common.RightPadBytes([]byte(msg), l)...)
	return data, nil
}

func getCryptoDataData() (hexutil.Bytes, error) {
	fnId, err := getFnId("getCryptoData()")
	if err != nil {
		return nil, err
	}

	var data []byte
	data = append(data, fnId...)
	//data = append(data, common.LeftPadBytes(address.Bytes(), 32)...)
	return data, nil
}

func buildTxArgs(from common.Address, input hexutil.Bytes) *SendTxArgs {
	gas := hexutil.Uint64(200000)

	return &SendTxArgs{
		From:     from,
		To:       &managerContractAddress,
		Gas:      &gas,
		GasPrice: &hexutil.Big{},
		Value:    &hexutil.Big{},
		Input:    &input,
	}
}

func buildCallArgs(data hexutil.Bytes) *CallArgs {
	return &CallArgs{
		From:     common.Address{},
		To:       &managerContractAddress,
		Gas:      100000,
		GasPrice: hexutil.Big{},
		Value:    hexutil.Big{},
		Data:     data,
	}
}

func getFnId(fnSignature string) ([]byte, error) {
	hash := sha3.NewLegacyKeccak256()
	_, err := hash.Write([]byte(fnSignature))
	if err != nil {
		return nil, err
	}
	return hash.Sum(nil)[:4], nil
}
