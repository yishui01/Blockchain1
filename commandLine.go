package main

import (
	"flag"
	"fmt"
	"os"
)

const usage = `
	createChain --address ADDRESS "create a blockchain"
	addBlock --data  DATA      "add a block to blockchain"
	send --from FORM --to TO --amount AMOUNT  "send coin from FROM to TO"
	getBalance --address ADDRESS  "get balance of the address"
	printChain			"print all blocks"
`
const AddBlockCmdString = "addBlock"
const PrintChainCmdString = "printChain"
const CreateChainCmdString = "createChain"
const GetBalanceCmdString = "getBalance"

type Cli struct {
	//bc *BlockChain
}

func (cli *Cli) AddBlock(data string) {
	//bc := GetBlockChainHandler()  //TODO
	//bc.AddBlock(data)
}

func (cli *Cli) PrintChain() {
	//打印数据
	bc := GetBlockChainHandler()
	defer bc.db.Close()
	it := bc.NewIterator()
	for {
		block := it.Next()
		fmt.Printf("Version: %d\n", block.Version)
		fmt.Printf("PrevBlockHash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("MerKelRoot: %x\n", block.MerKelRoot)
		fmt.Printf("TimeStamp: %d\n", block.TimeStamp)
		fmt.Printf("Bits: %d\n", block.Bits)
		fmt.Printf("Nonce: %d\n", block.Nonce)
		//fmt.Printf("Data: %x\n", block.Data) TODO
		fmt.Printf("IsValid: %v\n", NewProofOfWork(block).IsValid())
		fmt.Printf("\n\n")
		if len(block.PrevBlockHash) == 0 {
			fmt.Println("打印完毕")
			break;
		}
	}
}

func (cli *Cli) printUsage() {
	fmt.Println("无效的输入")
	fmt.Println(usage)
	os.Exit(1);
}

func (cli *Cli) ParaCheck() {
	if len(os.Args) < 2 {
		fmt.Println("参数无效！")
		cli.printUsage()
	}
}

func (cli *Cli) Run() {
	cli.ParaCheck()

	createChainCmd := flag.NewFlagSet(CreateChainCmdString, flag.ExitOnError)
	addBlockCmd := flag.NewFlagSet(AddBlockCmdString, flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet(GetBalanceCmdString, flag.ExitOnError)
	printChainCmd := flag.NewFlagSet(PrintChainCmdString, flag.ExitOnError)

	addPara := addBlockCmd.String("data", "", "block transaction info!")
	createChainPara := createChainCmd.String("address", "", "address info!")
	getBalanceCmdPara := getBalanceCmd.String("address", "", "address info!")

	switch os.Args[1] {
	case CreateChainCmdString:
		//创建区块链
		err := createChainCmd.Parse(os.Args[2:])
		CheckErr(err, "Run0000出错")
		if createChainCmd.Parsed() {
			if *createChainPara == "" {
				fmt.Println("address参数不能为空")
				cli.printUsage()
			}
			cli.CreateChain(*createChainPara)
		}
	case AddBlockCmdString:
		//添加区块
		err := addBlockCmd.Parse(os.Args[2:])
		CheckErr(err, "Run1111出错")
		if addBlockCmd.Parsed() {
			if *addPara == "" {
				fmt.Println("data参数不能为空")
				cli.printUsage()
			}
			cli.AddBlock(*addPara)
		}
	case GetBalanceCmdString:
		//获取余额
		err := getBalanceCmd.Parse(os.Args[2:])
		CheckErr(err, "getBalance出错")
		if getBalanceCmd.Parsed() {
			if *getBalanceCmdPara == "" {
				fmt.Println("address参数不能为空")
				cli.printUsage()
			}
			cli.GetBalance(*getBalanceCmdPara)
		}
	case PrintChainCmdString:
		//打印输出
		err := printChainCmd.Parse(os.Args[2:])
		CheckErr(err, "打印出错")
		if printChainCmd.Parsed() {
			cli.PrintChain()
		}

	default:
		cli.printUsage()
	}
}

func (cli *Cli) CreateChain(address string) {
	bc := InitBlockChain(address)
	defer bc.db.Close()
	fmt.Println("Create blockchain successfully!")
}

//获取当前用户的余额
func (cli *Cli) GetBalance(address string) float64 {
	bc := GetBlockChainHandler() //TODO
	defer bc.db.Close()
	utxos := bc.FindUTXO(address)
	var total float64 = 0.0
	//遍历所有的utxo，获取金总数
	for _, utxo := range utxos {
		total += utxo.Value
	}
	fmt.Println("当前",address,"的余额为",total)
	return total
}
