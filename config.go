package main

import (
	"fmt"
	"log"

	"github.com/jinzhu/configor"
)

//Config contains global app's configuration
var Config AppConfig

// AppConfig contains app's configuration
type AppConfig struct {
	Server struct {
		Port string `default:":80"`
	}
	DataDB struct {
		User     string
		Host     string
		Password string
		Database string
	}
	AppDB struct {
		User     string
		Host     string
		Password string
		Database string
	}
}

//LoadFromFile method loads and parses config file
func (c *AppConfig) LoadFromFile(url string) {
	err := configor.Load(&Config, url)
	if err != nil {
		log.Fatalf("Can't load the config file: %s", err)
	}
}

func (c *AppConfig) DataDBSourceName() string {
	return fmt.Sprintf("%s:%s@(%s)/%s", c.DataDB.User, c.DataDB.Password, c.DataDB.Host, c.DataDB.Database)
}

func (c *AppConfig) AppDBSourceName() string {
	return fmt.Sprintf("%s:%s@(%s)/%s", c.AppDB.User, c.AppDB.Password, c.AppDB.Host, c.AppDB.Database)
}
