package main

import (
	"fmt"
	"github.com/fatih/color"
	"jxblockchain/block"
	"jxblockchain/wallet"
)

func main() {

	//wallet_swk := wallet.NewWallet()     //孙悟空
	//wallet_zbj := wallet.NewWallet()     //猪八戒
	wallet_johnhai := wallet.NewWallet() //矿工

	//fmt.Printf("孙悟空的account:%s\n", wallet_swk.BlockchainAddress())
	//fmt.Printf("猪八戒的account:%s\n", wallet_zbj.BlockchainAddress())
	//fmt.Printf("矿工  的account:%s\n", wallet_johnhai.BlockchainAddress())

	blockchain := block.NewBlockchain(wallet_johnhai.BlockchainAddress(), 5000)
	blockchain.Print()
	color.Magenta("===矿工帐号信息====\n")
	color.Magenta("矿工private_key\n %v\n", wallet_johnhai.PrivateKeyStr())
	color.Magenta("矿工publick_key\n %v\n", wallet_johnhai.PublicKeyStr())
	color.Magenta("矿工blockchain_address\n %s\n", wallet_johnhai.BlockchainAddress())
	color.Magenta("===============\n")
	//钱包 提交一笔交易
	//t := wallet.NewTransaction(
	//	wallet_johnhai.PrivateKey(),
	//	wallet_johnhai.PublicKey(),
	//	wallet_johnhai.BlockchainAddress(),
	//	wallet_zbj.BlockchainAddress(),
	//	28)
	//
	////区块链 打包交易
	//isAdded := blockchain.AddTransaction(
	//	wallet_johnhai.BlockchainAddress(),
	//	wallet_zbj.BlockchainAddress(),
	//	28,
	//	wallet_johnhai.PublicKey(),
	//	t.GenerateSignature())
	//
	//fmt.Println("这笔交易验证通过吗? ", isAdded)
	//
	//t2 := wallet.NewTransaction(
	//	wallet_swk.PrivateKey(),
	//	wallet_swk.PublicKey(),
	//	wallet_swk.BlockchainAddress(),
	//	wallet_zbj.BlockchainAddress(),
	//	84)
	//fmt.Println("这笔交易验证通过吗? ", isAdded)
	//

	////区块链 打包交易
	//isAdded = blockchain.AddTransaction(
	//	wallet_swk.BlockchainAddress(),
	//	wallet_zbj.BlockchainAddress(),
	//	84,
	//	wallet_swk.PublicKey(),
	//	t2.GenerateSignature())
	//
	//fmt.Println("这笔交易验证通过吗? ", isAdded)
	//
	//blockchain.Mining()
	//blockchain.Print()

	//fmt.Printf("孙悟空 %d\n", blockchain.CalculateTotalAmount(wallet_swk.BlockchainAddress()))
	//fmt.Printf("猪八戒 %d\n", blockchain.CalculateTotalAmount(wallet_zbj.BlockchainAddress()))
	//fmt.Printf("矿工   %d\n", blockchain.CalculateTotalAmount(wallet_johnhai.BlockchainAddress()))

	//fmt.Println("*******************************************")
	fmt.Println("hello")
	//
	//blockchain.GetTransactionByHash(blockchain.Chain()[1].Transactions()[0].TransactionHash())

}
