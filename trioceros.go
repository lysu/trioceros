package trioceros

import (
	"bytes"
	"github.com/spf13/viper"
	crypt "github.com/xordataexchange/crypt/config"
	"io"
	"os"
"strings"
)

// LocalCacheEnable enhance viper remote provider to support local file downgrading
// support restart application when etcd or consul has down..
var LocalCacheEnable bool = true

type triocerosConfigProvider struct{}

func (rc triocerosConfigProvider) Get(rp viper.RemoteProvider) (io.Reader, error) {
	cm, err := getConfigManager(rp)
	if err != nil {
		return nil, err
	}
	b, err := cm.Get(rp.Path())
	if err != nil {
		return nil, err
	}

	if LocalCacheEnable {

	}

	return bytes.NewReader(b), nil
}

func (rc triocerosConfigProvider) Watch(rp viper.RemoteProvider) (io.Reader, error) {
	cm, err := getConfigManager(rp)
	if err != nil {
		return nil, err
	}
	resp := <-cm.Watch(rp.Path(), nil)
	err = resp.Error
	if err != nil {
		return nil, err
	}
	backupRemoteConfig()
	return bytes.NewReader(resp.Value), nil
}

func backupRemoteConfig() {
	if !LocalCacheEnable {
		return
	}
}

func getConfigManager(rp viper.RemoteProvider) (crypt.ConfigManager, error) {

	var cm crypt.ConfigManager
	var err error

	if rp.SecretKeyring() != "" {
		kr, err := os.Open(rp.SecretKeyring())
		defer kr.Close()
		if err != nil {
			return nil, err
		}
		if rp.Provider() == "etcd" {
			cm, err = crypt.NewEtcdConfigManager(toMachines(rp.Endpoint()), kr)
		} else {
			cm, err = crypt.NewConsulConfigManager(toMachines(rp.Endpoint()), kr)
		}
	} else {
		if rp.Provider() == "etcd" {
			cm, err = crypt.NewStandardEtcdConfigManager(toMachines(rp.Endpoint()))
		} else {
			cm, err = crypt.NewStandardConsulConfigManager(toMachines(rp.Endpoint()))
		}
	}
	if err != nil {
		return nil, err
	}
	return cm, nil

}

func toMachines(endpoint string) []string {
	machines := strings.Split(endpoint, ",")
	return machines
}

func init() {
	viper.RemoteConfig = &triocerosConfigProvider{}
}
