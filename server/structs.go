package server

type Request []struct {
	EnvType        string `json:"envType"`
	GrowthHostIP   string `json:"growthHostIP,omitempty"`
	PlatformHostIP string `json:"platformHostIp,omitempty"`
	MsgHostIP      string `json:"msgHostIP,omitempty"`
	GrowthHostIP   string `json:"growthHostIp,omitempty"`
}
