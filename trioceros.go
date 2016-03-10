package trioceros

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	crypt "github.com/xordataexchange/crypt/config"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	backupTimeFormat = "2006-01-02T15-04-05.000"
)

// LocalCacheEnable enhance viper remote provider to support local file downgrading
// support restart application when etcd or consul has down..
var LocalCacheEnable = true

var LocalConfigFile = "/Users/xx/config_backup"

type triocerosConfigProvider struct{}

func (rc triocerosConfigProvider) Get(rp viper.RemoteProvider) (io.Reader, error) {
	cm, err := getConfigManager(rp)
	if err != nil {
		return nil, err
	}
	b, err := cm.Get(rp.Path())
	if err != nil {
		backupReader, err2 := restoreSnapshotConfig()
		if err2 == nil {
			fmt.Println("remote failure and use local config")
			return backupReader, nil
		}
		return nil, err
	}
	if LocalCacheEnable {
		err = backupSnapshotConfig(b)
		if err != nil {
			return nil, err
		}
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
	err = backupSnapshotConfig(resp.Value)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(resp.Value), nil
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

func restoreSnapshotConfig() (io.Reader, error) {
	if !LocalCacheEnable {
		return nil, fmt.Errorf("remote configserver failured but localCacheDisabled.")
	}
	_, err := os.Stat(LocalConfigFile)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("can't found backup configfile")
	}
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(LocalConfigFile, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func backupSnapshotConfig(confBin []byte) error {
	if !LocalCacheEnable {
		return nil
	}
	err := os.MkdirAll(filepath.Dir(LocalConfigFile), 0744)
	if err != nil {
		return fmt.Errorf("can't make directories for new configfiles %s", err)
	}

	name := LocalConfigFile
	mode := os.FileMode(0644)
	info, err := os.Stat(name)
	if err == nil {
		mode = info.Mode()
		newName := backupName(name)
		if err := os.Rename(name, newName); err != nil {
			return fmt.Errorf("can't rename configfiles: %s", err)
		}
		if err := chown(name, info); err != nil {
			return err
		}
	}

	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("can't open new configfiles: %s", err)
	}
	defer f.Close()

	_, err = f.Write(confBin)
	return err
}

func backupName(name string) string {
	dir := filepath.Dir(name)
	filename := filepath.Base(name)
	ext := filepath.Ext(filename)
	prefix := filename[:len(filename)-len(ext)]
	t := time.Now()
	timestamp := t.Format(backupTimeFormat)
	return filepath.Join(dir, fmt.Sprintf("%s-%s%s", prefix, timestamp, ext))
}

func init() {
	viper.RemoteConfig = &triocerosConfigProvider{}
}
