package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"strings"
	"time"
	"url-shortener/internal"
	"url-shortener/store"
)

type server struct {
	store   StoreInferface
	baseUrl string
}

// naming storeinterface to not match with store package
type StoreInferface interface {
	GetShortCode(url string) (code string, err error)
	GetLongUrl(code string) (urlObj internal.UrlInfo, err error)
}


func (s *server) getCode(w http.ResponseWriter, r *http.Request) {

	var req CreateCodeRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON Body - Request body does not matched", http.StatusBadRequest)
		return
	}
	
	code, err := s.store.GetShortCode(req.Url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortUrl := s.baseUrl + "/" + code
	data := CreateCodeResponse{
		Code:     code,
		ShortURL: shortUrl,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(data)
}

func (s *server) getLongUrl(w http.ResponseWriter, r *http.Request) {

	code := r.PathValue("code")
	if code == "" {
		http.Error(w, "Short code is required", http.StatusBadRequest)
		return
	}

	urlObj, err := s.store.GetLongUrl(code)
	if err != nil {
		if errors.Is(err, internal.ErrUrlNotFound) {
			http.Error(w, "URL not found", http.StatusNotFound) // 404
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError) // 500
		return
	}

	// w.Header().Set("Location", longUrl)
	// w.WriteHeader(http.StatusFound) // 302

	//this is standard approach rather than upper code
	http.Redirect(w, r, urlObj.LongUrl, http.StatusFound)
}

func (s *server) getLongUrlObj(w http.ResponseWriter, r *http.Request) {

	code := r.PathValue("code")
	if code == "" {
		http.Error(w, "Short code is required", http.StatusBadRequest)
		return
	}

	urlObj, err := s.store.GetLongUrl(code)
	if err != nil {
		if errors.Is(err, internal.ErrUrlNotFound) {
			http.Error(w, "URL not found", http.StatusNotFound) // 404
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError) // 500
		return
	}

	data := GetUrlObjResponse{
		LongUrl:     urlObj.LongUrl,
		CreatedAt: urlObj.CreatedAt.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func setupServer(s *server) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/shorten", s.getCode)
	mux.HandleFunc("GET /{code}", s.getLongUrl)
	mux.HandleFunc("GET /api/v1/links/{code}", s.getLongUrlObj)
	return mux
}


func main() {

	srv := server{
		store:   store.New(),
		baseUrl: "",
	}
	addrFlag := flag.String("addr", ":8080", "service port")
	baseFlag := flag.String("base", "http://localhost:8080", "service url")
	flag.Parse()

	baseUrl := strings.TrimRight(*baseFlag, "/")
	srv.baseUrl = baseUrl

	mux := setupServer(&srv)

	// log the errors
	if err := http.ListenAndServe(*addrFlag, mux); err != nil {
		log.Fatal(err)
	}
}
