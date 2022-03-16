package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"strconv"
	"text/template"

	"github.com/myrachanto/blockchain/block"
	"github.com/myrachanto/blockchain/utils"
	"github.com/myrachanto/blockchain/wallet"
)

const tempDir = "templates"

type WalletServer struct {
	port    uint16
	gateway string
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
		// io.WriteString(w, "heelo world")
	default:
		log.Printf("ERROR: Invalid HTTP Method")
	}
}
func (ws *WalletServer) Wallet(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		w.Header().Add("Content-Type", "application/json")
		mywallet := wallet.NewWallet()
		m, _ := mywallet.MarshalJSON()
		io.WriteString(w, string(m[:]))
	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Error: Invalid Http methods")
	}
}

func (ws *WalletServer) CreateTransaction(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		decoder := json.NewDecoder(req.Body)
		var t wallet.TransactionRequest
		err := decoder.Decode(&t)
		if err != nil {
			log.Printf("ERROR: %v", err)
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}
		if !t.Validate() {
			log.Println("ERROR: missing field(s)")
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}

		publicKey := utils.PublicKeyFromString(*t.SenderPublicKey)
		privateKey := utils.PrivateKeyFromString(*t.SenderPrivateKey, publicKey)
		value, err := strconv.ParseFloat(*t.Value, 32)
		if err != nil {
			log.Println("ERROR: parse error")
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}
		value32 := float32(value)

		w.Header().Add("Content-Type", "application/json")
		fmt.Println(privateKey)
		fmt.Println(privateKey)
		fmt.Printf("%.1f\n", value32)
		// mywallet := wallet.NewWallet()
		// m, _ := mywallet.MarshalJSON()
		// io.WriteString(w, string(m[:]))
		transaction := wallet.NewTransaction(privateKey, publicKey, *t.SenderBlockchainAddress, *t.RecipientBlockchainAddress, value32)
		signature := transaction.GenerateSignature()
		signatureStr := signature.String()
		bt := &block.TransactionRequest{
			t.SenderBlockchainAddress,
			t.RecipientBlockchainAddress,
			t.SenderPublicKey,
			&value32,
			&signatureStr,
		}
		m, err := json.Marshal(bt)
		if err != nil {
			log.Println("ERROR: Error marshalling the transaction")
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}
		buf := bytes.NewBuffer(m)
		resp, err := http.Post(ws.Gateway()+"/transactions", "application/json", buf)
		if err != nil {
			log.Println("ERROR: Error connecting with the backend")
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}
		if resp.StatusCode == 201 {
			log.Println("Successfuy created a block")
			io.WriteString(w, string(utils.JsonStatus("success")))
			return
		}
		io.WriteString(w, string(utils.JsonStatus("fail")))
	default:
		log.Printf("Error: Something went wrong eith the transaction!")
	}
}

func (ws *WalletServer) WalletAmount(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		blockachaineAddress := req.URL.Query().Get("blockchain_address")
		// log.Println(">>>>>>>>>>>blockachaineAddress", blockachaineAddress)
		endpoint := fmt.Sprintf("%s/amount", ws.Gateway())
		client := &http.Client{}
		bcsreq, _ := http.NewRequest("GET", endpoint, nil)
		q := bcsreq.URL.Query()
		q.Add("blockchain_address", blockachaineAddress)
		bcsreq.URL.RawQuery = q.Encode()
		bcsresp, err := client.Do(bcsreq)
		if err != nil {
			log.Printf("ERROR: %v", err)
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}
		w.Header().Add("Content-Type", "application/json")

		if bcsresp.StatusCode == 200 {
			log.Println("Successfuy created a block")
			decoder := json.NewDecoder(bcsresp.Body)
			var bar block.AmountResponse
			err := decoder.Decode(&bar)
			if err != nil {
				log.Printf("ERROR: %v", err)
				io.WriteString(w, string(utils.JsonStatus("fail")))
				return
			}
			m, _ := json.Marshal(struct {
				Message string  `json:"message"`
				Amount  float32 `json:"amount"`
			}{
				Message: "success",
				Amount:  bar.Amount,
			})
			log.Println(">>>>>>>>>>>", string(m[:]))
			io.WriteString(w, string(m[:]))
			return
		} else {
			io.WriteString(w, string(utils.JsonStatus("fail")))
		}
	default:
		log.Println("Error: Invalid Http Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}
func (ws *WalletServer) Run() {
	http.HandleFunc("/ms", ws.Index)
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/wallet", ws.Wallet)
	http.HandleFunc("/wallet/amount", ws.WalletAmount)
	http.HandleFunc("/transaction", ws.CreateTransaction)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(ws.Port())), nil))
}
