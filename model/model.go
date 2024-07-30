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
	Data AsrRespData `json:"data"`
}

type AsrRespData struct {
	Sid     string `json:"sid"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Result Result `json:"result"`
		Status int    `json:"status"`
	} `json:"data"`
}

type Result struct {
	Ls  bool   `json:"ls"`
	Rg  []int  `json:"rg"`
	Sn  int    `json:"sn"`
	Pgs string `json:"pgs"`
	Ws  []Ws   `json:"ws"`
}

type Ws struct {
	Bg int  `json:"bg"`
	Cw []Cw `json:"cw"`
}

type Cw struct {
	Sc int    `json:"sc"`
	W  string `json:"w"`
}
