package main

import "net/http"

func HttpError(w http.ResponseWriter, code int) {
	http.Error(w, "Error: "+http.StatusText(code), code)
}
