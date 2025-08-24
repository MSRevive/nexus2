package response

import (
	"net/http"

	json "github.com/sugawarayuuta/sonnet"
)

//make public so other packages can create custom responses.
type Response struct {
	Status bool `json:"status"`
	Code int `json:"code"`
	Error string `json:"error,omitempty"`

	Data interface{} `json:"data,omitempty"`
	
	w http.ResponseWriter
}

func (r Response) SendJson() {
	r.w.Header().Set("Content-Type", "application/json")
	r.w.WriteHeader(r.Code)
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

func Created(w http.ResponseWriter, data interface{}) {
	resp := Response{
		Status: true,
		Code: http.StatusCreated,
		Error: "",
		Data: data,
		w: w,
	}
	resp.SendJson()
}

func StillProcessing(w http.ResponseWriter, data interface{}) {
	resp := Response{
		Status: true,
		Code: http.StatusAccepted,
		Error: "",
		Data: data,
		w: w,
	}
	resp.SendJson()
}

func OKNoContent(w http.ResponseWriter) {
	resp := Response{
		Status: true,
		Code: http.StatusNoContent,
		Error: "",
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

func OKChar(w http.ResponseWriter, data interface{}) {
	resp := Response{
		Status: true,
		Code: http.StatusOK,
		Error: "",
		Data: data,
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

func GenericError(w http.ResponseWriter) {
	resp := Response{
		Status: false,
		Code: http.StatusInternalServerError,
		Error: "Internal server error!",
		Data: nil,
		w: w,
	}
	resp.SendJson()
}

func DepreciatedError(w http.ResponseWriter) {
	resp := Response{
		Status: false,
		Code: http.StatusUpgradeRequired,
		Error: "Server is out of date!",
		Data: nil,
		w: w,
	}
	resp.SendJson()
}

func NotAvailable(w http.ResponseWriter) {
	resp := Response{
		Status: false,
		Code: http.StatusNotImplemented,
		Error: "Not available!",
		w: w,
	}
	resp.SendJson()
}