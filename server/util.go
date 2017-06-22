package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"strings"

	"strconv"

	"github.com/golang/glog"
	"github.com/nlopes/slack"
	"github.com/valyala/fasthttp"
)

func SetErrorResponse(ctx *fasthttp.RequestCtx, statusCode, statusType, statusMessage string, httpStatus int) {
	log.Println(statusCode, statusType, statusMessage)
	var response Response
	response.Status.StatusCode = statusCode
	response.Status.StatusType = statusType
	response.Status.Message = statusMessage
	glog.Infoln("Error Reponse " + ToJsonString(response))
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Response.SetBodyString(ToJsonString(response))
	ctx.SetStatusCode(httpStatus)
}

func SetSuccessResponse(ctx *fasthttp.RequestCtx, statusCode, statusType, statusMessage string, httpStatus int, data interface{}) {
	var response Response
	response.Status.StatusCode = statusCode
	response.Status.StatusType = statusType
	response.Status.Message = statusMessage
	response.Data = data
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Response.SetBodyString(ToJsonString(response))
	glog.Infoln("Success Reponse " + ToJsonString(response))
	ctx.SetStatusCode(httpStatus)
}

func UpdateStatusAddIPs(to int, public, private, instanceId string, db *sql.DB) bool {
	var query = "UPDATE machines_info SET public_ip = ?, private_ip= ?, status = ? where instance_id = ?;"
	//err := db.Prepare(query)
	result, err := db.Exec(query, &public, &private, &to, &instanceId)
	checkError(err)
	if rowscount, _ := result.RowsAffected(); rowscount > 0 {
		return true
	} else {
		return false
	}
}

func UpdateStatus(to int, instanceId string, db *sql.DB) bool {
	var query = "UPDATE machines_info SET status = ? where instance_id = ?"
	//err := db.Prepare(query)
	result, err := db.Exec(query, &to, &instanceId)
	checkError(err)
	if rowscount, _ := result.RowsAffected(); rowscount > 0 {
		return true
	} else {
		return false
	}
}

func CheckServerAvailability(serverType string, db *sql.DB) (instanceName, region string) {
	var query = "select instance_id, region from machines_info where status = 4 and stack = " + GetTheStackValue(serverType) + " limit 1"
	rows, err := db.Query(query)
	defer rows.Close()
	checkError(err)
	for rows.Next() {
		rows.Scan(&instanceName, &region)
	}
	return
}

func CheckWheatherBeingUsed(env, instanceId string, db *sql.DB) (inUse int) {
	var query = "SELECT in_use FROM machines_info WHERE instance_id='" + instanceId + "' and stack = " + GetTheStackValue(env)
	rows, err := db.Query(query)
	defer rows.Close()
	checkError(err)
	for rows.Next() {
		rows.Scan(&inUse)
	}
	return
}

func CheckWheatherBelongsToSameStack(stack, instanceId string, db *sql.DB) bool {
	var (
		query               = "SELECT instance_id FROM machines_info WHERE stack=" + GetTheStackValue(stack) + " and status = 2"
		instanceIdFromTable string
	)
	instanceIdsSlice := make([]string, 10)
	rows, err := db.Query(query)
	defer rows.Close()
	checkError(err)
	for rows.Next() {
		rows.Scan(&instanceIdFromTable)
		instanceIdsSlice = append(instanceIdsSlice, instanceIdFromTable)
	}
	//return SearchInList(instanceIdsSlice, instanceId)
	return true
}

func InsertIntoJobs(createdBy, envType, instanceId, action, public, private, msgHost, platformHost, growthHost string, clearprevious bool, status int, db *sql.DB) bool {
	var (
		query = "INSERT INTO jobs (created_by, created_at, updated_at, env, instance_id, clear_previous_run, public_ip, private_ip, msg_host, platform_host, growth_host, action, status, in_progress)" +
			"VALUES (?, now(), now(), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	)
	result, err := db.Exec(query, &createdBy, &envType, &instanceId, &clearprevious, &public, &private, &msgHost, &platformHost, &growthHost, &action, &status, 0)
	checkError(err)
	if rows, _ := result.RowsAffected(); rows > 0 {
		return true
	} else {
		return false
	}
}

func UpdateJob(id string, db *sql.DB) bool {
	var (
		query = "UPDATE jobs SET status=1 where id = ?"
	)
	result, err := db.Exec(query, &id)
	checkError(err)
	if rows, _ := result.RowsAffected(); rows > 0 {
		return true
	} else {
		return false
	}
}

func GetUserEmailId(id string, db *sql.DB) (email string) {
	var (
		query = "SELECT created_by FROM jobs WHERE id = " + id
	)
	rows, err := db.Query(query)
	defer rows.Close()
	checkError(err)
	for rows.Next() {
		rows.Scan(&email)
	}
	return
}

func NotifyUser(userId, message string) {
	var colors = GetEnvColorMapping()
	detailsToSend := strings.Split(message, ":")
	api := slack.New(SLACK_TOKEN)
	params := slack.PostMessageParameters{
		Username: "Docker-Server",
	}
	fields := []slack.AttachmentField{
		slack.AttachmentField{
			Title: "Docker Environment Name",
			Value: GetText(detailsToSend[0]),
			Short: true,
		},
		slack.AttachmentField{
			Title: "Public IP",
			Value: detailsToSend[1],
			Short: true,
		},
		slack.AttachmentField{
			Title: "Private IP",
			Value: detailsToSend[2],
			Short: true,
		},
		slack.AttachmentField{
			Title: "Action Name",
			Value: detailsToSend[3],
			Short: true,
		},
	}
	attachment := slack.Attachment{
		Pretext: "",
		Text:    "",
		Fields:  fields,
		Color:   colors[detailsToSend[0]],
	}
	params.Attachments = []slack.Attachment{
		attachment,
	}
	channelID, timestamp, err := api.PostMessage(userId, "Details of your server(s)", params)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)

}

func GetTheStackValue(serverType string) string {
	switch serverType {
	case "msg":
		return "1"
	case "platform":
		return "2"
	case "growth":
		return "3"
	}
	return "0"
}

func GetText(serverType string) string {
	switch serverType {
	case "msg":
		return "Messaging Infra Docker Environment"
	case "platform":
		return "Platform Docker Environment"
	case "growth":
		return "Growth Docker Environment"
	}
	return "ERROR"
}

func GetEnvColorMapping() (colormap map[string]string) {
	colormap = make(map[string]string)
	colormap["msg"] = "good"
	colormap["platform"] = "warning"
	colormap["growth"] = "danger"
	return
}

func ToJsonString(p interface{}) string {
	bytes, err := json.Marshal(p)
	if err != nil {
		fmt.Println("Error", err.Error())
	}
	return string(bytes)
}

func SearchInList(list []string, valueToBeSearched string) bool {
	var result bool = false
	for _, val := range list {
		fmt.Println(val, "    :   ", valueToBeSearched)
		if val == valueToBeSearched {
			result = true
		}
	}
	return result
}

func GetJobDetailsById(jobId string, db *sql.DB) (message string) {
	var (
		env, publicIP, privateIP, action string
	)
	id, _ := strconv.Atoi(jobId)
	query := "SELECT env, public_ip, private_ip, action FROM jobs WHERE id=?"
	rows, err := db.Query(query, id)
	defer rows.Close()
	checkError(err)
	for rows.Next() {
		rows.Scan(&env, &publicIP, &privateIP)
	}
	message = env + ":" + publicIP + ":" + privateIP + ":" + action
	return
}
