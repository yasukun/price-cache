package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/BurntSushi/toml"
)

type Config struct {
	OneForge OneForgeConfig `toml:"oneforge"`
}

type OneForgeConfig struct {
	URL     string          `toml:"url"`
	Pairs   string          `toml:"pairs"`
	ApiKey  string          `toml:"api_key"`
	Headers []HeadersConfig `toml:"headers"`
}

type HeadersConfig struct {
	Name  string `toml:"name"`
	Value string `toml:"value"`
}

var config Config

// init ...
func init() {
	var c Config
	if _, err := toml.DecodeFile("oneforge.toml", &c); err != nil {
		log.Fatalln(err)
	}
	config = c
}

// targetApi ...
func targetApi() string {
	return fmt.Sprintf(config.OneForge.URL, config.OneForge.Pairs, config.OneForge.ApiKey)
}

// TakePrice ...
func TakePrice() (string, error) {
	var result string
	url := targetApi()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return result, err
	}
	for _, header := range config.OneForge.Headers {
		req.Header.Set(header.Name, header.Value)
	}
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	byteResult, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}
	return string(byteResult), nil
}

func main() {
	result, err := TakePrice()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(result)
}
