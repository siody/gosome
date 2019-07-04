package config

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/etcdserver/api/v3rpc/rpctypes"
)

// RemoteReadIn read config from etcd
func RemoteReadIn(etcd *clientv3.Client, prefix string) error {
	path := prefix + "/" + GetAPP() + "/" + GetCluster() + ".yaml"
	fmt.Printf("read from path: %s in etcd.\n", path)
	config, err := etcd.Get(context.Background(), path)
	if err != nil {
		switch err {
		case context.Canceled:
			fmt.Printf("ctx is canceled by another routine: %v\n", err)
		case context.DeadlineExceeded:
			fmt.Printf("ctx is attached with a deadline is exceeded: %v\n", err)
		case rpctypes.ErrEmptyKey:
			fmt.Printf("client-side error: %v\n", err)
		default:
			fmt.Printf("bad cluster endpoints, which are not etcd servers: %v\n", err)
		}
		return err
	}
	if len(config.Kvs) == 0 {
		return errors.New("can`t find config in etcd")
	}
	fmt.Printf("get config file\n%s\n", string(config.Kvs[0].Value))
	viper.SetConfigType("yaml")
	viper.ReadConfig(bytes.NewBuffer(config.Kvs[0].Value))

	for _, h := range hooks {
		h()
	}

	return nil
}
