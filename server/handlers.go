package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/nlopes/slack"
	"github.com/valyala/fasthttp"
	shell "github.com/vivekvasvani/docker-server/remoteshell"
	utils "github.com/vivekvasvani/docker-server/utils"
)

var wg sync.WaitGroup

const SLACK_TOKEN = ""

func CheckAndStartInstance(ctx *fasthttp.RequestCtx, db *sql.DB) {
	var request Request
	err := json.Unmarshal(ctx.Request.Body(), &request)
	checkError(err)
	switch request.EnvType {
	case "msg":
		instanceId, region := CheckServerAvailability(request.EnvType, db)
		instanceId = strings.TrimSpace(instanceId)
		//If Instance is empty i.e. no servers are available
		if instanceId == "" || region == "" {
			SetErrorResponse(ctx, "9001", "ERROR", "No server aviable", http.StatusOK)
			return
		} else {
			// Start the instance and call describe to get the status
			current, previous := utils.StartInstance(instanceId)
			//Send failure response if instance is already in running state
			if previous == "running" {
				SetErrorResponse(ctx, "9002", "ERROR", "Can not start instance as the current state is :"+previous, http.StatusOK)
			}

			if previous == "stopped" && current == "pending" {
				ok, state, publicip, privateip, err := CheckStatusOfDescribe("running", instanceId)
				UpdateStatusAddIPs(2, publicip, privateip, instanceId, db)
				if ok {
					SetSuccessResponse(ctx, "2001", "SUCCESS", "Successfully Retrived Information from EC2", http.StatusOK,
						&Data{
							InstanceId: instanceId,
							State:      state,
							PublicIp:   publicip,
							PrivateIp:  privateip,
						})
				} else {
					SetErrorResponse(ctx, "5001", "ERROR", err.Error(), http.StatusOK)
				}
			} else {
				SetErrorResponse(ctx, "9002", "ERROR", "Not able to start instance", http.StatusInternalServerError)
			}

		}

	case "devx":

	case "labs":

	}
}

func StartInstancesGet(ctx *fasthttp.RequestCtx, db *sql.DB) {
	instanceId := ctx.UserValue("instanceid").(string)
	current, previous := utils.StartInstance(instanceId)
	fmt.Println(current, previous)
	if previous == "running" {
		SetErrorResponse(ctx, "9002", "ERROR", "Server is already in running state", http.StatusInternalServerError)
	}
	if previous == "stopped" && current == "pending" {
		ok, state, publicip, privateip, err := CheckStatusOfDescribe("running", instanceId)
		UpdateStatusAddIPs(2, publicip, privateip, instanceId, db)
		if ok {
			SetSuccessResponse(ctx, "2001", "SUCCESS", "Successfully Retrived Information from EC2", http.StatusOK,
				&Data{
					InstanceId: instanceId,
					State:      state,
					PublicIp:   publicip,
					PrivateIp:  privateip,
				})
		} else {
			SetErrorResponse(ctx, "5001", "ERROR", err.Error(), http.StatusOK)
		}
	} else {
		SetErrorResponse(ctx, "9002", "ERROR", "Not able to start instance", http.StatusInternalServerError)
	}
}

func StopInstancesWithPayload(ctx *fasthttp.RequestCtx, db *sql.DB) {
	type FinalResponse struct {
		DataAll []Data `json:response`
	}
	var (
		request    StopInstances
		checkAllOk bool = true
	)

	err := json.Unmarshal(ctx.Request.Body(), &request)
	checkError(err)
	Resp := make([]Data, len(request.InstanceIds))
	for _, val := range request.InstanceIds {
		state, _, _ := utils.DescribeInstances(val)
		if state == "stopping" || state == "stopped" {
			UpdateStatusAddIPs(3, "", "", val, db)
			continue
		}
		wg.Add(1)
		go utils.StopInstance(val, &wg)
	}
	wg.Wait()

	for i, val := range request.InstanceIds {
		state, publicip, privateip := utils.DescribeInstances(val)
		if state == "" {
			checkAllOk = false
		} else {
			UpdateStatusAddIPs(4, "", "", val, db)
		}
		Resp[i] = Data{
			InstanceId: val,
			State:      state,
			PublicIp:   publicip,
			PrivateIp:  privateip,
		}
	}

	if checkAllOk {
		SetSuccessResponse(ctx, "2001", "SUCCESS", "Successfully executed ec2-stop-instances", http.StatusOK,
			&FinalResponse{
				DataAll: Resp,
			})
	} else {
		SetErrorResponse(ctx, "5001", "ERROR", "State is unknown for some instances.", http.StatusOK)
	}

}

