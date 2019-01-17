package main

func main() {
	bc := NewBlockChain()
	cli := Cli{bc}
	cli.Run()
}
