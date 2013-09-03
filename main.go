package main

import (
	"fmt"
	"flag"
	"os"
)

func main() {
	flag.Parse()
	if len(config.consumer_key) == 0 || len(config.consumer_secret) == 0 {
		fmt.Println("You must set the --consumer_key and --consumer_secret flags.")
		fmt.Println("---")
		flag.Usage()
		os.Exit(1)
	}

	kp := NewKuaipan(config.consumer_key, config.consumer_secret)
	kp.Debug(true)
	kp.SetAccessToken(config.oauth_token, config.oauth_token_secret)
	if !kp.Authorized() {
		if ret :=kp.Authorize(); ret {
			token, secret := kp.GetAccessToken()
			config.oauth_token = token
			config.oauth_token_secret = secret
			config.Write()
		} else {
			fmt.Printf("authorized failed")
			os.Exit(1)
		}
	}

	info, _ := kp.AccountInfo()
	kp.Metadata("/", nil)
	kp.Share("ooo.txt", "oooo", "1234")
	kp.CreateFolder("1中医")
	kp.Delete("1中医", true)

	fmt.Printf("%v\n", info)
}
