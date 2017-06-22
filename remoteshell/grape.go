package remoteshell

import (
	"fmt"
	"sync"
)

type (
	grape struct {
		Input    Input
		ssh      grapeSSH
		Config   Config
		Servers  Servers
		commands commands
	}
)

var wg sync.WaitGroup

func NewGrape(Input *Input) *grape {
	app := grape{}
	var err error

	app.Input = *Input

	//validate Input
	if err = app.Input.validate(); err != nil {
		panic(err)
	}
	//set Config into place
	if err = app.Config.set(app.Input.ConfigPath); err != nil {
		panic(err)
	}
	// data !
	if app.Servers, err = app.Config.getServersFromConfig(app.Input.ServerGroup); err != nil {
		panic(err)
	}
	if app.commands, err = app.Config.getCommandsFromConfig(app.Input.CommandName); err != nil {
		panic(err)
	}
	//load private key
	if err = app.ssh.setKey(app.Input.KeyPath); err != nil {
		panic(err)
	}
	return &app
}

func (app *grape) Run() {
	for _, server := range app.Servers {
		app.runOnServer(server)
	}
}

func (app *grape) runOnServer(server server) {
	client, err := app.ssh.newClient(server)
	if err != nil {
		server.Fatal = err.Error()
	} else {
		client.execCommands(app.commands, *app)
	}
	//server.printOutput()
}

/*
func (s *server) printOutput() {
	out, _ := yaml.Marshal(s)
	fmt.Println(string(out))
}
*/

func (app *grape) VerifyAction() {
	var char = "n"
	fmt.Println("The following command will run on the following Servers:")
	fmt.Printf("command `%s` will run over `%s`.\n", app.Input.CommandName, app.Input.ServerGroup)
	fmt.Println("commands:")
	for k, v := range app.commands {
		fmt.Printf("\t#%d - `%s` \n", k, v)
	}
	fmt.Println("Servers:")
	for k, v := range app.Servers {
		fmt.Printf("\t#%d - %s [%s@%s] \n", k, v.Name, v.User, v.Host)
	}
	if app.Input.VerifyFlag {
		fmt.Println("-y used.forced to continue.")
		return
	}
	fmt.Print("\n -- are your sure? [y/N] : ")
	if _, err := fmt.Scanf("%s", &char); err != nil || char != "y" {
		panic("type y to continue")
	}
}
