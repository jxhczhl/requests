package libs

type RequestParams struct {
	Method         string                 `json:"Method"`
	Url            string                 `json:"Url"`
	Params         map[string]string      `json:"Params"`
	Headers        map[string]string      `json:"Headers"`
	Cookies        map[string]string      `json:"Cookies"`
	Data           map[string]string      `json:"Data"`
	Json           map[string]interface{} `json:"Json"`
	Body           string                 `json:"Body"`
	Auth           []string               `json:"Auth"`
	Timeout        int                    `json:"Timeout"`
	AllowRedirects bool                   `json:"AllowRedirects"`
	Proxies        string                 `json:"Proxies"`
	Verify         bool                   `json:"Verify"`
	Cert           []string               `json:"Cert"`
	Ja3            string                 `json:"Ja3"`
	TLSExtensions  string                 `json:"TLSExtensions"`
	HTTP2Settings  string                 `json:"HTTP2Settings"`
}
