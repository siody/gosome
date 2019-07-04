package config

import (
	"errors"
	"fmt"
	"github.com/rs/xid"
	"github.com/spf13/viper"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"
)

var (
	//APPNAMEKEY appname key in viper
	APPNAMEKEY = "appname"

	deployMu      sync.Mutex
	deployAtomics atomic.Value
	// DeployName mpaas or k8s type is deploy
	DeployName unsafe.Pointer
)

type deploy struct {
	name, magic    string
	readin         func() error
	getCluster     func() string
	getContainerID func() string
}

//K8SRead get config from env when is deploy in k8s
func K8SRead() error {
	config, exists := os.LookupEnv(viper.GetString(APPNAMEKEY))
	if !exists && config != "" {
		return errors.New("config env is empty or not exists")
	}
	viper.SetConfigType("yaml")
	viper.ReadConfig(strings.NewReader(config))
	return nil
}

// K8SGetCluster get k8s namespace
func K8SGetCluster() string {
	if cluster, ok := os.LookupEnv("ENV_CLUSTER"); ok && cluster != "" {
		return cluster
	}
	return "beta"
}

// K8SContainerID get unique id
func K8SContainerID() string {
	if cluster, ok := os.LookupEnv("MESOS_CONTAINER_NAME"); ok && cluster != "" {
		return cluster
	}
	return ""
}

//MpaasRead get config when is deploy in mpaas
func MpaasRead() error {
	viper.AddConfigPath("runtime")
	viper.AddConfigPath("/letv/app/runtime")
	cluster, ok := os.LookupEnv("ENV_CLUSTER")
	if ok && cluster != "" {
		viper.SetConfigName(viper.GetString(APPNAMEKEY) + "-" + cluster)
		fmt.Printf("mpaas config file on %s cluster\n", cluster)
	} else {
		viper.SetConfigName(viper.GetString(APPNAMEKEY))
		fmt.Printf("mpaas config file on raw\n")
	}
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	return nil
}

// MpaasGetCluster get mpaas cluster
func MpaasGetCluster() string {
	if cluster, ok := os.LookupEnv("ENV_CLUSTER"); ok && cluster != "" {
		return cluster
	}
	return "beta"
}

// MpaasContainerID get container unique id
func MpaasContainerID() string {
	if cluster, ok := os.LookupEnv("MESOS_CONTAINER_NAME"); ok && cluster != "" {
		return cluster
	}
	return ""
}

var hooks []func()

//AddHook hook will be execute when readin success
func AddHook(h func()) {
	hooks = append(hooks, h)
}

//ReadIn get config in letv
//1. detect deploy way
//2. use the conposed method
func ReadIn() error {
	dep := detectDeploy()
	if dep == nil {
		return errors.New("deploy is unsupport")
	}
	fmt.Println("Found deployed on " + dep.name + " and begin read in config")
	err := dep.readin()
	if err != nil {
		return err
	}
	for _, h := range hooks {
		h()
	}
	return nil
}

func detectDeploy() *deploy {
	if d := (*deploy)(DeployName); d != nil {
		return d
	}
	deploys, _ := deployAtomics.Load().([]deploy)
	for _, d := range deploys {
		if appname, exists := os.LookupEnv(d.magic); appname != "" && exists {
			viper.Set(APPNAMEKEY, appname)
			viper.Set("deploy", d.name)
			atomic.StorePointer(&DeployName, unsafe.Pointer(&d))
			return &d
		}
	}

	devReadIn()
	viper.Set("deploy", "dev")
	dev := &deploy{
		name:   viper.GetString("appname"),
		magic:  "",
		readin: devReadIn,
		getCluster: func() string {
			return viper.GetString("cluster")
		},
		getContainerID: func() string {
			return viper.GetString("app") + xid.New().String()
		},
	}
	atomic.StorePointer(&DeployName, unsafe.Pointer(dev))
	return dev
}

//RegisterDeploy registers an deploy for read in config
func RegisterDeploy(d deploy) {
	deployMu.Lock()
	deploys, _ := deployAtomics.Load().([]deploy)
	deployAtomics.Store(append(deploys, d))
	deployMu.Unlock()
}

const (
	//MPAAS name of mpaas
	MPAAS = "mpaas"
	//K8S name of k8s
	K8S = "k8s"
)

func init() {
	RegisterDeploy(deploy{
		name:           MPAAS,
		magic:          "ENV_APP",
		readin:         MpaasRead,
		getCluster:     MpaasGetCluster,
		getContainerID: MpaasContainerID,
	})
	RegisterDeploy(deploy{
		name:           K8S,
		magic:          "APP_NAME",
		readin:         K8SRead,
		getCluster:     K8SGetCluster,
		getContainerID: K8SContainerID,
	})
}
