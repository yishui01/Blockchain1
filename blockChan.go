package main

type BlockChain struct {
	blocks []*Block
}

//创建一个区块链
func NewBlockChain() *BlockChain {
	block := NewGenesisBlock()
	return &BlockChain{blocks: []*Block{block}}
}

//往区块链里面加区块
func (bc *BlockChain) AddBlock(data string) {
	prevBlockHash := bc.blocks[len(bc.blocks)-1].Hash
	block := NewBlock(data, prevBlockHash)
	bc.blocks = append(bc.blocks, block)
}
