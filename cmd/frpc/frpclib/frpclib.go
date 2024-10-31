package frpclib

import (
	"encoding/json"
	"fmt"
	"github.com/fatedier/frp/cmd/frpc/sub"
	"github.com/fatedier/frp/pkg/util/system"
	"github.com/fatedier/frp/pkg/util/version"
	"strings"
)

// RunMultipleClientConfig Export
type RunMultipleClientConfig sub.RunMultipleClientConfig

func RunClientWithUid(uid string, cfgFilePath string, isStrictConfigMode bool, isShowVersion bool) (errString string) {
	system.EnableCompatibilityMode()
	if isShowVersion {
		fmt.Println(version.Full())
	}
	err := sub.RunClientWithUid(uid, cfgFilePath, isStrictConfigMode)
	if err != nil {
		return err.Error()
	}
	return ""

}

func RunMultipleClientsWithUid(runMultipleClientConfigListJson string, isStrictConfigMode bool, isShowVersion bool) (errString string) {
	system.EnableCompatibilityMode()

	// 将 JSON 字符串解析为 map[string]string
	var runMultipleClientConfigList []sub.RunMultipleClientConfig
	err := json.Unmarshal([]byte(runMultipleClientConfigListJson), &runMultipleClientConfigList)
	if err != nil {
		return "Error: Failed to parse JSON input - " + err.Error()
	}

	// 如果需要显示版本信息
	if isShowVersion {
		fmt.Println(version.Full())
	}

	// 调用子模块的函数
	err = sub.RunMultipleClientsWithUid(runMultipleClientConfigList, isStrictConfigMode)
	if err != nil {
		return err.Error()
	}

	return ""
}

func Close(uid string) (ret bool) {
	return sub.Close(uid)
}

func GetUids() string {
	uids := sub.GetUids()
	return strings.Join(uids, ",")
}

func IsRunning(uid string) (running bool) {
	return sub.IsRunning(uid)
}
