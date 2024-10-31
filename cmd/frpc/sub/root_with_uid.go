// uid_control.go

package sub

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fatedier/frp/client"
	"github.com/fatedier/frp/pkg/config"
	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/fatedier/frp/pkg/config/v1/validation"
	"github.com/fatedier/frp/pkg/util/log"
)

// UID control variables
var (
	services      = make(map[string]*client.Service)
	servicesMutex sync.RWMutex // To ensure thread safety
)

type RunMultipleClientConfig struct {
	Uid  string
	Path string
}

// GetUids returns a list of all active UIDs.
func GetUids() []string {
	servicesMutex.RLock()
	defer servicesMutex.RUnlock()

	uids := make([]string, 0, len(services))
	for uid := range services {
		uids = append(uids, uid)
	}
	return uids
}

// IsRunning checks if a service with the given UID is running.
func IsRunning(uid string) bool {
	servicesMutex.RLock()
	defer servicesMutex.RUnlock()

	return services[uid] != nil
}

func getServiceByUid(uid string) *client.Service {
	servicesMutex.RLock()
	defer servicesMutex.RUnlock()

	return services[uid]
}

func putServiceByUid(uid string, svr *client.Service) {
	servicesMutex.Lock()
	defer servicesMutex.Unlock()

	services[uid] = svr
}

func delServiceByUid(uid string) {
	servicesMutex.Lock()
	defer servicesMutex.Unlock()

	delete(services, uid)
}

// Close gracefully closes the service associated with the UID.
func Close(uid string) bool {
	svr := getServiceByUid(uid)
	if svr != nil {
		svr.GracefulClose(500 * time.Millisecond)
		delServiceByUid(uid)
		return true
	}
	return false
}

// RunClientWithUid runs the client with a given UID and config file path.
func RunClientWithUid(uid string, cfgFilePath string, strictConfigMode bool) error {
	cfg, proxyCfgs, visitorCfgs, isLegacyFormat, err := config.LoadClientConfig(cfgFilePath, strictConfigMode)
	if err != nil {
		return err
	}
	if isLegacyFormat {
		fmt.Printf("WARNING: ini format is deprecated and will be removed in the future. " +
			"Please use yaml/json/toml format instead!\n")
	}

	warning, err := validation.ValidateAllClientConfig(cfg, proxyCfgs, visitorCfgs)
	if warning != nil {
		fmt.Printf("WARNING: %v\n", warning)
	}
	if err != nil {
		return err
	}
	return startServiceWithUid(uid, cfg, proxyCfgs, visitorCfgs, cfgFilePath)
}

// RunMultipleClientsWithUid 运行多个带有 UID 控制的客户端服务
// 接收包含 UID 和路径的结构体列表
func RunMultipleClientsWithUid(configs []RunMultipleClientConfig, strictConfigMode bool) error {
	var wg sync.WaitGroup
	for _, config := range configs {
		wg.Add(1)
		go func(cfg RunMultipleClientConfig) {
			defer wg.Done()
			err := RunClientWithUid(cfg.Uid, cfg.Path, strictConfigMode)
			if err != nil {
				fmt.Printf("frpc service error for UID [%s] with config file [%s]: %v\n", cfg.Uid, cfg.Path, err)
			}
		}(config)
		// 防止系统过载，添加一个小的延迟
		time.Sleep(time.Millisecond)
	}
	wg.Wait()
	return nil
}

// startServiceWithUid starts the service with UID control.
func startServiceWithUid(
	uid string,
	cfg *v1.ClientCommonConfig,
	proxyCfgs []v1.ProxyConfigurer,
	visitorCfgs []v1.VisitorConfigurer,
	cfgFile string,
) error {
	defer delServiceByUid(uid)

	log.InitLogger(cfg.Log.To, cfg.Log.Level, int(cfg.Log.MaxDays), cfg.Log.DisablePrintColor)

	if cfgFile != "" {
		log.Infof("start frpc service with UID [%s] for config file [%s]", uid, cfgFile)
		defer log.Infof("frpc service with UID [%s] for config file [%s] stopped", uid, cfgFile)
	} else {
		log.Infof("start frpc service with UID [%s]", uid)
		defer log.Infof("frpc service with UID [%s] stopped", uid)
	}

	svr, err := client.NewService(client.ServiceOptions{
		Common:         cfg,
		ProxyCfgs:      proxyCfgs,
		VisitorCfgs:    visitorCfgs,
		ConfigFilePath: cfgFile,
	})
	if err != nil {
		return err
	}

	putServiceByUid(uid, svr)

	shouldGracefulClose := cfg.Transport.Protocol == "kcp" || cfg.Transport.Protocol == "quic"
	if shouldGracefulClose {
		go handleTermSignal(svr)
	}
	err = svr.Run(context.Background())
	if err != nil {
		delServiceByUid(uid)
	}
	return err
}
