package main

import (
	"bufio"
	"fmt"
	"github.com/mattn/go-colorable"
	"io"
	"os/exec"
	"strings"
	"sync"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
)

var ColorStdout io.Writer

func init() {
	ColorStdout = colorable.NewColorableStdout()
}

func runCmd(cmd *exec.Cmd, name string, onLog func(line string) string) {
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
		exLog := ""
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			//log.Println(line)
			if onLog != nil {
				l := onLog(line)
				if l != "" {
					exLog += l + " "
				}
			}
			if strings.Contains(line, "Started OK") {
				wg.Done()
				if exLog != "" {
					_, _ = fmt.Fprintln(ColorStdout, Reset+" ->", name, Green+"OK", White+fmt.Sprintf("(%s)", strings.Trim(exLog, " ")))
				} else {
					_, _ = fmt.Fprintln(ColorStdout, Reset+" ->", name, Green+"OK")
				}
			}
		}
	}()
	wg.Wait()
}
