package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
)

const reward = 12.5 //奖励

type Transaction struct {
	TXID      []byte;    //交易ID
	TXInputs  []TXInput  //输入
	TXOutputs []TXOutput //输出
}

//input 判断传进来的string能不能解锁这个UTXO
func (input *TXInput) CanUnlockUTXOWith(unlockData string) bool {
	return input.ScriptSig == unlockData
}

//检查当前用户是否是UTXO的所有者
func (output *TXOutput) CanBeUnlockWith(unlockData string) bool {
	return output.ScriptPubKey == unlockData;
}

func NewTransaction(from, to string, amount float64, bc *BlockChain) {

}

type TXInput struct {
	TXID      []byte //所引用输出的交易ID
	Vout      int64  //所引用的output的索引值
	ScriptSig string //解锁脚本，指明可以使用某个output的条件
}

type TXOutput struct {
	Value        float64 //支付给收款方的金额
	ScriptPubKey string  //锁定脚本，指定收款方的地址
}

//创建coinbase交易，只有收款人，没有付款人，是矿工的奖励交易
func NewCoinBaseTx(address string, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("reward to %s %d btc", address, reward)
	}
	inputs := TXInput{
		TXID:      []byte{},
		Vout:      -1,
		ScriptSig: data,
	}
	outputs := TXOutput{
		Value:        reward,
		ScriptPubKey: address,
	}
	tx := Transaction{TXID: []byte{}, TXInputs: []TXInput{inputs}, TXOutputs: []TXOutput{outputs}}
	tx.SetTXID()
	return &tx;
}

func (tx *Transaction) IsCoinbase() bool {
	if len(tx.TXInputs) == 1 {
		if len(tx.TXInputs[0].TXID) == 0 && tx.TXInputs[0].Vout == -1 {
			return true;
		}
	}
	return false;
}

//生成交易ID
func (tx *Transaction) SetTXID() {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(tx)
	CheckErr(err, "编码transaction失败")
	hash := sha256.Sum256(buffer.Bytes())
	tx.TXID = hash[:]
}
