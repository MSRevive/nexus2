package response

import (
	"net/http"
	json "github.com/goccy/go-json"
)

//make public so other packages can create custom responses.
type Response struct {
	Code int `json:"code"`
	Status bool `json:"status"`
	Error string `json:"error"`
	Data interface{} `json:"data"`
	IsBanned *bool `json:"isBanned,omitempty"`
	IsAdmin *bool `json:"isAdmin,omitempty"`
	
	w http.ResponseWriter `json:"-"`
}

func (r *Response) SendJson() {
	r.w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(r.w).Encode(r)
}

func TooManyRequests(w http.ResponseWriter) {
	resp := Response{
		Status: false,
		Code: http.StatusTooManyRequests,
		Error: "Too many requests",
		Data: nil,
		w: w,
	}
	resp.SendJson()
}

func InternalServerError(w http.ResponseWriter) {
	resp := Response{
		Status: false,
		Code: http.StatusInternalServerError,
		Error: "Internal server error!",
		Data: nil,
		w: w,
	}
	resp.SendJson()
}

func OK(w http.ResponseWriter, data interface{}) {
	resp := Response{
		Status: true,
		Code: http.StatusOK,
		Error: "",
		Data: data,
		w: w,
	}
	resp.SendJson()
}

func OKChar(w http.ResponseWriter, isBanned bool, isAdmin bool, data interface{}) {
	resp := Response{
		Status: true,
		Code: http.StatusOK,
		Error: "",
		Data: data,
		IsBanned: &isBanned,
		IsAdmin: &isAdmin,
		w: w,
	}
	resp.SendJson()
}

func Result(w http.ResponseWriter, b bool) {
	resp := Response{
		Status: true,
		Code: http.StatusOK,
		Error: "",
		Data: b,
		w: w,
	}
	resp.SendJson()
}

func BadRequest(w http.ResponseWriter, err error) {
	resp := Response{
		Status: false,
		Code: http.StatusBadRequest,
		Error: err.Error(),
		Data: nil,
		w: w,
	}
	resp.SendJson()
}

func Error(w http.ResponseWriter, err error) {
	resp := Response{
		Status: false,
		Code: http.StatusInternalServerError,
		Error: err.Error(),
		Data: nil,
		w: w,
	}
	resp.SendJson()
}