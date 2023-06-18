package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"jxblockchain/block"
	"jxblockchain/utils"
	"jxblockchain/wallet"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

const tempDir = "walletServer/htmltemplate"

type WalletServer struct {
	port    uint16
	gateway string //区块链的节点地址
}

func NewWalletServer(port uint16, gateway string) *WalletServer {
	return &WalletServer{port, gateway}
}

func (ws *WalletServer) Port() uint16 {
	return ws.port
}

func (ws *WalletServer) Gateway() string {
	return ws.gateway
}

func (ws *WalletServer) Index(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		t, _ := template.ParseFiles(path.Join(tempDir, "index.html"))
		t.Execute(w, "")
	default:
		log.Printf("ERROR: 非法的HTTP请求方式")
	}
}

func (ws *WalletServer) Wallet(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	//设置允许的方法
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

	switch req.Method {
	case http.MethodPost:
		w.Header().Add("Content-Type", "application/json")
		log.Println("生成钱包post")
		myWallet := wallet.NewWallet()
		m, _ := myWallet.MarshalJSON()
		io.WriteString(w, string(m[:]))
	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Println("ERROR: 非法的HTTP请求方式")
	}
}

func (ws *WalletServer) walletByPrivatekey(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//设置允许的方法
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	switch req.Method {
	case http.MethodPost:

		w.Header().Add("Content-Type", "application/json")
		privatekey := req.FormValue("privatekey")
		log.Println("walletByPrivatekey:", privatekey)
		myWallet := wallet.LoadWallet(privatekey)
		m, _ := myWallet.MarshalJSON()
		io.WriteString(w, string(m[:]))
	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Println("ERROR: Invalid HTTP Method")
	}
}

type TransactionRequest struct {
	SenderPrivateKey           *string `json:"sender_private_key"`
	SenderBlockchainAddress    *string `json:"sender_blockchain_address"`
	RecipientBlockchainAddress *string `json:"recipient_blockchain_address"`
	SenderPublicKey            *string `json:"sender_public_key"`
	Value                      *string `json:"value"`
}

func (tr *TransactionRequest) Validate() bool {
	if tr.SenderPrivateKey == nil ||
		tr.SenderBlockchainAddress == nil ||
		tr.RecipientBlockchainAddress == nil || strings.TrimSpace(*tr.RecipientBlockchainAddress) == "" ||
		tr.SenderPublicKey == nil ||
		tr.Value == nil || len(*tr.Value) == 0 {
		return false
	}
	return true
}

func (ws *WalletServer) CreateTransaction(
	w http.ResponseWriter,
	req *http.Request) {
	log.Printf("Call CreateTransaction  METHOD:%s\n", req.Method)
	if req.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}
	defer req.Body.Close()
	switch req.Method {
	case http.MethodPost:
		// 解析请求体中的数据为 TransactionRequest 结构
		var t TransactionRequest
		log.Println("req.Body==", req.Body)
		decoder := json.NewDecoder(req.Body)
		decoder.Decode(&t)
		log.Printf("\n\n\n")
		log.Println("发送人公钥SenderPublicKey ==", *t.SenderPublicKey)
		log.Println("发送人私钥SenderPrivateKey ==", *t.SenderPrivateKey)
		log.Println("发送人地址SenderBlockchainAddress ==", *t.SenderBlockchainAddress)
		log.Println("接收人地址RecipientBlockchainAddress ==", *t.RecipientBlockchainAddress)
		log.Println("金额Value ==", *t.Value)
		log.Printf("\n\n\n")
		// 将发送人的公钥字符串转换为公钥对象
		publicKey := utils.PublicKeyFromString(*t.SenderPublicKey)
		// 使用发送人的私钥字符串和公钥对象生成私钥对象
		privateKey := utils.PrivateKeyFromString(*t.SenderPrivateKey, publicKey)
		// 将金额字符串转换为无符号整数
		value, err := strconv.ParseUint(*t.Value, 10, 64)
		if err != nil {
			log.Println("ERROR: parse error")
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}
		// 验证字段是否完整
		if !t.Validate() {
			log.Println("ERROR: missing field(s)")
			io.WriteString(w, string(utils.JsonStatus("Validate fail")))
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Add("Content-Type", "application/json")

		// 交易签名
		// 创建交易对象并进行签名
		transaction := wallet.NewTransaction(privateKey, publicKey,
			*t.SenderBlockchainAddress, *t.RecipientBlockchainAddress, value)
		signature := transaction.GenerateSignature()
		signatureStr := signature.String()
		color.Red("signature:%s", signature)
		// 创建发送给 BlockServer 的交易请求对象
		bt := &block.TransactionRequest{
			SenderBlockchainAddress:    t.SenderBlockchainAddress,
			RecipientBlockchainAddress: t.RecipientBlockchainAddress,
			SenderPublicKey:            t.SenderPublicKey,
			Value:                      &value,
			Signature:                  &signatureStr,
		}
		// 将交易请求对象转换为 JSON 字符串
		m, _ := json.Marshal(bt)
		color.Green("提交给BlockServer交易:%s", m)
		buf := bytes.NewBuffer(m)
		// 创建请求体，并向 BlockServer 提交交易请求
		resp, _ := http.Post(ws.Gateway()+"/transactions", "application/json", buf)

		if resp.StatusCode == 201 {
			// 201是哪里来的？请参见blockserver  Transactions方法的  w.WriteHeader(http.StatusCreated)语句
			io.WriteString(w, string(utils.JsonStatus("success")))
			return
		}
		io.WriteString(w, string(utils.JsonStatus("fail")))

	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Println("ERROR: 非法的HTTP请求方式")
	}
}

