package plugin_manager

import (
	"fmt"
	"sync"

	"github.com/langgenius/dify-plugin-daemon/internal/cluster"
	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

type PluginManager struct {
	m sync.Map

	cluster *cluster.Cluster
}

var (
	manager *PluginManager
)

func InitGlobalPluginManager(cluster *cluster.Cluster, configuration *app.Config) {
	manager = &PluginManager{
		cluster: cluster,
	}
	manager.Init(configuration)
}

func GetGlobalPluginManager() *PluginManager {
	return manager
}

func (p *PluginManager) Add(plugin entities.PluginRuntimeInterface) error {
	identity, err := plugin.Identity()
	if err != nil {
		return err
	}
	p.m.Store(identity, plugin)
	return nil
}

func (p *PluginManager) List() []entities.PluginRuntimeInterface {
	var runtimes []entities.PluginRuntimeInterface
	p.m.Range(func(key, value interface{}) bool {
		if v, ok := value.(entities.PluginRuntimeInterface); ok {
			runtimes = append(runtimes, v)
		}
		return true
	})
	return runtimes
}

func (p *PluginManager) Get(identity string) entities.PluginRuntimeInterface {
	if v, ok := p.m.Load(identity); ok {
		if r, ok := v.(entities.PluginRuntimeInterface); ok {
			return r
		}
	}
	return nil
}

func (p *PluginManager) Init(configuration *app.Config) {
	// TODO: init plugin manager
	log.Info("start plugin manager daemon...")

	// init redis client
	if err := cache.InitRedisClient(
		fmt.Sprintf("%s:%d", configuration.RedisHost, configuration.RedisPort),
		configuration.RedisPass,
	); err != nil {
		log.Panic("init redis client failed: %s", err.Error())
	}

	if err := dify_invocation.InitDifyInvocationDaemon(
		configuration.PluginInnerApiURL, configuration.PluginInnerApiKey,
	); err != nil {
		log.Panic("init dify invocation daemon failed: %s", err.Error())
	}

	// start local watcher
	p.startLocalWatcher(configuration)

	// start remote watcher
	p.startRemoteWatcher(configuration)
}
