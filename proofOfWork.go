package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

type ProofOfWork struct {
	block  *Block   //区块
	target *big.Int //目标值,只要找到一个比这个数小的值，就算完成目标
}

const targetBits = 24

//创建一个工作量证明
func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1);
	target.Lsh(target, uint(256-targetBits))
	pow := ProofOfWork{block: block, target: target}
	return &pow
}

func (pow *ProofOfWork) PrepareData(nonce int64) []byte {
	block := pow.block
	tmp := [][]byte{
		IntToByte(block.Version),
		block.PrevBlockHash,
		block.MerKelRoot,
		IntToByte(block.TimeStamp),
		IntToByte(targetBits), //难度值
		IntToByte(nonce),      //随机值
		block.Data}
	data := bytes.Join(tmp, []byte{})

	return data
}

//挖矿
func (pow *ProofOfWork) Run() (int64, []byte) {
	//1、拼装数据
	//2、哈希值转成big.Int
	var hash [32]byte;
	var nonce int64 = 0
	var hashInt big.Int;
	fmt.Println("开始挖矿")
	fmt.Printf("target hash:%x\n",pow.target.Bytes())

	for nonce < math.MaxInt64 {
		data := pow.PrepareData(nonce)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])
		if hashInt.Cmp(pow.target) == -1 { //hashInt比target小，目标达成
			fmt.Printf("found hash, nonce :%x, hash: %d\n", hash,nonce)
			break;
		} else {
			nonce++
		}
	}
	return nonce, hash[:]
}
