package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/bamzi/jobrunner"
	"github.com/labstack/echo"
	"github.com/yasukun/price-cache/lib"
)

var configPath = flag.String("c", "price-cache.toml", "config path")
var serverPort = flag.String("p", ":1323", "server port")
var pluginKey = flag.String("k", "", "plugin key")

// validate ...
func validate() error {
	flag.Parse()
	if *configPath != "" {
		_, err := os.Stat(*configPath)
		if err != nil {
			return errors.New(fmt.Sprintf("config %s does not exists", *configPath))
		}
	}
	if *pluginKey == "" {
		return errors.New("plugin key requierd")
	}
	return nil
}

func main() {
	var err error
	if err = validate(); err != nil {
		log.Fatalln(err)
	}

	configAbs := *configPath
	if !filepath.IsAbs(*configPath) {
		configAbs, err = filepath.Abs(*configPath)
		if err != nil {
			log.Fatalln(err)
		}
	}

	conf, err := lib.DecodeConfigToml(configAbs)
	if err != nil {
		log.Fatalln(err)
	}

	runner, err := lib.NewRunner(configAbs, *pluginKey, conf)
	if err != nil {
		log.Fatalln(err)
	}

	jobrunner.Start()
	jobrunner.Schedule(conf.Main.Schedule, runner)

	e := echo.New()

	go func() {
		if err := e.Start(*serverPort); err != nil {
			e.Logger.Info("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
