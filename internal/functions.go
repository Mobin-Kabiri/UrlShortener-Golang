package internal

import (
	"errors"
	"math"
	"net/url"
	"strings"
)

var (
	ErrInvalidProtocol = errors.New("NORMALIZER - provided url must be in http/https protocol")
	ErrBadFormat       = errors.New("NORMALIZER - provided url is not in standard format.")
	ErrEmptyHost       = errors.New("NORMALIZER - url's host is empty")
	ErrEmptyInput      = errors.New("NORMALIZER - provided url is empty")
	ErrMaxUrlReached   = errors.New("ENCODER - there is no space for a new url because of counter")
)

// 62 valid charactars (for base62)
const validChar string = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func EncodeToBase62String(number uint32) (string, error) {

	if number == math.MaxUint32 {
		return "", ErrMaxUrlReached
	}

	if number == 0 {
		return string(validChar[0]), nil
	}
	// unint8 = element's range is 0-61
	tmpArr := make([]uint8, 0, 8)
	for number != 0 {
		remainder := number % 62
		tmpArr = append(tmpArr, uint8(remainder))
		number = number / 62
	}

	// it can only contains ASCII so we can use byte
	result := make([]byte, 0, len(tmpArr))
	for i := len(tmpArr) - 1; i >= 0; i-- {
		result = append(result, validChar[tmpArr[i]])
	}

	return string(result), nil
}

// returning error to acknowledge the problem
func Normalize(longURL string) (string, error) {

	trimmed := strings.TrimSpace(longURL)
	if trimmed == "" {
		return "", ErrEmptyInput
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", ErrBadFormat
	}

	// we can only lowercase these parts
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)

	protocol := parsed.Scheme
	if protocol != "http" && protocol != "https" {
		return "", ErrInvalidProtocol
	}

	if parsed.Host == "" {
		return "", ErrEmptyHost
	}

	// fragment is a local section and it does not represent main resource
	parsed.Fragment = ""

	// we do not need the port because of knowing the default protocol port
	if (parsed.Scheme == "http" && parsed.Port() == "80") ||
		(parsed.Scheme == "https" && parsed.Port() == "443") {
		parsed.Host = parsed.Hostname()
	}

	// removing the last slash if the url has some path
	if len(parsed.Path) > 1 && strings.HasSuffix(parsed.Path, "/") {
		parsed.Path = strings.TrimRight(parsed.Path, "/")
	}

	// saving same address if it have / at the end or not
	if parsed.Path == "" {
		parsed.Path = "/"
	}

	// sorting the query, example:
	// .../get/userid=55,productid=1 == .../get/productid=1,userid=55
	if parsed.RawQuery != "" {
		parsed.RawQuery = parsed.Query().Encode()
	}

	return parsed.String(), nil

}
