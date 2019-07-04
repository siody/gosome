package config

import (
	"github.com/spf13/viper"
)

// GetCluster get cluster name according to env
func GetCluster() string {
	d := detectDeploy()
	return d.getCluster()
}

// GetID get container ID according to env
func GetID() string {
	d := detectDeploy()
	return d.getContainerID()
}

// GetAPP get app name
func GetAPP() string {
	detectDeploy()
	return viper.GetString(APPNAMEKEY)
}

func devReadIn() error {
	viper.AddConfigPath("runtime")
	viper.AddConfigPath("conf")
	viper.AddConfigPath("config")
	viper.SetConfigName("dev")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	for _, h := range hooks {
		h()
	}
	return nil
}
