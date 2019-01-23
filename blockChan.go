package main

import (
	"./bolt"
	"fmt"
	"os"
)

const dbFile = "blockChain.db"
const blockBucket = "bucket"
const lastHashKey = "lastkey"

const genesisInfo = "你们的皇帝回来了"

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

//返回指定地址能够支配的utxo的交易集合
func (bc *BlockChain) FindUTXOTransactions(address string) []Transaction {
	//包含目标utxo的交易集合
	var UTXOTransactions []Transaction
	//存储使用过的utxo的集合 map[交易ID]索引数组
	//这里要考虑多个索引的问题，所以要用数组来存索引ID
	spentUTXO := make(map[string][]int64)
	it := bc.NewIterator()

	for {
		//遍历区块
		block := it.Next()

		//遍历区块类的交易
		for _, tx := range block.Transactions {
			//遍历input
			//目的：找到已经消耗过的utxo，把他们放到一个集合里
			//需要两个字段来标识使用过的utxo   交易ID和output的索引

			if !(tx.IsCoinbase()) {
				for _, input := range tx.TXInputs {
					if input.CanUnlockUTXOWith(address) {
						spentUTXO[string(tx.TXID)] = append(spentUTXO[string(tx.TXID)], input.Vout)
					}
				}
			}

			//遍历交易里的output
			//目的：找到所有能支配的utxo
			for currentIndex, output := range tx.TXOutputs {
				//检查当前的output是否已经被消耗（是否在上面input的那个数组中），如果已被消耗，continue
				if spentUTXO[string(tx.TXID)] != nil {
					//如果找到了，代表当前交易里面有已被消耗的utxo
					indexes := spentUTXO[string(tx.TXID)]
					for _, index := range indexes {
						if int64(currentIndex) == int64(index) {
							//当前output的索引和那个记录的已被消耗的output的索引相等，那么这个output是消耗了的
							break
						}
					}
				}
				//遍历output
				//比对传入的地址和这个utxo的地址所有者，是不是同一个，是的
				if output.CanBeUnlockWith(address) {
					UTXOTransactions = append(UTXOTransactions, *tx)
				}
			}

		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return UTXOTransactions
}
