package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

var GlobalConfig *Config

func init() {
	GlobalConfig = NewConfig()
	GlobalConfig.Reload()
}

type Config struct {
	PacketSendChanLimit    uint32 `json:"SendLimit"`
	PacketReceiveChanLimit uint32 `json:"ReceiveLimit"`
	Ip                     string `json:"ip"`
	Port                   uint16 `json:"port"`
}

func NewConfig() *Config {
	return &Config{
		PacketReceiveChanLimit: 128,
		PacketSendChanLimit:    128,
		Ip:                     "0.0.0.0",
		Port:                   8000,
	}
}

func (config *Config) Reload() {
	file, err := os.Open("./app.json")
	if err != nil {
		log.Println("app.json不存在或者无法打开")
		return
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(err)
		return
	}
	if err := json.Unmarshal(data, config); err != nil {
		log.Println("json解析失败")
		return
	}
}
