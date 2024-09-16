package config

import (
	"github.com/raphael-p/gocommon/validate"
)

type ServerConfig struct {
	Port uint `json:"port"`
}

type DBConfig struct {
	Username string                     `json:"username"`
	Password validate.JSONField[string] `json:"password" zeroable:"true"`
}

type BrendanConfig struct {
	Server   ServerConfig `json:"server"`
	Database DBConfig     `json:"database"`
}

var Values *BrendanConfig = &BrendanConfig{}
