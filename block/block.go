package block

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"jxblockchain/utils"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

const MINING_DIFFICULT = 2
const REWARD_ORIGIN = "BUFFER BLOCKCHAIN"
const MINING_REWARD = 500
const MINING_TIMER_SEC = 10
const (
	//以下参数可以添加到启动参数
	BLOCKCHAIN_PORT_RANGE_START      = 5000
	BLOCKCHAIN_PORT_RANGE_END        = 5003
	NEIGHBOR_IP_RANGE_START          = 0
	NEIGHBOR_IP_RANGE_END            = 0
	BLOCKCHIN_NEIGHBOR_SYNC_TIME_SEC = 10
)

type Block struct {
	nonce        int
	previousHash [32]byte
	timestamp    int64
	transactions []*Transaction
}

func NewBlock(nonce int, previousHash [32]byte, txs []*Transaction) *Block {
	b := new(Block)
	b.timestamp = time.Now().UnixNano()
	b.nonce = nonce
	b.previousHash = previousHash
	b.transactions = txs
	return b
}

func (b *Block) PreviousHash() [32]byte {
	return b.previousHash
}

func (b *Block) Nonce() int {
	return b.nonce
}

func (b *Block) Transactions() []*Transaction {
	return b.transactions
}

func (b *Block) Print() {
	log.Printf("%-15v:%30d\n", "timestamp", b.timestamp)
	//fmt.Printf("timestamp       %d\n", b.timestamp)
	log.Printf("%-15v:%30d\n", "nonce", b.nonce)
	log.Printf("%-15v:%30x\n", "previous_hash", b.previousHash)
	//log.Printf("%-15v:%30s\n", "transactions", b.transactions)
	for _, t := range b.transactions {
		t.Print()
	}
}

//mux：mux是一个sync.Mutex类型的字段，它是一个互斥锁。sync.Mutex是Go语言标准库中提供的一种锁机制，用于在并发编程中保护共享资源的访问。
//mux的目的是在多个并发操作访问共享资源时提供互斥访问，防止数据竞争和冲突。
//

// muxNeighbors：muxNeighbors也是一个sync.Mutex类型的字段，
// 它也是一个互斥锁。它的作用是保护neighbors字段的访问，因为neighbors是一个共享的字符串切片（[]string），
// 可能会被多个并发操作同时访问和修改。通过使用互斥锁，可以确保在修改neighbors字段时只有一个goroutine（Go语言中的并发执行单元）能够访问它，
// 从而避免数据不一致性和竞争条件。
type Blockchain struct {
	transactionPool   []*Transaction
	chain             []*Block
	blockchainAddress string
	port              uint16
	mux               sync.Mutex

	neighbors    []string
	muxNeighbors sync.Mutex
}

// 新建一条链的第一个区块
// NewBlockchain(blockchainAddress string) *Blockchain
// 函数定义了一个创建区块链的方法，它接收一个字符串类型的参数 blockchainAddress，
// 它返回一个区块链类型的指针。在函数内部，它创建一个区块链对象并为其设置地址，
// 然后创建一个创世块并将其添加到区块链中，最后返回区块链对象。
// 区块链创建和初始化
func NewBlockchain(blockchainAddress string, port uint16) *Blockchain {
	bc := new(Blockchain)
	b := &Block{}
	//给创世纪块的矿工奖励
	bc.AddTransaction(REWARD_ORIGIN, blockchainAddress, MINING_REWARD, nil, nil)
	bc.CreateBlock(0, b.Hash()) //创世纪块
	bc.blockchainAddress = blockchainAddress
	bc.port = port
	return bc
}

func (bc *Blockchain) Chain() []*Block {
	return bc.chain
}

func (bc *Blockchain) Run() {

	bc.StartSyncNeighbors()
	bc.ResolveConflicts()
	bc.StartMining()
}

//更新了区块链对象的邻居节点信息

func (bc *Blockchain) SetNeighbors() {
	bc.neighbors = utils.FindNeighbors(
		utils.GetHost(), bc.port,
		NEIGHBOR_IP_RANGE_START, NEIGHBOR_IP_RANGE_END,
		BLOCKCHAIN_PORT_RANGE_START, BLOCKCHAIN_PORT_RANGE_END)

	color.Blue("邻居节点：%v", bc.neighbors)
}

