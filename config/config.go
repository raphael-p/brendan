package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/raphael-p/brendan/utils"
)

type ServerConfig struct {
	Port uint `json:"port"`
}

type DBConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type BrendanConfig struct {
	Server   ServerConfig `json:"server"`
	Database DBConfig     `json:"database"`
}

var Config *BrendanConfig

func InitialiseConfig() {
	filePath := os.Getenv(envars.configFilepath)
	if filePath == "" {
		fmt.Printf("$%s not set, using default config\n", envars.configFilepath)
		filePath = filepath.Join(utils.GetExecDirectory("."), "default.json")
	}

	file, err := os.Open(filePath)
	if err != nil {
		panic(fmt.Sprint("could not open config file: ", err))
	}
	defer file.Close()

	Config = &BrendanConfig{}
	if err = json.NewDecoder(file).Decode(Config); err != nil {
		panic(fmt.Sprint("could not parse config file: ", err))
	}
}
