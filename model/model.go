package model

type Request struct {
	Data   string `json:"data"`
	IsLast bool   `json:"isLast"`
}

type Response struct {
	Statue struct {
		Code   int    `json:"code"`
		ErrMsg string `json:"err_msg"`
	} `json:"statue"`
	Data string `json:"data"`
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

func (t *Result) String() string {
	var wss string
	for _, v := range t.Ws {
		wss += v.String()
	}
	return wss
}

type Ws struct {
	Bg int  `json:"bg"`
	Cw []Cw `json:"cw"`
}

func (w *Ws) String() string {
	var wss string
	for _, v := range w.Cw {
		wss += v.W
	}
	return wss
}

type Cw struct {
	Sc int    `json:"sc"`
	W  string `json:"w"`
}
