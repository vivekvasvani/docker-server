package remoteshell

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"sync"

	"golang.org/x/crypto/ssh"
)

type (
	grapeSSH struct {
		keySigner ssh.Signer
	}
	grapeSSHClient struct {
		*ssh.Client
	}
	std struct {
		Out string
		Err string
	}
	sshOutput struct {
		Command command
		Std     std
	}
	sshOutputArray []*sshOutput
	sshError       error
)

func (gSSH *grapeSSH) newError(errMsg string) sshError {
	return errors.New(errMsg)
}

func (gSSH *grapeSSH) setKey(KeyPath KeyPath) sshError {
	privateBytes, err := ioutil.ReadFile(string(KeyPath))
	if err != nil {
		return gSSH.newError("Could not open idendity file.")
	}
	privateKey, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return gSSH.newError(fmt.Sprint("Could not parse identity file."))
	}
	gSSH.keySigner = privateKey
	return nil

}

func (gSSH *grapeSSH) newClient(server server) (*grapeSSHClient, sshError) {
	client, err := ssh.Dial("tcp", server.Host, &ssh.ClientConfig{
		User: server.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(gSSH.keySigner),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	})
	if err != nil {
		return nil, gSSH.newError("Could not establish ssh connection")
	}
	return &grapeSSHClient{client}, nil
}

func (client *grapeSSHClient) execCommand(cmd command, wg *sync.WaitGroup) *sshOutput {
	output := &sshOutput{
		Command: cmd,
	}
	session, err := client.NewSession()
	if err != nil {
		output.Std.Err = "Could not establish ssh session"
	} else {
		var stderr, stdout bytes.Buffer
		session.Stdout, session.Stderr = &stdout, &stderr
		session.Run(string(cmd))
		session.Close()
		output.Std = std{
			Out: stdout.String(),
			Err: stderr.String(),
		}
	}
	if wg != nil {
		wg.Done()
	}
	return output
}

func (client *grapeSSHClient) execCommands(commands commands, app grape) sshOutputArray {
	output := sshOutputArray{}
	for _, command := range commands {
		if app.Input.AsyncFlag {
			wg.Add(1)
			go client.execCommand(command, &wg)
		} else {
			output = append(output, client.execCommand(command, nil))
		}
	}
	if app.Input.AsyncFlag {
		wg.Wait()
	}
	return output
}
