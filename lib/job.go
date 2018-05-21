package lib

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"plugin"

	"github.com/go-redis/redis"
)

const PLUGIN_INIT_METHOD = "LoadConf"
const PLUGIN_RUN_METHOD = "TakePrice"

type Runner struct {
	PluginPath   string
	PluginConfig string
	ListKey      string
	Client       *redis.Client
}

// NewRunner ...
func NewRunner(configAbs, pluginKey string, conf Config) (Runner, error) {
	var runner Runner
	var pluginPathAbs string
	var listKey string
	var err error
	if v, ok := conf.Plugin[pluginKey]; ok {
		pluginPathAbs, err = filepath.Abs(v.Path)
		if err != nil {
			return runner, err
		}
		listKey = v.ListKey
	} else {
		return runner, errors.New(fmt.Sprintf("plugin key:%s not exists", pluginKey))
	}
	client := redis.NewClient(&redis.Options{
		Addr:     conf.Ledisdb.Addr,
		Password: conf.Ledisdb.Password,
		DB:       conf.Ledisdb.DB,
	})
	runner = Runner{
		PluginPath:   pluginPathAbs,
		PluginConfig: configAbs,
		ListKey:      listKey,
		Client:       client,
	}
	return runner, nil
}

// (r Runner) execPlugin ...
func (r Runner) execPlugin() (string, error) {
	var jsonStr string
	if _, err := os.Stat(r.PluginPath); err != nil {
		return jsonStr, errors.New(fmt.Sprintf("plugin %s does not exists", r.PluginPath))
	}

	p, err := plugin.Open(r.PluginPath)
	if err != nil {
		return jsonStr, err
	}
	f, err := p.Lookup(PLUGIN_INIT_METHOD)
	if err != nil {
		return jsonStr, errors.New(fmt.Sprintf("plugin.Lookup error name: %s, method: %s, error: %v", r.PluginPath, PLUGIN_INIT_METHOD, err))
	}
	if err = f.(func(string) error)(r.PluginConfig); err != nil {
		return jsonStr, errors.New(fmt.Sprintf("method execute error name: %s, method: %s, error: %v", r.PluginPath, PLUGIN_INIT_METHOD, err))
	}
	f2, err := p.Lookup(PLUGIN_RUN_METHOD)
	if err != nil {
		return jsonStr, errors.New(fmt.Sprintf("plugin.Lookup error name: %s, method: %s, error: %v", r.PluginPath, PLUGIN_RUN_METHOD, err))
	}
	jsonStr, err = f2.(func() (string, error))()
	if err != nil {
		return jsonStr, errors.New(fmt.Sprintf("method execute error name: %s, method: %s, error: %v", r.PluginPath, PLUGIN_RUN_METHOD, err))
	}
	return jsonStr, nil
}

// (r Runner) Run ...
func (r Runner) Run() {
	jsonStr, err := r.execPlugin()
	if err != nil {
		log.Println(err)
		return
	}
	if err = r.Client.RPush(r.ListKey, jsonStr).Err(); err != nil {
		log.Println(fmt.Sprintf("RPush error: %v", err))
		return
	}
	return
}
