package remoteshell

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

const (
	KEY_PATH     = "/Users/vivekvasvani/.ssh/id_rsa"
	MSG_YML      = "../remoteshell/messaging.yml"
	PLATFORM_YML = "../remoteshell/platform.yml"
	GROWTH_YML   = "../remoteshell/growth.yml"
)

//Total Steps 4 : create --> clone --> start
func GetStepsArrayWithoutClean(output map[string]string) []*Input {
	var (
		YML_LOCATION      string
		remothshellstruct = make([]*Input, 3)
	)
	YML_LOCATION = GetUpdatedYMLLocation(output)
	remothshellstruct[0] = CreateRemoteShellStruct(true, false, "create", "prod", KEY_PATH, YML_LOCATION)
	remothshellstruct[1] = CreateRemoteShellStruct(true, true, "clone", "prod", KEY_PATH, YML_LOCATION)
	remothshellstruct[2] = CreateRemoteShellStruct(true, false, "start", "prod", KEY_PATH, YML_LOCATION)
	return remothshellstruct
}

//Total Steps 4 : clean --> create --> clone --> start
func GetStepsArrayWithClean(output map[string]string) []*Input {
	var (
		YML_LOCATION      string
		remothshellstruct = make([]*Input, 4)
	)
	YML_LOCATION = GetUpdatedYMLLocation(output)
	remothshellstruct[0] = CreateRemoteShellStruct(true, false, "clean", "prod", KEY_PATH, YML_LOCATION)
	remothshellstruct[1] = CreateRemoteShellStruct(true, false, "create", "prod", KEY_PATH, YML_LOCATION)
	remothshellstruct[2] = CreateRemoteShellStruct(true, true, "clone", "prod", KEY_PATH, YML_LOCATION)
	remothshellstruct[3] = CreateRemoteShellStruct(true, false, "start", "prod", KEY_PATH, YML_LOCATION)
	return remothshellstruct
}

//Total Steps 1 : teardown
func GetOnlyTearDownSteps(output map[string]string) []*Input {
	var (
		YML_LOCATION      string
		remothshellstruct = make([]*Input, 1)
	)
	YML_LOCATION = GetUpdatedYMLLocation(output)
	remothshellstruct[0] = CreateRemoteShellStruct(true, false, "teardown", "prod", KEY_PATH, YML_LOCATION)
	return remothshellstruct
}

func ReturnYMLFileLocation(env string) (YML_LOCATION string) {
	switch env {
	case "msg":
		YML_LOCATION = MSG_YML
	case "platform":
		YML_LOCATION = PLATFORM_YML
	case "growth":
		YML_LOCATION = GROWTH_YML
	}
	return
}

func CreateRemoteShellStruct(verifyFlagB, asyncFlagB bool, commandName, serverGroup, keyPath, ymlLocation string) *Input {
	return &Input{
		VerifyFlag:  VerifyFlag(verifyFlagB),  //Enable force execution of commands
		AsyncFlag:   AsyncFlag(asyncFlagB),    //false for synchronous
		CommandName: CommandName(commandName), //command name in yml file
		ServerGroup: ServerGroup(serverGroup), //Server Group
		KeyPath:     KeyPath(keyPath),         //public key path
		ConfigPath:  ConfigPath(ymlLocation),  //yml file location
	}
}

func GetReplaceAndCreate(sessionMap map[string]string, baseFileLocation string) (bool, string, string) {
	targetFileLocation := strings.Replace(baseFileLocation, ".yml", "_backup.yml", -1)
	input, err := ioutil.ReadFile(baseFileLocation)
	if err != nil {
		fmt.Println(err)
		return false, "File Not Found", ""
	}

	valuesToBeReplaced := CreateMapForYML(sessionMap)
	fmt.Println("Here is the map :   ", valuesToBeReplaced)
	for i, value := range valuesToBeReplaced {
		if strings.ContainsAny(string(input), "${"+strconv.Itoa(i)) {
			input = bytes.Replace(input, []byte("${"+strconv.Itoa(i)+"}"), []byte(value), -1)
		}
	}
	if err = ioutil.WriteFile(targetFileLocation, input, 0666); err != nil {
		return false, "Error in writing file on file-systems", ""
	}
	fmt.Println(string(input))
	return false, targetFileLocation, targetFileLocation
}

func CreateMapForYML(output map[string]string) []string {
	fmt.Println(output)
	var mapForYml = make([]string, 3)
	switch output["env"] {
	case "msg":
		mapForYml[0] = output["privateIP"]
		mapForYml[1] = output["growthHost"]
		mapForYml[2] = output["platformHost"]
	case "platform":
		mapForYml[0] = output["privateIP"]
		mapForYml[1] = output["msgHost"]
	case "growth":
		mapForYml[0] = output["privateIP"]
		mapForYml[1] = output["msgHost"]
	}
	fmt.Println("----------", mapForYml)
	return mapForYml
}

func GetUpdatedYMLLocation(output map[string]string) string {
	YML_LOCATION := ReturnYMLFileLocation(output["env"])
	_, err, targetLocation := GetReplaceAndCreate(output, YML_LOCATION)
	if err != "" {
		fmt.Println(err)
	}
	return targetLocation
}
