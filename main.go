package main

import (
	"encoding/json"
	"fmt"
	"github.com/kamioair/qf/utils/qconfig"
	"github.com/kamioair/qf/utils/qio"
	"github.com/kamioair/qf/utils/qlauncher"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var (
	configPath = "../config/config.yaml"
)

func main() {
	qlauncher.Run(start, nil)
}

func start() {

	launcherConfig := "./launcher.yaml"
	qconfig.ChangeFilePath(launcherConfig)
	configPath = qconfig.Get("base", "config", "./config/config.yaml")
	configPath = qio.GetFullPath(configPath)

	_, _ = fmt.Fprintln(ColorStdout, Reset+"--------------------------------------")
	// 启动网络发现模块

	// 启动Broker
	openBroker()

	// 启动路由模块
	devCode := openRoute()
	if devCode == "" {
		return
	}

	// 启动功能模块
	modules := qconfig.Get("", "modules", []string{})
	for _, m := range modules {
		openModules(m, devCode)
	}
	//openModules(devCode)

	_, _ = fmt.Fprintln(ColorStdout, Reset+"--------------------------------------")
}

func openBroker() {
	brokerFile := qio.GetFullPath("./bin/broker.exe")
	if qio.PathExists(brokerFile) == false {
		return
	}

	// 执行命令并捕获输出
	runCmd(exec.Command(brokerFile, configPath), "broker", func(line string) string {
		if strings.Contains(line, "ws://127.0.0.1") {
			sp := strings.Split(line, ":")
			ssp := strings.Split(sp[len(sp)-1], "/")
			return fmt.Sprintf("Port:%s", ssp[0])
		}
		return ""
	})
	time.Sleep(time.Second * 2)
}

func openRoute() string {
	routeFile := qio.GetFullPath("./bin/route.exe")
	if qio.PathExists(routeFile) == false {
		return ""
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	code := ""

	// 执行命令并捕获输出
	args := map[string]string{}
	args["ConfigPath"] = configPath
	str, _ := json.Marshal(args)
	runCmd(exec.Command(routeFile, string(str)), "route", func(line string) string {
		if strings.HasPrefix(line, "[DeviceInfo]:") {
			c := strings.TrimPrefix(line, "[DeviceInfo]:")
			c = strings.Trim(c, " ")
			sp := strings.Split(c, "^")
			code = sp[0]
			wg.Done()
			return fmt.Sprintf("Id:%s Name:%s", code, sp[1])
		}
		return ""
	})

	wg.Wait()

	time.Sleep(time.Second * 2)
	return code
}

func openModules(modulePath string, devCode string) {
	modulePath = qio.GetFullPath(modulePath)
	if qio.PathExists(modulePath) == false {
		return
	}

	args := map[string]string{}
	args["ConfigPath"] = configPath
	args["DeviceCode"] = devCode
	str, _ := json.Marshal(args)
	runCmd(exec.Command(modulePath, string(str)), qio.GetFileNameWithoutExt(modulePath), nil)

	time.Sleep(time.Millisecond * 500)
}

//
//func openDevManager() {
//	brokerFile := qio.GetFullPath("./bin/device.exe")
//	if qio.PathExists(brokerFile) == false {
//		return
//	}
//
//	// 执行命令并捕获输出
//	args := map[string]string{}
//	args["ConfigPath"] = "../config/config.yaml"
//	str, _ := json.Marshal(args)
//	runCmd(exec.Command(brokerFile, string(str)), "Device", nil)
//}
//
//func openRoute() string {
//	routeFile := qio.GetFullPath("./bin/router.exe")
//	if qio.PathExists(routeFile) == false {
//		return ""
//	}
//
//	wg := sync.WaitGroup{}
//	wg.Add(1)
//	code := ""
//	// 执行命令并捕获输出
//	args := map[string]string{}
//	args["ConfigPath"] = "../config/config.yaml"
//	str, _ := json.Marshal(args)
//	runCmd(exec.Command(routeFile, string(str)), "Broker", func(line string) {
//		if strings.HasPrefix(line, "[DeviceCode]:") {
//			code = strings.TrimPrefix(line, "[DeviceCode]:")
//			code = strings.Trim(code, " ")
//			wg.Done()
//		}
//	})
//
//	wg.Wait()
//	return code
//}
//
