package store

import (
	"fmt"
	"math"
	"sync"
	"time"
	"url-shortener/internal"
)

type Store struct {
	counter       uint32
	mu            sync.RWMutex // rw allows us to read parallel
	longUrlToCode map[string]string
	codeToLongUrl map[string]internal.UrlInfo
}

func New() (s *Store) {
	return &Store{
		counter:       916132832, // its 62^5 to get at least len=6 shortCode
		longUrlToCode: make(map[string]string),
		codeToLongUrl: make(map[string]internal.UrlInfo),
	}
}

func NewForTesting() (s *Store) {
	return &Store{
		counter:       math.MaxUint32,
		longUrlToCode: make(map[string]string),
		codeToLongUrl: make(map[string]internal.UrlInfo),
	}
}

func (s *Store) GetLongUrl(code string) (internal.UrlInfo, error) {
	if len(code) > 8 || len(code) < 6 {
		return internal.UrlInfo{}, fmt.Errorf("%w: %w", internal.ErrNotFound, internal.ErrCodeInvalidLength)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	urlObj, exist := s.codeToLongUrl[code]
	if !exist {
		return internal.UrlInfo{}, internal.ErrNotFound
	}

	return urlObj, nil
}

func (s *Store) GetShortCode(longUrl string) (string, error) {
	normalizedUrl, err := internal.Normalize(longUrl)
	if err != nil {
		// err could be: ErrInvalidProtocol, ErrBadFormat, ErrEmptyHost, ErrEmptyInput
		return "", fmt.Errorf("%w: %w", internal.ErrInvalidURL, err)
	}

	// this section is for first reading try - just read lock ---
	s.mu.RLock()
	code, exist := s.longUrlToCode[normalizedUrl]
	if exist {
		s.mu.RUnlock()
		return code, nil
	}
	s.mu.RUnlock()
	// ---

	s.mu.Lock()
	defer s.mu.Unlock()

	// double check: if another goroutine has used same url between rlock and lock
	if code, exist := s.longUrlToCode[normalizedUrl]; exist {
		return code, nil
	}

	encodedCode, err := internal.EncodeToBase62String(s.counter)
	if err != nil {
		return "", err
	}
	s.counter += 1
	s.longUrlToCode[normalizedUrl] = encodedCode
	urlObj := internal.UrlInfo{LongUrl: normalizedUrl, CreatedAt: time.Now().UTC()}
	s.codeToLongUrl[encodedCode] = urlObj

	return encodedCode, nil
}
