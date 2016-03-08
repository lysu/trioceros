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
		err = viper.WatchRemoteConfig()
		if err != nil {
			panic(err)
		}
	}()

	for {
		time.Sleep(2 * time.Second)
		fmt.Println("value: ", viper.GetInt("a"))
	}

}
