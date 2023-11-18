// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package main

import "net/http"

func HttpError(w http.ResponseWriter, code int) {
	http.Error(w, "Error: "+http.StatusText(code), code)
}
