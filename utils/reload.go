package utils

import (
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strconv"
)

//Reload 重启服务
func Reload(sss ...func()) error {
	if runtime.GOOS == "windows" {
		return errors.New("该系统不支持 reload ")
	}

	//重启之前需要做一些事情，比如服务注销、数据库关闭等等
	for _, f := range sss {
		f()
	}

	pid := os.Getpid()
	cmd := exec.Command("kill", "-1", strconv.Itoa(pid))
	return cmd.Run()
}