package main

import (
	"net/http"
)

func newStaticHandler(options *Options) http.HandlerFunc {
	imageDir := options.StaticDir

	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("file")
		http.ServeFile(w, r, imageDir+"/"+slug)
	}
}