func (ws *WalletServer) WalletAmount(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("Call WalletAmount  METHOD:%s\n", req.Method)
	if req.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}
	switch req.Method {
	case http.MethodPost:

		var data map[string]interface{}
		// 解析JSON数据

		err := json.NewDecoder(req.Body).Decode(&data)
		if err != nil {
			http.Error(w, "无法解析JSON数据", http.StatusBadRequest)
			return
		}

		// 获取JSON字段的值
		blockchainAddress := data["blockchain_address"].(string)
		color.Blue("请求查询账户%s的余额", blockchainAddress)

		// 构建请求数据
		requestData := struct {
			BlockchainAddress string `json:"blockchain_address"`
		}{
			BlockchainAddress: blockchainAddress,
		}

		// 将请求数据编码为JSON
		jsonData, err := json.Marshal(requestData)
		if err != nil {
			fmt.Printf("编码JSON时发生错误:%v", err)
			return
		}

		bcsResp, _ := http.Post(ws.Gateway()+"/amount", "application/json", bytes.NewBuffer(jsonData))

		//返回给客户端
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Add("Content-Type", "application/json")
		if bcsResp.StatusCode == 200 {
			decoder := json.NewDecoder(bcsResp.Body)
			var bar block.AmountResponse
			err := decoder.Decode(&bar)
			if err != nil {
				log.Printf("ERROR: %v", err)
				io.WriteString(w, string(utils.JsonStatus("fail")))
				return
			}

			resp_message := struct {
				Message string `json:"message"`
				Amount  uint64 `json:"amount"`
			}{
				Message: "success",
				Amount:  bar.Amount,
			}
			m, _ := json.Marshal(resp_message)
			io.WriteString(w, string(m[:]))
		} else {
			io.WriteString(w, string(utils.JsonStatus("fail")))
		}
	default:
		log.Printf("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (ws *WalletServer) walletGetBlockChain(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("Call GetChain  METHOD:%s\n", req.Method)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//设置允许的方法
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	switch req.Method {
	case http.MethodGet:
		resp, err := http.Get(ws.Gateway() + "/")
		if err != nil {
			log.Printf("ERROR: Failed to send GET request: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("ERROR: Unexpected response status code: %v", resp.StatusCode)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Read the response body
		responseBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("ERROR: Failed to read response body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Set response headers
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Write the response body to the client
		_, err = w.Write(responseBody)
		if err != nil {
			log.Printf("ERROR: Failed to write response: %v", err)
			return
		}

	default:
		log.Printf("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}

}
func (ws *WalletServer) walletByNumCheckBlock(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//设置允许的方法
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	switch req.Method {
	case http.MethodPost:

		number := req.FormValue("num")

		// 构建请求体
		body := strings.NewReader("num=" + number)

		// 发起 POST 请求
		resp, err := http.Post(ws.Gateway()+"/getblockbynum", "application/x-www-form-urlencoded", body)
		if err != nil {
			// 处理错误
			fmt.Println("请求错误:", err)
			return
		}
		defer resp.Body.Close()

		// 读取响应内容
		responseBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// 处理错误
			fmt.Println("读取响应错误:", err)
			return
		}
		// 将响应内容写入响应对象
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(responseBody))

	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Println("ERROR: Invalid HTTP Method")
	}
}
func (ws *WalletServer) WalletGetBlockByHash(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//设置允许的方法
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	switch req.Method {
	case http.MethodPost:
		GetBlockByHash := req.FormValue("hashBlock")

		// 构建请求体
		body := strings.NewReader("hashBlock=" + GetBlockByHash)
		resp, err := http.Post(ws.Gateway()+"/getBlockByHash", "application/x-www-form-urlencoded", body)
		if err != nil {
			// 处理错误
			fmt.Println("请求错误:", err)
			return
		}
		defer resp.Body.Close()

		// 读取响应内容
		responseBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// 处理错误
			fmt.Println("读取响应错误:", err)
			return
		}
		// 将响应内容写入响应对象
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(responseBody))
	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Println("ERROR: Invalid HTTP Method")

	}
}
func (ws *WalletServer) WalletGetTransByHash(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//设置允许的方法
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	switch req.Method {
	case http.MethodPost:
		GetTransactionByHash := req.FormValue("transactionHash")

		// 构建请求体
		body := strings.NewReader("transactionHash=" + GetTransactionByHash)
		resp, err := http.Post(ws.Gateway()+"/getTransactionByHash", "application/x-www-form-urlencoded", body)
		if err != nil {
			// 处理错误
			fmt.Println("请求错误:", err)
			return
		}
		defer resp.Body.Close()

		// 读取响应内容
		responseBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// 处理错误
			fmt.Println("读取响应错误:", err)
			return
		}
		// 将响应内容写入响应对象
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(responseBody))
	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Println("ERROR: Invalid HTTP Method")

	}

}

func (ws *WalletServer) Run() {

	fs := http.FileServer(http.Dir("walletServer/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", ws.Index)
	http.HandleFunc("/wallet", ws.Wallet)
	http.HandleFunc("/walletByPrivatekey", ws.walletByPrivatekey)
	http.HandleFunc("/transaction", ws.CreateTransaction)
	http.HandleFunc("/wallet/amount", ws.WalletAmount)
	http.HandleFunc("/wallet/walletByNumCheckBlock", ws.walletByNumCheckBlock)
	http.HandleFunc("/wallet/walletGetBlockChain", ws.walletGetBlockChain)
	http.HandleFunc("/wallet/walletGetBlockByHash", ws.WalletGetBlockByHash)
	http.HandleFunc("/wallet/walletGetTransByHash", ws.WalletGetTransByHash)

	log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(ws.Port())), nil))
}
