package remoteshell

import (
	"errors"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type (
	Servers  []server
	commands []command
	command  string
	server   struct {
		Name   string         `yaml:"name"`
		Host   string         `yaml:"host"`
		User   string         `yaml:"user"`
		Fatal  string         `yaml:"fatal"`
		Output sshOutputArray `yaml:"output"`
	}
	Config struct {
		Version  string              `yaml:"version"`
		Servers  map[string]Servers  `yaml:"servers"`
		Commands map[string]commands `yaml:"commands"`
	}
	ConfigError error
)

func (conf *Config) newError(errMsg string) ConfigError {
	return errors.New(errMsg)
}

func (conf *Config) set(ConfigPath ConfigPath) error {
	data, err := ioutil.ReadFile(string(ConfigPath))
	if err != nil {
		return conf.newError(fmt.Sprintf("Could not open %s  ", ConfigPath))
	}
	if err := yaml.Unmarshal([]byte(data), &conf); err != nil {
		return conf.newError(fmt.Sprintf("Could not parse Config file. Make sure its yaml."))
	}
	return nil
}

func (conf *Config) getServersFromConfig(ServerGroup ServerGroup) (Servers, ConfigError) {
	group, ok := conf.Servers[string(ServerGroup)]
	fmt.Println("Here...")
	fmt.Println(group, ok, ServerGroup)
	if !ok {
		return nil, conf.newError(fmt.Sprintf("Could not find [%s] in server group.", ServerGroup))
	}
	return group, nil
}

func (conf *Config) getCommandsFromConfig(CommandName CommandName) (commands, ConfigError) {
	commands, ok := conf.Commands[string(CommandName)]
	if !ok {
		return nil, conf.newError(fmt.Sprintf("Command %s was not found.", CommandName))
	}
	return commands, nil
}