//To Availability in Database
func Checkavailability(ctx *fasthttp.RequestCtx, db *sql.DB) {
	type Data struct {
		InstanceId string `json:instanceID`
		Region     string `json:region`
	}
	envtype := ctx.UserValue("type").(string)
	instance, region := CheckServerAvailability(envtype, db)
	if strings.TrimSpace(instance) == "" || strings.TrimSpace(region) == "" {
		SetErrorResponse(ctx, "5001", "ERROR", "No server aviable", http.StatusOK)
	} else {
		SetSuccessResponse(ctx, "2001", "SUCCESS", "Atleast one box is available", http.StatusOK,
			&Data{
				InstanceId: instance,
				Region:     region,
			})
	}
}

//Check instance status
func CheckInstanceStatus(ctx *fasthttp.RequestCtx, db *sql.DB) {
	instanceId := ctx.UserValue("instanceid").(string)
	state, publicip, privateip := utils.DescribeInstances(instanceId)
	if state != "" {
		SetSuccessResponse(ctx, "2001", "SUCCESS", "Successfully Retrived Information from EC2", http.StatusOK,
			&Data{
				InstanceId: instanceId,
				State:      state,
				PublicIp:   publicip,
				PrivateIp:  privateip,
			})
	} else {
		SetErrorResponse(ctx, "5001", "ERROR", "Error in fetching ec2-describe-instances", http.StatusInternalServerError)
	}
}

func StartDockerEnvironment(ctx *fasthttp.RequestCtx, db *sql.DB) {
	var (
		startDockerRequest StartDokcerRequest
		createdBy          string
		env                string
		action             string
		instanceId         string
		publicIp           string
		privateIp          string
		messageInfraHostIp string
		platformHostIp     string
		growthHostIp       string
		clearPreviousRuns  bool
		//allEnvTypes        = []string{"msg", "platform", "growth"}
	)
	err := json.Unmarshal(ctx.Request.Body(), &startDockerRequest)
	//env := ctx.UserValue("env").(string)
	//action := ctx.UserValue("action").(string)
	checkError(err)
	ok, err := validateStartDockerRequest(env, ctx, db, startDockerRequest)
	if !ok {
		SetErrorResponse(ctx, "5001", "ERROR", err.Error(), http.StatusOK)
		return
	}
	createdBy = startDockerRequest.Email
	for _, pickOneEnv := range startDockerRequest.Envdetails {
		if pickOneEnv.Action != "" {
			env = pickOneEnv.Envid
			action = pickOneEnv.Action
			instanceId = pickOneEnv.InstanceID
			publicIp = pickOneEnv.PublicIP
			privateIp = pickOneEnv.PrivateIP
			messageInfraHostIp = pickOneEnv.MessageInfraHostIp
			platformHostIp = pickOneEnv.PlatformHostIp
			growthHostIp = pickOneEnv.GrowthHostIp
			clearPreviousRuns = pickOneEnv.CleanPreviousRun
			InsertIntoJobs(createdBy, env, instanceId, action, publicIp, privateIp, messageInfraHostIp, platformHostIp, growthHostIp, clearPreviousRuns, 0, db)
		}

	}
	SetSuccessResponse(ctx, "2001", "SUCCESS", "Successfully Submitted the job. We'll notify you via slack once docker is up and running", http.StatusOK, "")

	/*
		switch env {
		case "msg":
			switch action {
			case "init":
			case "start":
				for _, val := range startDockerRequest.Envdetails {
					if val.Envid == "msg" {
						createdBy = startDockerRequest.Email
						instanceId = val.InstanceID
						publicIp = val.PublicIP
						privateIp = val.PrivateIP
						clearPreviousRuns = val.CleanPreviousRun
						InsertIntoJobs(createdBy, val.Envid, instanceId, action, publicIp, privateIp, clearPreviousRuns, 0, db)
					}
				}
				SetSuccessResponse(ctx, "2001", "SUCCESS", "Successfully Submitted the job. We'll notify you via slack once docker is up and running", http.StatusOK, "")
			case "teardown":
				steps := shell.GetOnlyTearDownSteps("msg")
				remoteShellExecutor(steps)
				SetSuccessResponse(ctx, "2001", "SUCCESS", "Successfully cleared old entries", http.StatusOK, "")
			}
		case "platform":
			switch action {
			case "init":
			case "start":
				for _, val := range startDockerRequest.Envdetails {
					if val.Envid == "platform" {
						createdBy = startDockerRequest.Email
						instanceId = val.InstanceID
						publicIp = val.PublicIP
						privateIp = val.PrivateIP
						clearPreviousRuns = val.CleanPreviousRun
						InsertIntoJobs(createdBy, val.Envid, instanceId, action, publicIp, privateIp, clearPreviousRuns, 0, db)
					}
				}
				SetSuccessResponse(ctx, "2001", "SUCCESS", "Successfully Submitted the job. We'll notify you via slack once docker is up and running", http.StatusOK, "")
			case "teardown":
				steps := shell.GetOnlyTearDownSteps("platform")
				remoteShellExecutor(steps)
				SetSuccessResponse(ctx, "2001", "SUCCESS", "Successfully cleared old entries", http.StatusOK, "")
			}
		case "growth":
			if ok := InsertIntoJobsForCronScheduler(startDockerRequest, env, db); ok {
				SetSuccessResponse(ctx, "2001", "SUCCESS", "Successfully Submitted the job. We'll notify you via slack once docker is up and running", http.StatusOK, "")
			} else {
				SetErrorResponse(ctx, "9001", "ERROR", "Exception occurred while inserting data into queue", http.StatusInternalServerError)
			}

		case "all":
			switch action {
			case "init":
			case "start":
				for _, val := range startDockerRequest.Envdetails {
					if SearchInList(allEnvTypes, val.Envid) {
						createdBy = startDockerRequest.Email
						instanceId = val.InstanceID
						publicIp = val.PublicIP
						privateIp = val.PrivateIP
						clearPreviousRuns = val.CleanPreviousRun
						InsertIntoJobs(createdBy, val.Envid, instanceId, action, publicIp, privateIp, clearPreviousRuns, 0, db)
					}
				}
				SetSuccessResponse(ctx, "2001", "SUCCESS", "Successfully Submitted the job. We'll notify you via slack once docker is up and running", http.StatusOK, "")
			case "teardown":

				steps := shell.GetOnlyTearDownSteps("msg")
				remoteShellExecutor(steps)
				SetSuccessResponse(ctx, "2001", "SUCCESS", "Successfully cleared old entries", http.StatusOK, "")
			}
		}

	*/
}

