package conf

import "github.com/tidwall/gjson"

type Config struct {
	Host      string // Listen host
	Port      int    // Listen Port
	Proxy     string
	SecretKey string
	setting   gjson.Result
}
