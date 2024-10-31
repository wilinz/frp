package frpclib

import (
	"fmt"
	"github.com/fatedier/frp/cmd/frpc/sub"
	"github.com/fatedier/frp/pkg/util/system"
	"github.com/fatedier/frp/pkg/util/version"
	"strings"
)

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

func RunMultipleClientsWithUid(uids []string, cfgFilePaths []string, isStrictConfigMode bool, isShowVersion bool) (errString string) {
	system.EnableCompatibilityMode()
	if len(uids) != len(cfgFilePaths) {
		return "Error: The lengths of uids and cfgFilePaths must be the same."
	}
	if isShowVersion {
		fmt.Println(version.Full())
	}
	// Create the uidConfigMap from the slices
	uidConfigMap := make(map[string]string)
	for i := 0; i < len(uids); i++ {
		uidConfigMap[uids[i]] = cfgFilePaths[i]
	}

	// Call sub.RunMultipleClientsWithUid with the constructed map
	err := sub.RunMultipleClientsWithUid(uidConfigMap, isStrictConfigMode)
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
