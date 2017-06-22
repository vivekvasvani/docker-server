package utils

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/ec2"
)

const (
	ACCESS_KEY = "AKIAIMHDGG5YBTQ56X3A"
	SECRET_KEY = "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"
)

func DescribeInstances(instanceId string) (state, publicIp, privateIp string) {
	region := aws.Regions["ap-southeast-1"]
	auth, err := aws.GetAuth(ACCESS_KEY, SECRET_KEY, "", time.Now())
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
		os.Exit(1)
	}
	client := ec2.New(auth, region)
	resp, err := client.DescribeInstances([]string{instanceId}, nil)

	if err != nil {
		fmt.Printf("Error: %+v\n", err)
	}
	fmt.Println("Describe----->", resp, err)

	state = resp.Reservations[0].Instances[0].State.Name
	publicIp = resp.Reservations[0].Instances[0].IPAddress
	privateIp = resp.Reservations[0].Instances[0].PrivateIPAddress
	return
}

func StopInstance(instanceId string, wg *sync.WaitGroup) (state string) {
	region := aws.Regions["ap-southeast-1"]
	auth, err := aws.GetAuth(ACCESS_KEY, SECRET_KEY, "", time.Now())
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
		os.Exit(1)
	}
	client := ec2.New(auth, region)
	respS, errS := client.StopInstances(instanceId)
	if errS != nil || wg != nil {
		wg.Done()
		fmt.Println(errS)
	}
	state = respS.StateChanges[0].CurrentState.Name
	return
}

/*
ec2-Start-Instances
@param instance Id
@return currentState, previousState
*/
func StartInstance(instanceId string) (state, previous string) {
	region := aws.Regions["ap-southeast-1"]
	auth, err := aws.GetAuth(ACCESS_KEY, SECRET_KEY, "", time.Now())
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
		os.Exit(1)
	}
	client := ec2.New(auth, region)
	respS, errS := client.StartInstances(instanceId)
	fmt.Println("Start----->", respS, errS)
	state = respS.StateChanges[0].CurrentState.Name
	previous = respS.StateChanges[0].PreviousState.Name
	return
}
