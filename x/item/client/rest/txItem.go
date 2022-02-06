package rest

import (
	//"crypto/sha256"
	//"encoding/hex"
	///"fmt"
	"net/http"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"
	"github.com/trstlabs/trst/x/item/types"
)

// Used to not have an error if strconv is unused
var _ = strconv.Itoa(42)

type createItemRequest struct {
	BaseReq     rest.BaseReq `json:"base_req"`
	Creator     string       `json:"creator"`
	Title       string       `json:"title"`
	Description string       `json:"description"`

	Shippingcost    int64    `json:"shipping_cost"`
	Location        string   `json:"location"`
	Estimationcount int64    `json:"estimation_count"`
	Tags            []string `json:"tags"`
	Condition       int64    `json:"condition"`
	Shippingregion  []string `json:"shipping_region"`
	Depositamount   int64    `json:"deposit_amount"`
	InitMsg         []byte   `json:"init_msg"`
	AutoMsg         []byte   `json:"auto_msg"`
	Photos          []string `json:"photos"`
	TokenUri        string   `json:"token_uri"`
}

func createItemHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createItemRequest
		if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		_, err := sdk.AccAddressFromBech32(req.Creator)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		parsedTitle := req.Title

		parsedDescription := req.Description

		parsedShippingcost := req.Shippingcost

		parsedLocation := req.Location

		parsedEstimationcount := req.Estimationcount
		//var estimationcount = fmt.Sprint(req.Estimationcount)
		//var estimationcountHash = sha256.Sum256([]byte(estimationcount))
		//var estimationcountHashString = hex.EncodeToString(estimationcountHash[:])

		parsedTags := req.Tags

		parsedCondition := req.Condition

		parsedShippingregion := req.Shippingregion

		parsedDepositAmount := req.Depositamount
		parsedMsg := req.InitMsg
		parsedAutoMsg := req.AutoMsg

		parsedPhotos := req.Photos
		parsedTokenUri := req.TokenUri

		msg := types.NewMsgCreateItem(
			req.Creator,
			parsedTitle,
			parsedDescription,
			parsedShippingcost,
			parsedLocation,
			parsedEstimationcount,

			parsedTags,

			parsedCondition,
			parsedShippingregion,
			parsedDepositAmount,
			parsedMsg,
			parsedAutoMsg,
			parsedPhotos,
			parsedTokenUri,
		)

		tx.WriteGeneratedTxResponse(clientCtx, w, req.BaseReq, msg)
	}
}

type deleteItemRequest struct {
	BaseReq rest.BaseReq `json:"base_req"`
	Seller  string       `json:"creator"`
}

func deleteItemHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//id := mux.Vars(r)["id"]

		var req deleteItemRequest
		if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		id, e := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
		if e != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, e.Error())
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		_, err := sdk.AccAddressFromBech32(req.Seller)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewMsgDeleteItem(
			req.Seller,
			id,
		)

		tx.WriteGeneratedTxResponse(clientCtx, w, req.BaseReq, msg)
	}
}
