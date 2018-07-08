package grpchelper

import (
	"os"

	"github.com/spf13/viper"
)

// NewConfig create a new config, set host to empty if use local file
func NewConfig(host string, conf string) (*viper.Viper, error) {
	config := viper.New()
	if host != "" {
		config.AddRemoteProvider("consul", host, conf)
		config.SetConfigType("json")
		if err := config.ReadRemoteConfig(); err != nil {
			return nil, err
		}
	} else {
		fp, err := os.Open(conf)
		if err != nil {
			return nil, err
		}
		if err := config.ReadConfig(fp); err != nil {
			return nil, err
		}
	}

	return config, nil
}
