package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"os"
)

const reward = 12.5 //奖励

type Transaction struct {
	TXID      []byte;    //交易ID
	TXInputs  []TXInput  //输入
	TXOutputs []TXOutput //输出
}

//指明交易发起人可支付资金的来源
type TXInput struct {
	TXID      []byte //所引用utxo所在的交易ID
	Vout      int64  //所引用的utxo在那个交易中的output数组中的索引值
	ScriptSig string //解锁脚本，指明可以使用某个output的条件
}
//指明交易发起人要把钱打到哪里去
type TXOutput struct {
	Value        float64 //支付给收款方的金额
	ScriptPubKey string  //锁定脚本，指定收款方的地址
}

//input 判断传进来的string能不能解锁这个UTXO
func (input *TXInput) CanUnlockUTXOWith(unlockData string) bool {
	return input.ScriptSig == unlockData
}

//检查当前用户是否是UTXO的所有者
func (output *TXOutput) CanBeUnlockWith(unlockData string) bool {
	return output.ScriptPubKey == unlockData;
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

//创建普通交易， send的辅助函数
func NewTransaction(from, to string, amount float64, bc *BlockChain) *Transaction {

	//map[string][]int64  key : 交易ID   value:引用output的索引数组
	validUTXOs := make(map[string][]int64)
	var total float64
	/*所需要的，合理的utxo的集合*/   /*返回utxo的金额总和*/
	validUTXOs , total= bc.FindSuitableUTXOs(from, amount)

	//validUTXOs[0x1111111111] = []int64{1}
	//validUTXOs[0x2222222222] = []int64{0}
	//...
	//validUTXOs[0x3333333333] = []int64{0, 4, 8}
	if total < amount {
		fmt.Println("余额不足")
		os.Exit(1)
	}
	var inputs []TXInput
	var outputs []TXOutput
	//fmt.Println("返回的长度是",len(validUTXOs))
	//fmt.Println("返回的total是",total)

	//1、创建inputs
	//余额足够，进行output到input的转换
	//遍历有效的utxo合集
	for txId, outputIndexes := range validUTXOs {
		//遍历所有引用的utxo的索引，每一个索引需要创建一个input
		for _, index := range outputIndexes {
			input := TXInput{TXID: []byte(txId), Vout: int64(index), ScriptSig: from}
			inputs = append(inputs, input)
		}
	}
	//2、创建output
	output := TXOutput{amount, to}
	outputs = append(outputs, output)

	if total > amount { //总数大于所需，那么找零
		output := TXOutput{total - amount, from}
		outputs = append(outputs, output)
	}

	tx := Transaction{TXID: []byte{}, TXInputs: inputs, TXOutputs: outputs}
	tx.SetTXID()
	return &tx;
}

//validUTXOs  /*所需要的，合理的utxo的集合*/  //返回值为 集合、以及集合金额总和
func (bc *BlockChain) FindSuitableUTXOs(address string, amount float64) (map[string][]int64, float64) {
	txs := bc.FindUTXOTransactions(address) ////返回指定地址能够支配的utxo的交易集合
	valid_utxos := make(map[string][]int64)
	total := 0.0

	FINALLY:
	for _, tx := range txs {
		outputs := tx.TXOutputs
		for index, output := range outputs {
			if output.CanBeUnlockWith(address) {
				if total < amount {
					valid_utxos[string(tx.TXID)] = append(valid_utxos[string(tx.TXID)], int64(index))
					total += output.Value
				} else {
					break FINALLY
				}
			}
		}
	}

	return valid_utxos,total

}
