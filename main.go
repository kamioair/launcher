package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/kamioair/quick-utils/qio"
	"github.com/kamioair/quick-utils/qlauncher"
	"os/exec"
	"strings"
	"sync"
)

func main() {
	qlauncher.Run(start, stop)
}

func start() {
	// 启动Broker
	_ = openBroker()

	// 启动网络发现模块

	// 启动客户端管理模块
	openDevManager()

	// 启动路由模块
	devCode := openRoute()

	// 启动剩余模块
	openModules(devCode)
}

func stop() {

}

func openBroker() bool {
	brokerFile := qio.GetFullPath("./bin/broker/broker.exe")
	if qio.PathExists(brokerFile) == false {
		return false
	}

	// 执行命令并捕获输出
	runCmd(exec.Command(brokerFile), "Broker", nil)
	return true
}

func openDevManager() {
	brokerFile := qio.GetFullPath("./bin/device.exe")
	if qio.PathExists(brokerFile) == false {
		return
	}

	// 执行命令并捕获输出
	args := map[string]string{}
	args["ConfigPath"] = "../config/config.yaml"
	str, _ := json.Marshal(args)
	runCmd(exec.Command(brokerFile, string(str)), "Device", nil)
}

func openRoute() string {
	routeFile := qio.GetFullPath("./bin/router.exe")
	if qio.PathExists(routeFile) == false {
		return ""
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	code := ""
	// 执行命令并捕获输出
	args := map[string]string{}
	args["ConfigPath"] = "../config/config.yaml"
	str, _ := json.Marshal(args)
	runCmd(exec.Command(routeFile, string(str)), "Broker", func(line string) {
		if strings.HasPrefix(line, "[DeviceCode]:") {
			code = strings.TrimPrefix(line, "[DeviceCode]:")
			code = strings.Trim(code, " ")
			wg.Done()
		}
	})

	wg.Wait()
	return code
}

func openModules(devCode string) {
	files, err := qio.GetFiles("./bin")
	if err != nil {
		return
	}
	for _, f := range files {
		ext := qio.GetFileExt(f)
		if strings.ToLower(ext) != ".exe" {
			continue
		}
		name := strings.ToLower(qio.GetFileName(f))
		if name == "broker.exe" || name == "device.exe" || name == "router.exe" {
			continue
		}

		args := map[string]string{}
		args["ConfigPath"] = "../config/config.yaml"
		args["DeviceCode"] = devCode
		str, _ := json.Marshal(args)
		file := qio.GetFullPath(f)
		runCmd(exec.Command(file, string(str)), qio.GetFileNameWithoutExt(f), nil)
	}
}

func runCmd(cmd *exec.Cmd, name string, onLog func(line string)) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		// 获取标准输出的管道
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			fmt.Printf("%s Error obtaining stdout: %s\n", name, err)
			return
		}

		// 启动命令
		if err := cmd.Start(); err != nil {
			fmt.Printf("%s Error starting command: %s\n", name, err)
			return
		}

		// 创建一个扫描器来读取标准输出
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line)
			if onLog != nil {
				onLog(line)
			}
			if strings.Contains(line, "Started OK") {
				wg.Done()
			}
		}
	}()
	wg.Wait()
}
