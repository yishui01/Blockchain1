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

//func NewTransaction()  {
//
//}

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
	if data == ""{
		data  = fmt.Sprintf("reward to %s %d btc", address, reward)
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

//生成交易ID
func (tx *Transaction) SetTXID() {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(tx)
	CheckErr(err, "编码transaction失败")
	hash := sha256.Sum256(buffer.Bytes())
	tx.TXID = hash[:]
}
