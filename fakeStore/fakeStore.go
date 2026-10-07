package fakestore

import (
	"time"
	"url-shortener/internal"
)

type FakeStore struct {
	Code    string
	LongUrl string
	Err     error
}

func (s *FakeStore) GetLongUrl(code string) (internal.UrlInfo, error) {
	return internal.UrlInfo{LongUrl: s.LongUrl, CreatedAt: time.Now()}, s.Err
}

func (s *FakeStore) GetShortCode(longUrl string) (string, error) {
	return s.Code, s.Err
}