func (bc *Blockchain) SyncNeighbors() {
	bc.muxNeighbors.Lock()
	defer bc.muxNeighbors.Unlock()
	bc.SetNeighbors()
}

// 这段代码的目的是创建一个定时器，用于定期执行bc.SyncNeighbors()方法来同步更新区块链节点的邻居节点信息。
//通过调用bc.StartSyncNeighbors方法本身，可以实现定时器的循环调用，从而周期性地执行同步操作。

func (bc *Blockchain) StartSyncNeighbors() {
	bc.SyncNeighbors()
	_ = time.AfterFunc(time.Second*BLOCKCHIN_NEIGHBOR_SYNC_TIME_SEC, bc.StartSyncNeighbors)
}

// 将方法与这个Blockchain结构体的指针类型赋值给bc，让此方法与这个结构体相关联

func (bc *Blockchain) TransactionPool() []*Transaction {
	return bc.transactionPool
}

//这可以用于清除旧的交易数据，为区块链对象的交易池做准备，或者重置交易池以便进行新的交易收集和处理。

func (bc *Blockchain) ClearTransactionPool() {
	bc.transactionPool = bc.transactionPool[:0]
}

func (bc *Blockchain) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		//序列化后指定该字段的名称为chain
		Blocks []*Block `json:"chain"`
	}{
		Blocks: bc.chain,
	})
}

