package config

type envarType struct {
	ConfigFilepath string
}

var Envars envarType = envarType{
	ConfigFilepath: "BRENDAN_CONFIG_FILEPATH",
}
