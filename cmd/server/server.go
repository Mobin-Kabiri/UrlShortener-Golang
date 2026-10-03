package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"strings"
	"url-shortener/internal/store"
)

var storeObject *store.Store
var baseUrl string

func getCode(w http.ResponseWriter, r *http.Request) {
    
	var req CreateCodeRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON Body - Request body does not matched", http.StatusBadRequest)
		return
	}

    code,err := storeObject.GetShortCode(req.Url)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
		return
    }

    shortUrl := baseUrl + "/" + code
    data := CreateCodeResponse{
        Code:     code,
        ShortURL: shortUrl,
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(data)
}

func getLongUrl(w http.ResponseWriter, r *http.Request) {

    code := r.PathValue("code")
    if code == "" {
        http.Error(w, "Short code is required", http.StatusBadRequest)
        return
    }

    longUrl,err := storeObject.GetLongUrl(code)
    if err != nil {
		if errors.Is(err, store.ErrUrlNotFound) {
			http.Error(w, "URL not found", http.StatusNotFound) // 404
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError) // 500
		return
	}

    // w.Header().Set("Location", longUrl)
    // w.WriteHeader(http.StatusFound) // 302

    //this is standard approach rather than upper code
    http.Redirect(w, r, longUrl, http.StatusFound)
}


func main() {

	addrFlag := flag.String("addr", ":8080", "service port")
	baseFlag := flag.String("base", "http://localhost:8080", "service url")
	flag.Parse()

    baseUrl = strings.TrimRight(*baseFlag, "/")

    //init temp db
    storeObject = store.New()

    // requests format
    mux := http.NewServeMux()
    mux.HandleFunc("POST /api/shorten", getCode)
    mux.HandleFunc("GET /{code}", getLongUrl)

    // log the errors
    if err := http.ListenAndServe(*addrFlag, mux);
    err != nil {
    log.Fatal(err)
    }
}