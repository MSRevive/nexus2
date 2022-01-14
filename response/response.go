package response

import (
	"encoding/json"
	"net/http"
)

type response struct {
  Code int `json:"code"`
  Status bool `json:"status"`
  Error string `json:"error"`
  Data interface{} `json:"data"`
}

type responseResult struct {
	result bool `json:"result"`
}

func Raw(w http.ResponseWriter, status bool, code int, err error, data interface{}) {
  resp := response{
		Status: status,
		Code: code,
		Error: "",
		Data: data,
	}
  
  if err != nil {
		resp.Error = err.Error()
	}
  
	w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(resp)
}

func OK(w http.ResponseWriter, data interface{}) {
  Raw(w, true, http.StatusOK, nil, data)
}

func Result(w http.ResponseWriter, b bool) {
	Raw(w, true, http.StatusOK, nil, b)
}

func BadRequest(w http.ResponseWriter, err error) {
  Raw(w, false, http.StatusBadRequest, err, nil)
}

func Error(w http.ResponseWriter, err error) {
  Raw(w, false, http.StatusInternalServerError, err, nil)
}