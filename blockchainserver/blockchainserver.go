package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/myrachanto/blockchain/block"
	"github.com/myrachanto/blockchain/utils"
	"github.com/myrachanto/blockchain/wallet"
)

var cache map[string]*block.Blockchain = make(map[string]*block.Blockchain)

type BlockchainServer struct {
	port uint16
}

func NewBlockchainServer(port uint16) *BlockchainServer {
	return &BlockchainServer{port}
}
func (bcs *BlockchainServer) Port() uint16 {
	return bcs.port
}
func (bcs *BlockchainServer) GetChain(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		w.Header().Add("Content-Type", "application/json")
		bc := bcs.GetBlockchain()
		m, _ := bc.MarshalJSON()
		io.WriteString(w, string(m[:]))
	default:
		log.Printf("Error: Invalid HTTp method")
	}
}

func (bcs *BlockchainServer) Transaction(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Add("Content-Type", "application/json")
		bc := bcs.GetBlockchain()
		transactions := bc.TransactionPool()
		m, err := json.Marshal(struct {
			Transactions []*block.Transaction
			Length       int
		}{
			Transactions: transactions,
			Length:       len(transactions),
		})
		if err != nil {
			log.Printf("Error: %v ", err)
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}
		io.WriteString(w, string(m[:]))
	case http.MethodPost:
		decoder := json.NewDecoder(r.Body)
		var t block.TransactionRequest
		err := decoder.Decode(&t)
		if err != nil {
			log.Printf("Error: %v ", err)
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}
		if !t.Validate() {
			log.Println("Error: Missings field(s)")
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}
		publicKey := utils.PublicKeyFromString(*t.SenderPublicKey)
		signature := utils.SignatureFromString(*t.Signature)
		bc := bcs.GetBlockchain()
		isCreated := bc.CreateTransaction(*t.SenderBlockchainAddress, *t.RecipientBlockchainAddress, *t.Value, publicKey, signature)
		w.Header().Add("Content-Type", "application/json")
		var m []byte
		if !isCreated {
			w.WriteHeader(http.StatusBadRequest)
			m = utils.JsonStatus("fail")
		} else {

			w.WriteHeader(http.StatusCreated)
			m = utils.JsonStatus("success")
		}
		io.WriteString(w, string(m))

	case http.MethodPut:
		decoder := json.NewDecoder(r.Body)
		var t block.TransactionRequest
		err := decoder.Decode(&t)
		if err != nil {
			log.Printf("Error: %v ", err)
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}
		if !t.Validate() {
			log.Println("Error: Missings field(s)")
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}
		publicKey := utils.PublicKeyFromString(*t.SenderPublicKey)
		signature := utils.SignatureFromString(*t.Signature)
		bc := bcs.GetBlockchain()
		Isupdated := bc.AddTransaction(*t.SenderBlockchainAddress, *t.RecipientBlockchainAddress, *t.Value, publicKey, signature)
		w.Header().Add("Content-Type", "application/json")
		var m []byte
		if !Isupdated {
			w.WriteHeader(http.StatusBadRequest)
			m = utils.JsonStatus("fail")
		} else {

			w.WriteHeader(http.StatusCreated)
			m = utils.JsonStatus("success")
		}
		io.WriteString(w, string(m))
	case http.MethodDelete:
		bcs.GetBlockchain().ClearTransactionPool()
		io.WriteString(w, string(utils.JsonStatus("success")))
	default:
		log.Println("Error: Invalid Http Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}
func (bcs *BlockchainServer) GetBlockchain() *block.Blockchain {
	bc, ok := cache["blockchain"]
	if !ok {
		minersWallet := wallet.NewWallet()
		bc = block.NewBlockchain(minersWallet.BlockchainAddress(), bcs.Port())
		cache["blockchain"] = bc
		log.Printf("private key %v", minersWallet.PrivateKeyStr())
		log.Printf("public key %v", minersWallet.PublicKeyStr())
		log.Printf("Blockchain address %v", minersWallet.BlockchainAddress())
	}
	return bc
}
func (bcs *BlockchainServer) Mine(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		bc := bcs.GetBlockchain()
		isMined := bc.Mining()
		var m []byte
		if !isMined {
			w.WriteHeader(http.StatusBadRequest)
			m = utils.JsonStatus("fail")
		} else {
			m = utils.JsonStatus("success")

		}
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(m))
	default:
		log.Println("Error: Invalid Http Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}
func (bcs *BlockchainServer) StartMine(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		bc := bcs.GetBlockchain()
		bc.StartMining()
		var m []byte
		m = utils.JsonStatus("success")
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(m))
	default:
		log.Println("Error: Invalid Http Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) Amount(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		blockchainaddress := r.URL.Query().Get("blockchain_address")
		amount := bcs.GetBlockchain().CalculateTotalAmount(blockchainaddress)
		ar := &block.AmountResponse{amount}
		m, _ := ar.MarshalJSON()
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(m))
		// log.Println(">>>>>>>>>>>", string(m[:]))
	default:
		log.Println("Error: Invalid Http Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}
func (bcs *BlockchainServer) Run() {
	bcs.GetBlockchain().Run()
	http.HandleFunc("/", bcs.GetChain)
	http.HandleFunc("/transactions", bcs.Transaction)
	http.HandleFunc("/mine", bcs.Mine)
	http.HandleFunc("/amount", bcs.Amount)
	http.HandleFunc("/mine/start", bcs.StartMine)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(bcs.Port())), nil))
}
