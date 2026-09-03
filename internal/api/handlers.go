package api

import (
	"encoding/json"
	"net/http"
)

type OrderRequest struct{
	Asset string `json:"assest"`
	Side string `json:"side"`
	Price float64 `json:"price"`
	Quantity float64 `json:"quantity"`
}

func PlaceOrderHandler(w http.ResponseWriter,req *http.Request){
	
	userID,ok:=req.Context().Value(UserContextKey).(string)
	
	if !ok {
		http.Error(w, "Internal Server Fault: Missing Identity Context", http.StatusInternalServerError)
		return
	}

	req.Body=http.MaxBytesReader(w,req.Body,10240)

	var form OrderRequest
	 
	dec:=json.NewDecoder(req.Body)

	dec.DisallowUnknownFields()

	if err:=dec.Decode(&form);err!=nil{
		http.Error(w, "Malformed JSON or unexpected fields", http.StatusBadRequest)
		return
	}

	if form.Quantity<=0 || form.Price<=0{
		http.Error(w, "Price and Quantity must be strictly positive", http.StatusBadRequest)
		return
	}

	if form.Side!="BUY" && form.Side!="SELL"{
		http.Error(w, "Side must be exactly 'BUY' or 'SELL'", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)

	response:=map[string]string{
		"status":  "Order accepted for processing",
		"user_id": userID,
		"asset":   form.Asset,
	}

	json.NewEncoder(w).Encode(response)




}