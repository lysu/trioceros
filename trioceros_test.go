package trioceros_test

import (
	"testing"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"fmt"
	"time"
)

func TestViperEtcd(t *testing.T) {

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
		time.Sleep(2*time.Second)
		fmt.Println("value: ", viper.GetInt("a"))
	}

}