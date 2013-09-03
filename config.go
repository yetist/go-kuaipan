package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
)

// 判断文件或路径是否存在
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}


//配置项
type Config struct {
    filepath string
	consumer_key       string
	consumer_secret    string
	oauth_token        string
	oauth_token_secret string
}

var config Config

func init() {
	config.Init()
	config.Read()

	flag.StringVar(&config.consumer_key, "consumer_key", config.consumer_key, "consumer_key")
	flag.StringVar(&config.consumer_secret, "consumer_secret", config.consumer_secret, "consumer_secret")
}

func (c *Config) Init() {
    config.consumer_key = "xcwp0kb8kfUnj7WW"
    config.consumer_secret="RlFWN5jYI6H9h5yO"
	_, file, _, ok := runtime.Caller(0)
	if ok {
		c.filepath = path.Join(path.Dir(file), "config.json")
	} else {
		c.filepath = "config.json"
	}
}

func (c *Config) Read() {
	if Exists(c.filepath) {
		file, err := os.Open(c.filepath)
		defer file.Close()
		if err != nil {
			fmt.Println("配置文件读取失败")
			return
		}

		var cfg map[string]string
		dec := json.NewDecoder(file)
		err = dec.Decode(&cfg)

		if err != nil {
			fmt.Println("配置文件读取失败")
			return
		}

		if cfg["consumer_key"] != "" {
			c.consumer_key= cfg["consumer_key"]
		}
		if cfg["consumer_secret"] != "" {
			c.consumer_secret= cfg["consumer_secret"]
		}
		if cfg["oauth_token"] != "" {
			c.oauth_token= cfg["oauth_token"]
		}
		if cfg["oauth_token_secret"] != "" {
			c.oauth_token_secret= cfg["oauth_token_secret"]
		}
	}
}

func (c *Config) Write() (err error) {
	var data []byte
	jsmap := make(map[string]interface{})
	jsmap["consumer_key"] = c.consumer_key
	jsmap["consumer_secret"] = c.consumer_secret
	jsmap["oauth_token"] = c.oauth_token
	jsmap["oauth_token_secret"] = c.oauth_token_secret

	data, err = json.MarshalIndent(jsmap, "", "  ")
	if err != nil {
		return
	}
	fmt.Println(string(data))
	return ioutil.WriteFile(c.filepath, data, 0644)
}
