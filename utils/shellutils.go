package utils

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"bitbucket.org/myntra/nazgul-shadowfax/lib"
	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/elgs/gostrgen"
	"github.com/valyala/fasthttp"
	"golang.org/x/crypto/ssh"
)


//Create SSH connection with server name
func CreateConnection(server, username, pemFilePath string) *ssh.Client {
	pemBytes, err := ioutil.ReadFile(pemFilePath)
	if err != nil {
		log.Fatal(err)
	}
	signer, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		log.Fatalf("parse key failed:%v", err)
	}
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
	}
	conn, err := ssh.Dial("tcp", server+":22", config)
	if err != nil {
		log.Fatalf("dial failed 1 :%v", err)
		conn, err = ssh.Dial("tcp", server+":22", config)
		if err != nil {
			log.Fatalf("dial failed 2 :%v", err)
		}
	}
	return conn
}

//Get the session from connection
func GetSession(conn *ssh.Client) *ssh.Session {
	session, err := conn.NewSession()
	if err != nil {
		log.Fatalf("session failed:%v", err)
	}
	return session
}

//Execute command on the created session
func ExecuteCommand(conn *ssh.Client, command string) bool {
	session := GetSession(conn)
	fmt.Println("Executing ==> ", command)
	err := session.Start(command)
	session.Wait()
	if err != nil {
		fmt.Println("Run failed ExecuteCommand :", err)
		return false
	}
	return true
}

//To execute commands on local
func ExecuteCommandOnLocal(command string) bool {
	cmd := exec.Command(command)
	err := cmd.Start()
	cmd.Wait()
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

//Return process id runing on a specific port on a remote machine
func GetProcessId(serverIP string, port int) (runningPort string, err error) {
	command := "/usr/sbin/lsof -i tcp:" + strconv.Itoa(port) + " | tail -1 | awk '{ print $2; }'"
	fmt.Println("Info : Command to get process id : ", command)
	connection := CreateConnection(serverIP, "sysadmin", "sysadmin.pem")
	session := GetSession(connection)
	defer session.Close()
	output, err := session.Output(command)
	if err != nil {
		fmt.Println("Run failed GetProcessId:", err.Error())
	}
	runningPort = string(output)
	fmt.Println("Port Number : ", runningPort)
	return
}

//Generate random string
func GenerateRandomString(length int) string {
	charSet := gostrgen.Lower | gostrgen.Digit
	includes := ""   // optionally include some additional letters
	excludes := "Ol" //exclude big 'O' and small 'l' to avoid confusion with zero and one.
	str, err := gostrgen.RandGen(length, charSet, includes, excludes)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(str)
	return str
}

//Get free port on localhost
func GetPort() int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		log.Println(err)
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Println(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

//Check if file exists?
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("ERROR : Lookup Failed ", name)
			return false
		}
	}
	return true
}

// SCP client to copy data from local to remote machines
func CreateSCPClient(address, username, pemFilePath string) scp.Client {
	pemBytes, err := ioutil.ReadFile(pemFilePath)
	if err != nil {
		log.Fatal(err)
	}
	signer, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		log.Fatalf("parse key failed:%v", err)
	}
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
	}
	// Create a new SCP client
	return scp.NewClient(address+":22", config)
}

//Set Error Response
func SetErrorResponse(ctx *fasthttp.RequestCtx, slaveCount int) {
	var response Response
	ctx.Response.Header.Add("Content-Type", "application/json")	
}

//Set Success Response
func SetSuccessResponse(ctx *fasthttp.RequestCtx, masterIp, masterRunId, outputLocation, jtl, console string, slaveCount int) {
	var response Response
	ctx.Response.Header.Add("Content-Type", "application/json")	
}