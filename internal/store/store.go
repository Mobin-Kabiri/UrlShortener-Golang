package store

import (
	"errors"
	"sync"
	"url-shortener/internal"
)

var (
	ErrUrlNotFound = errors.New("STORE - url not found ")
)

type Store struct {
	counter uint32
	mu sync.RWMutex // rw allows us to read parallel
	longUrlToCode map[string]string
	codeToLongUrl map[string]string
}

func New() (s *Store) {
	return &Store{
        counter:       916132832, // its 62^5 to get at least len=6 shortCode
        longUrlToCode: make(map[string]string),
        codeToLongUrl: make(map[string]string),
    }
}

func (s *Store) GetLongUrl(code string) (string,error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	longUrl, exist := s.codeToLongUrl[code]
	if(!exist){
		return "", ErrUrlNotFound
	}
	return longUrl, nil
}

func (s *Store) GetShortCode(longUrl string) (string,error) {
	normalizedUrl, err := internal.Normalize(longUrl)
	if err!=nil{
		return "",err
	}

	s.mu.RLock()
	code, exist := s.longUrlToCode[normalizedUrl]
	if(exist){
		s.mu.RUnlock()
		return code, nil
	}
	s.mu.RUnlock()
	
	s.mu.Lock()
	defer s.mu.Unlock()

	// double check: if another goroutine has used same url between rlock and lock
	if code, exist := s.longUrlToCode[normalizedUrl]; exist {
		return code, nil
	}

	encodedCode := internal.EncodeToBase62String(s.counter)
	s.counter += 1
	s.longUrlToCode[normalizedUrl] = encodedCode
	s.codeToLongUrl[encodedCode] = normalizedUrl

	return encodedCode, nil
}