package main

import (
	"./bolt"
	"fmt"
	"os"
)

const dbFile = "blockChain.db"
const blockBucket = "bucket"
const lastHashKey = "lastkey"

const genesisInfo  = "你们的皇帝回来了"

type BlockChain struct {
	//blocks []*Block 已废弃

	//数据库的操作句柄
	db *bolt.DB
	//最后一个区块的hash值
	tail []byte
}

func isDBExist() bool {
	_, err := os.Stat(dbFile)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

//创建一个区块链
func InitBlockChain(address string) *BlockChain {
	//block := NewGenesisBlock()
	//return &BlockChain{blocks: []*Block{block}}
	if isDBExist() {
		fmt.Println("blockchain exist already, not to created!")
		os.Exit(1)
	}
	//func Open(path string, mode os.FileMode, options *Options) (*DB, error) {
	db, err := bolt.Open(dbFile, 0600, nil)
	CheckErr(err, "初始化BlockChain失败");
	var lastHash []byte

	//func (db *DB) Update(fn func(*Tx) error) error {
	err1 := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		//没有bucket，要去创建bucket，将数据写到数据库的bucket中
		coinbase := NewCoinBaseTx(address, genesisInfo)
		genesis := NewGenesisBlock(coinbase) //返回一个创建好的 // *block
		bucket, err2 := tx.CreateBucket([]byte(blockBucket))
		CheckErr(err2, "创建Bucket失败")
		err3 := bucket.Put(genesis.Hash, genesis.Serialize()) //往里写数据
		CheckErr(err3, "往bucket中写数据失败")
		err4 := bucket.Put([]byte(lastHashKey), genesis.Hash) //设置最后一个hash值
		CheckErr(err4, "往bucket中写lasthash失败")
		lastHash = genesis.Hash
		return nil;
	});
	CheckErr(err1, "找Bucket失败");
	return &BlockChain{db: db, tail: lastHash}
}

func GetBlockChainHandler() *BlockChain {
	if !isDBExist() {
		fmt.Println("Please create blockchain first!")
		os.Exit(1)
	}
	//block := NewGenesisBlock()
	//return &BlockChain{blocks: []*Block{block}}

	//func Open(path string, mode os.FileMode, options *Options) (*DB, error) {
	db, err := bolt.Open(dbFile, 0600, nil)
	CheckErr(err, "打开文件失败");
	var lastHash []byte

	//func (db *DB) Update(fn func(*Tx) error) error {
	err1 := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket != nil {
			//取出最后区块的hash值返回
			lastHash = bucket.Get([]byte(lastHashKey))
		}
		return nil;
	});
	CheckErr(err1, "找Bucket失败");
	return &BlockChain{db: db, tail: lastHash}
}

//往区块链里面加区块
func (bc *BlockChain) AddBlock(txs []*Transaction) {
	var prevHash []byte
	err1 := bc.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil {
			os.Exit(1)
		}
		prevHash = bucket.Get([]byte(lastHashKey))
		return nil
	})
	CheckErr(err1, "addblock的err1失败")

	block := NewBlock(txs, prevHash)
	err2 := bc.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil {
			os.Exit(1)
		}
		err3 := bucket.Put(block.Hash, block.Serialize()) //往里写数据
		CheckErr(err3, "addblock3")
		err4 := bucket.Put([]byte(lastHashKey), block.Hash)
		CheckErr(err4, "addblock4")
		bc.tail = block.Hash
		return nil
	})
	CheckErr(err2, "addblock的err2失败")
}

//迭代器，就是一个对象，它里面包含了一个游标，一直向前（后）移动，完成整个容器的遍历
type BlockChainIterator struct {
	currHash []byte
	db       *bolt.DB
}

//创建迭代器，同时初始化指向最后一个区块
func (bc *BlockChain) NewIterator() *BlockChainIterator {
	return &BlockChainIterator{currHash: bc.tail, db: bc.db}
}

//解序列化当前currentHash指向的block并返回指针，然后把currHash指向前一个hash
func (it *BlockChainIterator) Next() (block *Block) {
	err := it.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil {
			return nil
		}
		data := bucket.Get(it.currHash)
		block = Unserialize(data)
		it.currHash = block.PrevBlockHash
		return nil
	})

	CheckErr(err, "next出错")
	return
}
