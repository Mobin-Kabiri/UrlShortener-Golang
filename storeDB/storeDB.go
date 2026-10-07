package storeDB

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
	"url-shortener/internal"
)

type UrlRow struct {
	ID        uint      `gorm:"primaryKey"`
	Code      string    `gorm:"uniqueIndex;not null"`
	LongURL   string    `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type SqlStore struct {
	db      *gorm.DB
	mu      sync.Mutex
	counter uint32
}
const INITIAL_COUNTER = 916132832

func New(db *gorm.DB) (*SqlStore, error) {

	if err := db.AutoMigrate(&UrlRow{}); err != nil {
		return nil, errors.New("failed to migrate")
	}

	var lastRecord UrlRow
	counter := uint32(INITIAL_COUNTER)

	// Getting last record id
	err := db.Order("id desc").First(&lastRecord).Error
	if err == nil {
		counter = INITIAL_COUNTER + uint32(lastRecord.ID)
		return &SqlStore{
				db: db,
				counter: counter,
			}, nil
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return &SqlStore{
				db: db,
				counter: counter,
			}, nil
	} else {
		return nil, errors.New("failed to initialize")
	}	
}

func (s *SqlStore) GetLongUrl(code string) (internal.UrlInfo, error) {
	if len(code) > 8 || len(code) < 6 {
		return internal.UrlInfo{}, fmt.Errorf("%w: %w", internal.ErrNotFound, internal.ErrCodeInvalidLength)
	}

	var record UrlRow
	err := s.db.Where("code = ?", code).First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return internal.UrlInfo{}, internal.ErrNotFound
		}
		return internal.UrlInfo{}, err
	}

	return internal.UrlInfo{
		LongUrl:   record.LongURL,
		CreatedAt: record.CreatedAt,
	}, nil
}

func (s *SqlStore) GetShortCode(longUrl string) (string, error) {
	normalizedUrl, err := internal.Normalize(longUrl)
	if err != nil {
		return "", fmt.Errorf("%w: %w", internal.ErrInvalidURL, err)
	}

	var existing UrlRow
	err = s.db.Where("long_url = ?", normalizedUrl).First(&existing).Error
	if err == nil {
		return existing.Code, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}


	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.db.Where("long_url = ?", normalizedUrl).First(&existing).Error; err == nil {
		return existing.Code, nil
	}

	encodedCode, err := internal.EncodeToBase62String(s.counter)
	if err != nil {
		return "", err
	}

	record := UrlRow{
		Code:      encodedCode,
		LongURL:   normalizedUrl,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.db.Create(&record).Error; err != nil {
		return "", fmt.Errorf("failed to save in db: %w", err)
	}

	s.counter++
	return encodedCode, nil
}