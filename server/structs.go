package server

type Request struct {
	EnvType        string `json:"envType"`
	MsgHostIP      string `json:"msgHostIP"`
	GrowthHostIP   string `json:"growthHostIP"`
	PlatformHostIP string `json:"platformHostIp"`
}

//Response status mesaage
type StatusMessage struct {
	StatusCode string `json:"statusCode"`
	StatusType string `json:"statusType"`
	Message    string `json:"message"`
}

//Response structure
type Response struct {
	Status StatusMessage `json:"status"`
	Data   interface{}   `json:"data,omitempty"`
}

//StopInstances request
type StopInstances struct {
	InstanceIds []string `json:"instanceIds"`
}

type Data struct {
	InstanceId string `json:instanceID`
	State      string `json:currentState`
	PublicIp   string `json:publicIP`
	PrivateIp  string `json:privateIP`
}

type StartDokcerRequest struct {
	Email      string `json:"email"`
	Envdetails []struct {
		Envid              string `json:"envid"`
		InstanceID         string `json:"instanceId"`
		CleanPreviousRun   bool   `json:"cleanPreviousRun"`
		PublicIP           string `json:"publicIp"`
		PrivateIP          string `json:"privateIp"`
		MessageInfraHostIp string `json:"messageInfraHostIp"`
		PlatformHostIp     string `json:"platformHostIp"`
		GrowthHostIp       string `json:"growthHostIp"`
		Action             string `json:"action"`
	} `json:"envdetails"`
}
