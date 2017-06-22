package remoteshell

import (
	"errors"
	"flag"
)

type (
	AsyncFlag   bool
	ConfigPath  string
	KeyPath     string
	ServerGroup string
	CommandName string
	VerifyFlag  bool

	Input struct {
		AsyncFlag   AsyncFlag
		ConfigPath  ConfigPath
		KeyPath     KeyPath
		ServerGroup ServerGroup
		CommandName CommandName
		VerifyFlag  VerifyFlag
	}
	InputError error
)

func GetInputData() *Input {

	verifyActionFlagPtr := flag.Bool("y", false, "force yes")
	asyncFlagPtr := flag.Bool("async", false, "async - if true, parallel executing over servers")
	ConfigPathPtr := flag.String("c", "", "config file - yaml config file")
	KeyPathPtr := flag.String("i", "", "identity file - path to private key")
	ServerGroupPtr := flag.String("s", "", "server group - name of the server group")
	commandPtr := flag.String("cmd", "", "command name - name of the command to run")

	flag.Parse()

	return &Input{
		VerifyFlag:  VerifyFlag(*verifyActionFlagPtr),
		AsyncFlag:   AsyncFlag(*asyncFlagPtr),
		CommandName: CommandName(*commandPtr),
		ServerGroup: ServerGroup(*ServerGroupPtr),
		KeyPath:     KeyPath(*KeyPathPtr),
		ConfigPath:  ConfigPath(*ConfigPathPtr),
	}
}

func (Input *Input) newError(errMsg string) InputError {
	return errors.New(errMsg)
}

func (Input *Input) validate() InputError {
	if err := Input.ConfigPath.validate(Input); err != nil {
		return err
	}
	if err := Input.KeyPath.validate(Input); err != nil {
		return err
	}
	if err := Input.ServerGroup.validate(Input); err != nil {
		return err
	}
	if err := Input.CommandName.validate(Input); err != nil {
		return err
	}
	return nil
}

func (val *ConfigPath) validate(Input *Input) InputError {
	if *val == "" {
		return Input.newError("ConfigPath is empty please set grapes -c config.yml")
	}
	return nil
}

func (val *KeyPath) validate(Input *Input) InputError {
	if *val == "" {
		return Input.newError("identity file path is empty please set grapes -i ~/.ssh/id_rsa")
	}
	return nil
}

func (val *ServerGroup) validate(Input *Input) InputError {
	if *val == "" {
		return Input.newError("server group is empty please set grapes -s server_group")
	}
	return nil
}

func (val *CommandName) validate(Input *Input) InputError {
	if *val == "" {
		return Input.newError("command name is empty please set grapes -cmd whats_up")
	}
	return nil
}
