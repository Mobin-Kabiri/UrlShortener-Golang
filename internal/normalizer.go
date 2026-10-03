package internal

import (
	"errors"
	"net/url"
	"strings"
)

var (
	ErrInvalidProtocol = errors.New("NORMALIZER - provided url must be in http/https protocol")
	ErrBadFormat  = errors.New("NORMALIZER - provided url is not in standard format.")
	ErrEmptyHost = errors.New("NORMALIZER - url's host is empty")
	ErrEmptyInput  = errors.New("NORMALIZER - provided url is empty")
)

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

	// sorting the query, example:
	// .../get/userid=55,productid=1 == .../get/productid=1,userid=55
	if parsed.RawQuery != "" {
		parsed.RawQuery = parsed.Query().Encode()
	}

	return parsed.String(), nil

}