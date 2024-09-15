package config

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

var Config *BrendanConfig = &BrendanConfig{}