func UpdateJobStatusAndNotifyUser(ctx *fasthttp.RequestCtx, db *sql.DB) {
	jobId := ctx.UserValue("jobid").(string)
	if UpdateJob(jobId, db) {
		email := GetUserEmailId(jobId, db)
		api := slack.New(SLACK_TOKEN)
		users, err := api.GetUsers()
		checkError(err)
		for _, val := range users {
			if email == val.Profile.Email {
				NotifyUser(val.ID, GetJobDetailsById(jobId, db))
			}
		}
	}
}

func CheckStatusOfDescribe(expectedState, instanceId string) (bool, string, string, string, error) {
	timeout := time.After(60 * time.Second)
	tick := time.Tick(1 * time.Second)
	var (
		state   string
		public  string
		private string
	)
	// Keep trying until we're timed out or got a result or got an error
	for {
		select {
		case <-timeout:
			return false, state, public, private, errors.New("Timed Out!!! Please check after some time")
		case <-tick:
			state, public, private = utils.DescribeInstances(instanceId)
			if strings.TrimSpace(state) == strings.TrimSpace(expectedState) {
				return true, state, public, private, nil
			}
		}
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Print(err)
	}
}

func validateStartDockerRequest(env string, ctx *fasthttp.RequestCtx, db *sql.DB, startDockerRequest StartDokcerRequest) (bool, error) {
	if !strings.HasSuffix(startDockerRequest.Email, "@hike.in") {
		return false, errors.New("Not a valid Email Id")

	}

	for _, val := range startDockerRequest.Envdetails {
		if CheckWheatherBeingUsed(env, val.InstanceID, db) == 1 {
			return false, errors.New(val.InstanceID + " Is already being used by someone else. Please select some other instance-id")

		}

		if !CheckWheatherBelongsToSameStack(val.Envid, val.InstanceID, db) {
			return false, errors.New("Either mismatch of instance id and env-type or instance is not running")
		}
	}
	return true, nil
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

/*
func InsertIntoJobsForCronScheduler(startDockerRequest StartDokcerRequest, env string, db *sql.DB) bool {
	var result bool = false
	for _, val := range startDockerRequest.Envdetails {
		if val.Envid == env {
			createdBy := startDockerRequest.Email
			instanceId := val.InstanceID
			action := val.Action
			publicIp := val.PublicIP
			privateIp := val.PrivateIP
			clearPreviousRuns := val.CleanPreviousRun
			result = InsertIntoJobs(createdBy, val.Envid, instanceId, action, publicIp, privateIp, clearPreviousRuns, 0, db)
		}
	}
	return result
}
*/