func (bc *Blockchain) UnmarshalJSON(data []byte) error {
	//匿名结构体

	v := &struct {
		Blocks *[]*Block `json:"chain"`
	}{
		//需要接收一个指向目标对象的指针，以便将反序列化后的值赋给目标对象
		Blocks: &bc.chain,
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	return nil
}

// (bc *Blockchain) CreateBlock(nonce int, previousHash [32]byte) *Block
//  函数是在区块链上创建新的区块，它接收两个参数：一个int类型的nonce和一个字节数组类型的 previousHash，
//  返回一个区块类型的指针。在函数内部，它使用传入的参数来创建一个新的区块，
//  然后将该区块添加到区块链的链上，并清空交易池。？？？交易池为什么要被清空？里面是否还有没有打包的交易

// 创造新块原方法

//区块添加

func (bc *Blockchain) CreateBlock(nonce int, previousHash [32]byte) *Block {
	b := NewBlock(nonce, previousHash, bc.transactionPool)

	bc.chain = append(bc.chain, b)
	bc.transactionPool = []*Transaction{}

	// 删除其他节点的交易
	for _, n := range bc.neighbors {
		endpoint := fmt.Sprintf("http://%s/transactions", n)
		client := &http.Client{}
		req, _ := http.NewRequest("DELETE", endpoint, nil)
		//发送 HTTP 请求并获取响应
		resp, _ := client.Do(req)
		log.Printf("%v", resp)
	}
	return b
}

func (bc *Blockchain) Print() {
	for i, block := range bc.chain {
		color.Green("%s BLOCK %d %s\n", strings.Repeat("=", 25), i, strings.Repeat("=", 25))
		block.Print()
	}
	color.Yellow("%s\n\n\n", strings.Repeat("*", 50))
}

func (b *Block) Hash() [32]byte {
	m, _ := json.Marshal(b)
	return sha256.Sum256([]byte(m))
}

func (trans *Transaction) TransactionHash() [32]byte {
	m, _ := json.Marshal(trans)
	return sha256.Sum256([]byte(m))
}

func (b *Block) MarshalJSON() ([]byte, error) {

	return json.Marshal(struct {
		Timestamp    int64          `json:"timestamp"`
		Nonce        int            `json:"nonce"`
		PreviousHash string         `json:"previous_hash"`
		Transactions []*Transaction `json:"transactions"`
	}{
		Timestamp:    b.timestamp,
		Nonce:        b.nonce,
		PreviousHash: fmt.Sprintf("%x", b.previousHash),
		Transactions: b.transactions,
	})
}

func (b *Block) UnmarshalJSON(data []byte) error {
	var previousHash string
	v := &struct {
		Timestamp    *int64          `json:"timestamp"`
		Nonce        *int            `json:"nonce"`
		PreviousHash *string         `json:"previous_hash"`
		Transactions *[]*Transaction `json:"transactions"`
	}{
		Timestamp:    &b.timestamp,
		Nonce:        &b.nonce,
		PreviousHash: &previousHash,
		Transactions: &b.transactions,
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	ph, _ := hex.DecodeString(*v.PreviousHash)
	copy(b.previousHash[:], ph[:32])
	return nil
}

func (bc *Blockchain) LastBlock() *Block {
	return bc.chain[len(bc.chain)-1]
}

//交易添加

func (bc *Blockchain) AddTransaction(
	sender string,
	recipient string,
	value int64,
	senderPublicKey *ecdsa.PublicKey,
	s *utils.Signature) bool {
	t := NewTransaction(sender, recipient, value)

	//如果是挖矿得到的奖励交易，不验证
	if sender == REWARD_ORIGIN {
		bc.transactionPool = append(bc.transactionPool, t)
		return true
	}

	//判断有没有足够的余额
	log.Printf("transaction.go sender:%s  account=%d", sender, bc.CalculateTotalAmount(sender))
	if bc.CalculateTotalAmount(sender) <= uint64(value) {
		log.Printf("ERROR: %s ，你的钱包里没有足够的钱", sender)
		return false
	}

	if bc.VerifyTransactionSignature(senderPublicKey, s, t) {
		log.Println("SUCCESS:矿工验证交易")
		bc.transactionPool = append(bc.transactionPool, t)
		for _, block := range bc.transactionPool {
			block.Print()
			log.Println("查看交易池里的交易")
		}
		log.Println(len(bc.transactionPool))
		return true
	} else {
		log.Println("ERROR: 验证交易")
	}
	return false

}

func (bc *Blockchain) CreateTransaction(sender string, recipient string, value uint64,
	senderPublicKey *ecdsa.PublicKey, s *utils.Signature) bool {
	isTransacted := bc.AddTransaction(sender, recipient, int64(value), senderPublicKey, s)

	if isTransacted {
		for _, n := range bc.neighbors {
			publicKeyStr := fmt.Sprintf("%064x%064x", senderPublicKey.X.Bytes(),
				senderPublicKey.Y.Bytes())
			signatureStr := s.String()
			bt := &TransactionRequest{
				&sender, &recipient, &publicKeyStr, &value, &signatureStr}
			m, _ := json.Marshal(bt)
			buf := bytes.NewBuffer(m)
			endpoint := fmt.Sprintf("http://%s/transactions", n)
			client := &http.Client{}
			req, _ := http.NewRequest("PUT", endpoint, buf)
			resp, _ := client.Do(req)
			log.Printf("   **  **  **  CreateTransaction : %v", resp)
		}
	}

	return isTransacted
}

func (bc *Blockchain) CopyTransactionPool() []*Transaction {
	transactions := make([]*Transaction, 0)
	for _, t := range bc.transactionPool {
		transactions = append(transactions,
			NewTransaction(t.senderAddress,
				t.receiveAddress,
				t.value,
			))
	}
	log.Println("查看复制交易池里是否复制成功", transactions)
	return transactions
}

func (bc *Blockchain) ValidProof(nonce int,
	previousHash [32]byte,
	transactions []*Transaction,
	difficulty int,
) bool {
	//zeros := strings.Repeat("0", difficulty)
	zeros := "12"
	// tmpBlock := Block{nonce: nonce, previousHash: previousHash, transactions: transactions, timestamp: time.Now().UnixNano()}
	tmpBlock := Block{nonce: nonce, previousHash: previousHash, transactions: transactions}

	//log.Printf("tmpBlock%+v", tmpBlock)
	tmpHashStr := fmt.Sprintf("%x", tmpBlock.Hash())
	//log.Println("guessHashStr", tmpHashStr)
	return tmpHashStr[:difficulty] == zeros
}

func (bc *Blockchain) ProofOfWork() int {
	transactions := bc.CopyTransactionPool() //选择交易？控制交易数量？
	previousHash := bc.LastBlock().Hash()
	log.Println("pow工作量证明:", transactions)
	log.Println("pow工作量证明:", previousHash)

	nonce := 0
	begin := time.Now()
	for !bc.ValidProof(nonce, previousHash, transactions, MINING_DIFFICULT) {
		nonce += 1
	}

	end := time.Now()
	//log.Printf("POW spend Time:%f", end.Sub(begin).Seconds())
	log.Printf("POW spend Time:%f Second", end.Sub(begin).Seconds())
	log.Printf("POW spend Time:%s", end.Sub(begin))

	return nonce
}

// 将交易池的交易打包
func (bc *Blockchain) Mining() bool {
	bc.mux.Lock()

	defer bc.mux.Unlock()

	// 此处判断交易池是否有交易，你可以不判断，打包无交易区块
	if len(bc.transactionPool) == 0 {
		log.Println("交易池里无交易")
		return false
	}

	bc.AddTransaction(REWARD_ORIGIN, bc.blockchainAddress, MINING_REWARD, nil, nil)
	nonce := bc.ProofOfWork()
	previousHash := bc.LastBlock().Hash()
	bc.CreateBlock(nonce, previousHash)
	log.Println("打包成功" + string(bc.port))
	log.Println("action=mining, status=success")

	for _, n := range bc.neighbors {
		endpoint := fmt.Sprintf("http://%s/consensus", n)
		client := &http.Client{}
		req, _ := http.NewRequest("PUT", endpoint, nil)
		resp, _ := client.Do(req)
		log.Printf("%v", resp)
	}

	return true
}

//账户余额查询

func (bc *Blockchain) CalculateTotalAmount(accountAddress string) uint64 {
	var totalAmount uint64 = 0
	for _, block := range bc.Chain() {
		for _, _tx := range block.Transactions() {
			if accountAddress == _tx.receiveAddress {
				totalAmount = totalAmount + uint64(_tx.value)
			}
			if accountAddress == _tx.senderAddress {
				totalAmount = totalAmount - uint64(_tx.value)
			}
		}
	}
	return totalAmount
}

func (bc *Blockchain) StartMining() {
	bc.Mining()
	// 使用time.AfterFunc函数创建了一个定时器，它在指定的时间间隔后执行bc.StartMining函数（自己调用自己）。
	_ = time.AfterFunc(time.Second*MINING_TIMER_SEC, bc.StartMining)
	color.Yellow("minetime: %v\n", time.Now())
}

type AmountResponse struct {
	Amount uint64 `json:"amount"`
}

type BlockResponse struct {
	Timestamp    int64          `json:"timestamp"`
	Nonce        int            `json:"nonce"`
	PreviousHash string         `json:"previous_hash"`
	Transactions []*Transaction `json:"transactions"`
}

func (ar *AmountResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Amount uint64 `json:"amount"`
	}{
		Amount: ar.Amount,
	})
}

