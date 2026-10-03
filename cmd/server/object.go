package main

// type BaseResponse struct {
// 	description string `json:"description"`
// }

type CreateCodeRequest struct {
	Url string `json:"url"`
}

type CreateCodeResponse struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
}
