package config

type envarType struct {
	configFilepath string
}

var envars envarType = envarType{
	configFilepath: "BRENDAN_CONFIG_FILEPATH",
}
