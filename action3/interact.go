package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"test/counter" // 替换为实际的包路径

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// 查询action2合约信息
const (
	SepoliaRPC           = "https://sepolia.infura.io/v3/pk"
	PrivateKey           = "私钥"
	DeployedContractAddr = "0x416330B109993786f2c93Cf176d575C7d73179e8" // 部署后获取的合约地址
)

func main() {
	// 1. 连接 Sepolia 测试网
	client, err := ethclient.Dial(SepoliaRPC)
	if err != nil {
		log.Fatalf("连接网络失败：%v", err)
	}
	defer client.Close()

	// 2. 初始化合约实例
	contractAddr := common.HexToAddress(DeployedContractAddr)
	counterInstance, err := counter.NewCounter(contractAddr, client)
	if err != nil {
		log.Fatalf("初始化合约实例失败：%v", err)
	}

	// 3. 调用合约只读方法（GetCount）- 无需交易费
	fmt.Println("=== 调用只读方法 GetCount ===")
	count, err := counterInstance.GetCount(&bind.CallOpts{})
	if err != nil {
		log.Fatalf("查询计数失败：%v", err)
	}
	fmt.Printf("当前计数：%d\n", count)

	// 4. 调用合约写方法（Increment）- 需要发送交易
	fmt.Println("\n=== 调用写方法 Increment ===")
	// 配置交易签名器
	privateKey, err := crypto.HexToECDSA(PrivateKey)
	if err != nil {
		log.Fatalf("解析私钥失败：%v", err) // 常见错误：私钥含 0x 前缀、长度不对（需 64 个十六进制字符）
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(11155111))
	if err != nil {
		log.Fatalf("创建签名器失败：%v", err)
	}
	auth.GasLimit = 300000
	auth.GasPrice, err = client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("获取 Gas 价格失败：%v", err)
	}

	// 发送 Increment 交易
	tx, err := counterInstance.Increment(auth)
	if err != nil {
		log.Fatalf("执行 Increment 失败：%v", err)
	}
	fmt.Printf("交易哈希：%s\n", tx.Hash().Hex())
	fmt.Println("等待区块确认...（约 10 秒）")

	// 5. 等待交易确认后，再次查询计数
	_, err = bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		log.Fatalf("交易确认失败：%v", err)
	}
	fmt.Println("交易已确认！")

	// 再次查询计数
	newCount, err := counterInstance.GetCount(&bind.CallOpts{})
	if err != nil {
		log.Fatalf("查询更新后计数失败：%v", err)
	}
	fmt.Printf("更新后计数：%d\n", newCount)
}