type Transaction struct {
	senderAddress  string
	receiveAddress string
	value          int64
	hash           [32]byte
}

func NewTransaction(sender string, receive string, value int64) *Transaction {

	t := Transaction{
		senderAddress:  sender,
		receiveAddress: receive,
		value:          value,
	}
	t.hash = t.TransactionHash()
	return &t

}

func (bc *Blockchain) VerifyTransactionSignature(
	senderPublicKey *ecdsa.PublicKey, s *utils.Signature, t *Transaction) bool {

	//避开TxHash做验证签名，因为在签名时没有写入TxHash信息
	m, _ := json.Marshal(struct {
		Sender    string `json:"sender_blockchain_address"`
		Recipient string `json:"recipient_blockchain_address"`
		Value     int64  `json:"value"`
	}{
		Sender:    t.senderAddress,
		Recipient: t.receiveAddress,
		Value:     t.value,
	})

	h := sha256.Sum256([]byte(m))
	return ecdsa.Verify(senderPublicKey, h[:], s.R, s.S)
}

func (t *Transaction) Print() {
	color.Red("%s\n", strings.Repeat("~", 30))
	color.Cyan("发送地址             %s\n", t.senderAddress)
	color.Cyan("接受地址             %s\n", t.receiveAddress)
	color.Cyan("金额                 %d\n", t.value)

}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Sender    string `json:"sender_blockchain_address"`
		Recipient string `json:"recipient_blockchain_address"`
		Value     int64  `json:"value"`
		Hash      string `json:"transaction_hash"`
	}{
		Sender:    t.senderAddress,
		Recipient: t.receiveAddress,
		Value:     t.value,
		Hash:      fmt.Sprintf("%x", t.hash),
	})
}

