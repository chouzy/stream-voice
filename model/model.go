package model

type Request struct {
	Type int    `json:"type"`
	Data []byte `json:"data"`
}

type Response struct {
	Statue struct {
		Code   int    `json:"code"`
		ErrMsg string `json:"err_msg"`
	} `json:"statue"`
}
