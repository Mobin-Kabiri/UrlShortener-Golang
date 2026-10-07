package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"strings"
	"time"
	"url-shortener/storeDB"
	"url-shortener/internal"
	"url-shortener/store"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// based on cloudflare maximum url length
const MAX_URL_LENGTH = 32768

// timeout enviroment
const READ_TIMEOUT_SEC = 5
const READ_HEADER_TIMEOUT_SEC = 3
const WRITE_TIMEOUT_SEC = 10
const IDLE_TIMOUT_SEC = 60
const MAX_HEADER_SIZE = 16384

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

	r.Body = http.MaxBytesReader(w, r.Body, MAX_URL_LENGTH)

	var req CreateCodeRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON Body - Request body does not matched", http.StatusBadRequest)
		return
	}

	code, err := s.store.GetShortCode(req.Url)
	if err != nil {
		switch {
		case errors.Is(err, internal.ErrInvalidURL):
			http.Error(w, "Given url is invalid", http.StatusBadRequest) // 400
		case errors.Is(err, internal.ErrMaxUrlReached):
			http.Error(w, "Service is not available for now", http.StatusServiceUnavailable) // 503
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError) // 500
		}
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

	urlObj, err := s.store.GetLongUrl(code)
	if err != nil {
		if errors.Is(err, internal.ErrNotFound) {
			http.Error(w, "URL not found", http.StatusNotFound) // 404
			return
		} else if errors.Is(err, internal.ErrCodeInvalidLength) {
			http.Error(w, "Given code is not valid", http.StatusBadRequest) // 400
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

	urlObj, err := s.store.GetLongUrl(code)
	if err != nil {
		if errors.Is(err, internal.ErrNotFound) {
			http.Error(w, "URL not found", http.StatusNotFound) // 404
			return
		} else if errors.Is(err, internal.ErrCodeInvalidLength) {
			http.Error(w, "Given code is not valid", http.StatusBadRequest) // 400
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError) // 500
		return
	}

	data := GetUrlObjResponse{
		LongUrl:   urlObj.LongUrl,
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

func setupHttpServerConfiguration(addr string, mux *http.ServeMux) error {

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadTimeout:       READ_TIMEOUT_SEC * time.Second,
		ReadHeaderTimeout: READ_HEADER_TIMEOUT_SEC * time.Second,
		WriteTimeout:      WRITE_TIMEOUT_SEC * time.Second,
		IdleTimeout:       IDLE_TIMOUT_SEC * time.Second,
		MaxHeaderBytes:    MAX_HEADER_SIZE,
	}

	log.Println("Server is running ...")
	return httpServer.ListenAndServe()
}


func initDb() *storeDB.SqlStore{
	db, err := gorm.Open(sqlite.Open("myDatabase.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql db: %v", err)
	}

	// one connection avoids db locked err
	sqlDB.SetMaxOpenConns(1)

	dbStore, err := storeDB.New(db)
	if err != nil {
		log.Fatalf("failed to initialize sql store: %v", err)
	}

	return dbStore
}

func main() {


	addrFlag := flag.String("addr", ":8080", "service port")
	baseFlag := flag.String("base", "http://localhost:8080", "service url")
	storeFlag := flag.String("store", "memory", "service storing option")
	flag.Parse()

	var storeOption StoreInferface
	switch *storeFlag {
	case "memory":
		storeOption = store.New()
	case "db":
		storeOption = initDb()
	default:
		log.Fatal("invalid storing option")
	}

	// the interface is getting 'store' object
	srv := server{
		store:   storeOption,
		baseUrl: "",
	}

	baseUrl := strings.TrimRight(*baseFlag, "/")
	srv.baseUrl = baseUrl

	mux := setupServer(&srv)

	// log the errors
	if err := setupHttpServerConfiguration(*addrFlag, mux); err != nil {
		log.Fatal(err)
	}
}
