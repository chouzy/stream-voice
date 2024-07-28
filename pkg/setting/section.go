package setting

type Ast struct {
	HostUrl   string `json:"host_url" yaml:"hostUrl"`
	Appid     string `json:"appid" yaml:"appid"`
	ApiSecret string `json:"api_secret" yaml:"apiSecret"`
	ApiKey    string `json:"api_key" yaml:"apiKey"`
}