func (t *Transaction) UnmarshalJSON(data []byte) error {
	var hash string
	v := &struct {
		Sender    *string `json:"sender_blockchain_address"`
		Recipient *string `json:"recipient_blockchain_address"`
		Value     *int64  `json:"value"`
		Hash      *string `json:"transaction_hash"`
	}{
		Sender:    &t.senderAddress,
		Recipient: &t.receiveAddress,
		Value:     &t.value,
		Hash:      &hash,
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	ph, _ := hex.DecodeString(*v.Hash)
	copy(t.hash[:], ph[:32])
	return nil
}

func (bc *Blockchain) ValidChain(chain []*Block) bool {
	preBlock := chain[0]
	currentIndex := 1
	for currentIndex < len(chain) {
		b := chain[currentIndex]
		if b.previousHash != preBlock.Hash() {
			return false
		}

		if !bc.ValidProof(b.Nonce(), b.PreviousHash(), b.Transactions(), MINING_DIFFICULT) {
			return false
		}

		preBlock = b
		currentIndex += 1
	}
	return true
}

func (bc *Blockchain) ResolveConflicts() bool {
	var longestChain []*Block = nil
	maxLength := len(bc.chain)

	for _, n := range bc.neighbors {
		endpoint := fmt.Sprintf("http://%s/chain", n)
		resp, err := http.Get(endpoint)
		if err != nil {
			color.Red("                 错误 ：ResolveConflicts GET请求")
			return false
		} else {
			color.Green("                正确 ：ResolveConflicts  GET请求")
		}
		if resp.StatusCode == 200 {
			var bcResp Blockchain
			decoder := json.NewDecoder(resp.Body)
			err1 := decoder.Decode(&bcResp)

			if err1 != nil {
				color.Red("                 错误 ：ResolveConflicts Decode")
				return false
			} else {
				color.Green("                正确 ：ResolveConflicts  Decode")
			}

			chain := bcResp.Chain()
			color.Cyan("   ResolveConflicts   chain len:%d ", len(chain))
			fmt.Println(len(chain))
			fmt.Println(maxLength)

			fmt.Println(bc.ValidChain(chain))
			if len(chain) > maxLength && bc.ValidChain(chain) {
				maxLength = len(chain)
				longestChain = chain
			}
		}
	}

	color.Cyan("   ResolveConflicts   longestChain len:%d ", len(longestChain))

	if longestChain != nil {
		bc.chain = longestChain
		log.Printf("Resovle confilicts replaced")
		return true
	}
	log.Printf("Resovle conflicts not replaced")
	return false
}

type TransactionRequest struct {
	SenderBlockchainAddress    *string `json:"sender_blockchain_address"`
	RecipientBlockchainAddress *string `json:"recipient_blockchain_address"`
	SenderPublicKey            *string `json:"sender_public_key"`
	Value                      *uint64 `json:"value"`
	Signature                  *string `json:"signature"`
}

func (tr *TransactionRequest) Validate() bool {
	if tr.SenderBlockchainAddress == nil ||
		tr.RecipientBlockchainAddress == nil ||
		tr.SenderPublicKey == nil ||
		tr.Value == nil ||
		tr.Signature == nil {
		return false
	}
	return true
}

// BlockChainCheck 区块链查询 区块号
// 如果前端输入了一个数字，此时就接收数字转化成blockchain.chain[]

func (bc *Blockchain) GetBlockByNum(num int) *Block {
	//如果用户输入了区块号，则单独打印该区块号内的所有内容
	color.Green("%s BLOCK %d %s\n", strings.Repeat("=", 25), num, strings.Repeat("=", 25))
	b := bc.chain[num]
	return b
}

// BlockChainHashCheck 用户通过hash查找区块
//区块链查询 区块哈希

func (bc *Blockchain) GetBlockByHash(hash [32]byte) *Block {
	chain := bc.Chain()
	log.Println("hash in getBlockByHash:", hash)

	for _, block := range chain {

		log.Printf("blockHash: %v", block.Hash())
		if block.Hash() == hash {
			return block
		}
	}
	return nil
}

// GetTransactionByHash 用户通过交易hash查找交易
//交易查询

func (bc *Blockchain) GetTransactionByHash(transactionHash [32]byte) *Transaction {
	log.Println("hash in getBlockByHash:", transactionHash)
	for _, block := range bc.chain {
		for _, transactions := range block.transactions {
			if transactions.hash == transactionHash {
				log.Println("每个交易交易里的hash：", transactions.hash)
				return transactions
			}
		}
	}
	return nil
}
