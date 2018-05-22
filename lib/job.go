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

const PLUGIN_RUN_METHOD = "TakePrice"

type Runner struct {
	Plugin  *plugin.Plugin
	ListKey string
	Client  *redis.Client
	Debug   bool
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
	p, err := loadPlugin(pluginPathAbs)
	if err != nil {
		return runner, errors.New(fmt.Sprintf("load plugin error: %v", err))
	}
	runner = Runner{
		Plugin:  p,
		ListKey: listKey,
		Client:  client,
		Debug:   conf.Main.Debug,
	}
	return runner, nil
}

// loadPlugin ...
func loadPlugin(pluginPath string) (*plugin.Plugin, error) {
	var p *plugin.Plugin
	if _, err := os.Stat(pluginPath); err != nil {
		return p, errors.New(fmt.Sprintf("plugin %s does not exists", pluginPath))
	}

	p, err := plugin.Open(pluginPath)
	if err != nil {
		return p, err
	}
	return p, nil
}

// (r Runner) execPlugin ...
func (r Runner) execPlugin() (string, error) {
	var jsonStr string

	f, err := r.Plugin.Lookup(PLUGIN_RUN_METHOD)
	if err != nil {
		return jsonStr, errors.New(fmt.Sprintf("plugin.Lookup error %v, method %s", err, PLUGIN_RUN_METHOD))
	}
	jsonStr, err = f.(func() (string, error))()
	if err != nil {
		return jsonStr, errors.New(fmt.Sprintf("method execute error %v, method %s", err, PLUGIN_RUN_METHOD))
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
	if r.Debug {
		log.Println(jsonStr)
	}
	if err = r.Client.RPush(r.ListKey, jsonStr).Err(); err != nil {
		log.Println(fmt.Sprintf("RPush error: %v", err))
		return
	}
	return
}
