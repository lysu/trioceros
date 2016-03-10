package main

import (
	"fmt"
	_ "github.com/lysu/trioceros"
	"github.com/spf13/viper"
	"time"
)

func main() {

	viper.AddRemoteProvider("etcd", "http://127.0.0.1:4001", "/config/test.toml")
	viper.SetConfigType("toml")

	err := viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			err = viper.WatchRemoteConfig()
			if err != nil {
				fmt.Println("although all config storage has down, continue to retry watch.")
				time.Sleep(4 * time.Second)
			}
		}
	}()

	for {
		time.Sleep(2 * time.Second)
		fmt.Println("value: ", viper.GetInt("a"))
	}

}
