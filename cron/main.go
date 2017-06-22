package main

import (
	"database/sql"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/carlescere/scheduler"
	_ "github.com/go-sql-driver/mysql"
	"github.com/nlopes/slack"
	shell "github.com/vivekvasvani/docker-server/remoteshell"
	server "github.com/vivekvasvani/docker-server/server"
)

var (
	db *sql.DB
	wg sync.WaitGroup
)

const SLACK_TOKEN = "xoxp-2151902985-167537642311-189252965986-bbef2b7b92a16731b291eccf5cbb1384"

func main() {
	runtime.GOMAXPROCS(3)
	getDB()
	var wg sync.WaitGroup
	// Run every 2 seconds but not now.
	job := func() {
		output := FetchOneCandidate(db)
		//id, env, instance_id, clear_previous_run, public_ip, private_ip, action := FetchOneCandidate(db)
		//id, env, instance_id, clear_previous_run, public_ip, private_ip, action := FetchOneCandidate(db)
		fmt.Println(output)
		id, _ := strconv.Atoi(output["id"])
		env := output["env"]
		if id > 0 && env != "" {
			wg.Add(1)
		} else {
			return
		}
		fmt.Println("------------------------------------------------------------------------------------------------------------------")
		fmt.Println("Executing Job ID : -------> ", id, "    Current Time : --------> ", time.Now())
		fmt.Println("------------------------------------------------------------------------------------------------------------------")

		switch output["action"] {
		case "start":
			func() {
				defer wg.Done()
				steps := shell.GetStepsArrayWithoutClean(output)
				go ExecuteUpdaeAndNotify(output, steps)
			}()

		case "teardown":
			func() {
				defer wg.Done()
				steps := shell.GetOnlyTearDownSteps(output)
				go ExecuteUpdaeAndNotify(output, steps)
			}()

		case "stopinstance":
			func() {
				defer wg.Done()
				//go StopInstanceAndNotifyUser()
			}()
		}
	}
	/*
		go func() {
			defer wg.Done()
			switch action {
			case "start":
				ExecuteUpdaeAndNotify(id, env, public_ip, private_ip)
			}
		}()
	*/

	scheduler.Every(5).Seconds().NotImmediately().Run(job)
	// Keep the program from not exiting.
	runtime.Goexit()
}

func getDB() {
	var err error

	db, err = sql.Open("mysql", "root:hike@tcp(10.128.20.71:3306)/docker")
	if err != nil {
		fmt.Print(err.Error())
		panic("Not able to Connect To DataBase")
	}
	err = db.Ping()
	if err != nil {
		fmt.Print("Error :", err)
	}
}

func FetchOneCandidate(db *sql.DB) map[string]string { //(int, string, string, bool, string, string, string) {
	var (
		id               int
		env              string
		instanceID       string
		clearPreviousRun bool
		publicIP         string
		privateIP        string
		msgHost          string
		platformHost     string
		growthHost       string
		action           string
		output           = make(map[string]string)
	)
	var query = "SELECT id, env, instance_id, clear_previous_run, public_ip, private_ip, msg_host, platform_host, growth_host, action from jobs WHERE status = 0 and in_progress = 0 ORDER BY created_at limit 1"
	rows, err := db.Query(query)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&id, &env, &instanceID, &clearPreviousRun, &publicIP, &privateIP, &msgHost, &platformHost, &growthHost, &action)
	}
	output["id"] = strconv.Itoa(id)
	output["env"] = env
	output["instanceID"] = instanceID
	output["clearPreviousRun"] = strconv.FormatBool(clearPreviousRun)
	output["publicIP"] = publicIP
	output["privateIP"] = privateIP
	output["msgHost"] = msgHost
	output["platformHost"] = platformHost
	output["growthHost"] = growthHost
	output["action"] = action
	return output
	//return id, env, instance_id, clear_previous_run, public_ip, private_ip, action
}

func remoteShellExecutor(steps []*shell.Input) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("\r\nFatal: %s \n\n", err)
			os.Exit(1)
		}
	}()
	for _, val := range steps {
		app := shell.NewGrape(val)
		app.VerifyAction()
		app.Run()
	}
}

func UpdateProgress(id string) bool {
	var (
		query = "UPDATE jobs SET in_progress = 1 where id = ?"
	)
	result, err := db.Exec(query, &id)
	CheckError(err)
	if rows, _ := result.RowsAffected(); rows > 0 {
		return true
	} else {
		return false
	}
}

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error --->", err.Error())
	}
}

func ExecuteUpdaeAndNotify(output map[string]string, steps []*shell.Input) {
	//Update progress status
	UpdateProgress(output["id"])

	//Execute steps
	remoteShellExecutor(steps)

	//Update the job status from 0 to 1 and NotifyUser
	if server.UpdateJob(output["id"], db) {
		email := server.GetUserEmailId(output["id"], db)
		api := slack.New(SLACK_TOKEN)
		users, err := api.GetUsers()
		if err != nil {
			fmt.Println(err)
		}
		for _, val := range users {
			if email == val.Profile.Email {
				server.NotifyUser(val.ID, output["env"]+":"+output["publicIP"]+":"+output["privateIP"]+":"+strings.ToUpper(output["action"]))
			}
		}
	}
}
