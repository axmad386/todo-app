package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Empty struct {
}

var Blank Empty

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Success(w http.ResponseWriter, data interface{}, code int) bool {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	res := Response{
		Status:  "Success",
		Message: "Success",
		Data:    data,
	}
	js, err := json.Marshal(res)
	if err != nil {
		fmt.Fprint(w, err)
		return false
	}
	w.Write(js)
	return true
}

func NotFound(w http.ResponseWriter, id int, module string, column string) bool {
	if column == "" {
		column = "ID"
	}
	return Fail(w, Response{
		Status:  "Not Found",
		Message: fmt.Sprintf("%s with %s %d Not Found", module, column, id),
	}, http.StatusNotFound)
}

func CantNull(w http.ResponseWriter, input string) bool {
	return Fail(w, Response{
		Status:  "Bad Request",
		Message: fmt.Sprintf("%s cannot be null", input),
	}, http.StatusBadRequest)
}
func Fail(w http.ResponseWriter, data Response, code int) bool {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	res := Response{
		Status:  data.Status,
		Message: data.Message,
		Data:    Blank,
	}
	js, err := json.Marshal(res)
	if err != nil {
		fmt.Fprint(w, err)
		return false
	}
	w.Write(js)
	return true
}
