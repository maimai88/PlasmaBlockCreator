package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/matterinc/PlasmaCommons/transaction"
	"github.com/valyala/fasthttp"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	common "github.com/ethereum/go-ethereum/common"
	"github.com/matterinc/PlasmaBlockCreator/foundationdb"
)

type listUTXOsRequest struct {
	For               string `json:"for"`
	BlockNumber       int    `json:"blockNumber"`
	TransactionNumber int    `json:"transactionNumber"`
	OutputNumber      int    `json:"outputNumber"`
	Limit             int    `json:"limit,omitempty"`
}

type singleUTXOdetails struct {
	BlockNumber       int    `json:"blockNumber"`
	TransactionNumber int    `json:"transactionNumber"`
	OutputNumber      int    `json:"outputNumber"`
	Value             string `json:"value"`
}

type listUTXOsResponse struct {
	Error bool                `json:"error"`
	UTXOs []singleUTXOdetails `json:"utxos"`
}

type ListUTXOsHandler struct {
	db         *fdb.Database
	utxoLister *foundationdb.UTXOlister
}

func NewListUTXOsHandler(db *fdb.Database) *ListUTXOsHandler {
	lister := foundationdb.NewUTXOlister(db)
	handler := &ListUTXOsHandler{db, lister}
	return handler
}

func (h *ListUTXOsHandler) HandlerFunc(ctx *fasthttp.RequestCtx) {
	var requestJSON listUTXOsRequest
	err := json.Unmarshal(ctx.PostBody(), &requestJSON)
	if err != nil {
		writeEmptyFasthttpResponse(ctx)
		return
	}

	forBytes := common.FromHex(requestJSON.For)
	address := common.Address{}
	copy(address[:], forBytes)
	blockNumber := uint32(requestJSON.BlockNumber)
	transactionNumber := uint32(requestJSON.TransactionNumber)
	outputNumber := uint8(requestJSON.OutputNumber)
	limit := 50
	if requestJSON.Limit != 0 {
		limit = requestJSON.Limit
	}
	if limit > 100 {
		limit = 100
	}
	// limit := 0
	utxos, err := h.utxoLister.GetUTXOsForAddress(address, blockNumber, transactionNumber, outputNumber, limit, false)
	if err != nil {
		writeEmptyFasthttpResponse(ctx)
		return
	}
	details := make([]singleUTXOdetails, len(utxos))
	for i, utxo := range utxos {
		detail := transaction.ParseIndexIntoUTXOdetails(utxo)
		responseDetails := singleUTXOdetails{int(detail.BlockNumber), int(detail.TransactionNumber),
			int(detail.OutputNumber), detail.Value}
		details[i] = responseDetails
	}
	writeFasthttpResponse(ctx, details)
	return
}

func writeEmptyResponse(w http.ResponseWriter) {
	response := listUTXOsResponse{false, []singleUTXOdetails{}}
	json.NewEncoder(w).Encode(response)

}

func writeResponse(w http.ResponseWriter, details []singleUTXOdetails) {
	response := listUTXOsResponse{false, details}
	json.NewEncoder(w).Encode(response)
}

func writeEmptyFasthttpResponse(ctx *fasthttp.RequestCtx) {
	response := listUTXOsResponse{false, []singleUTXOdetails{}}
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	ctx.Response.Header.Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	body, _ := json.Marshal(response)
	ctx.SetBody(body)

}

func writeFasthttpResponse(ctx *fasthttp.RequestCtx, details []singleUTXOdetails) {
	response := listUTXOsResponse{false, details}
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	ctx.Response.Header.Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	body, _ := json.Marshal(response)
	ctx.SetBody(body)
}
