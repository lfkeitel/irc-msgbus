package config

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/lfkeitel/yobot/utils"
	"github.com/naoina/toml"
)

type Config struct {
	Main   MainConfig
	IRC    IRCConfig
	HTTP   HTTPConfig
	Routes map[string]RouteConfig
}

type MainConfig struct {
	Debug      bool
	ExtraDebug bool
	ModulesDir string
	Modules    []string
	DataDir    string
}

type IRCConfig struct {
	Server                string
	Port                  int
	Nick                  string
	TLS                   bool
	InsecureTLS           bool
	Channels              []string
	AutoJoinAlertChannels bool

	SASL struct {
		UseSASL  bool `toml:"sasl"`
		Login    string
		Password string
	}
}

type HTTPConfig struct {
	Address string
}

type RouteConfig struct {
	Enabled  bool
	Channels []string
	Username string
	Password string
	Alias    string
	Settings map[string]string
}

func LoadConfig(filename string) (conf *Config, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()

	if filename == "" {
		filename = "config.toml"
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var con Config
	if err := toml.Unmarshal(buf, &con); err != nil {
		return nil, err
	}
	return setSensibleDefaults(&con)
}

func setSensibleDefaults(con *Config) (*Config, error) {
	con.Main.ModulesDir = utils.FirstString(con.Main.ModulesDir, "modules")
	con.Main.DataDir = utils.FirstString(con.Main.DataDir, "data")
	return con, nil
}

// ModuleDataDir returns the path to a modules data directory.
func (c *Config) ModuleDataDir(name string) string {
	return filepath.Join(c.Main.DataDir, name)
}
