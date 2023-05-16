package service

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"io"
	"log"
	"oj/Entity"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RunResult struct {
	Output    string  `json:"output"`
	Time      float64 `json:"time"`
	Memory    string  `json:"memory"`
	Exception string  `json:"exception"`
	Number    int     `json:"number"`
}

func getPrefix(language string) string {
	switch language {
	case "c++":
		return "cpp"
	case "python":
		return "py"
	case "java":
		return "java"
	case "go":
		return "go"
	}

	return ""
}

func saveCode(code, language string) (string, error) {
	language = getPrefix(language)
	filename := fmt.Sprintf("%s.%s", uuid.New(), language)
	err := os.WriteFile(filename, []byte(code), 0644)
	if err != nil {
		return "", err
	}
	return filename, nil
}

func compile(path, language string) ([]string, error) {
	switch language {
	case "go":
		command := []string{"/root/go/oj/" + path[0:len(path)-4]}
		cmd := exec.Command("go", "build", path)
		err := cmd.Run()
		os.Chown(path[0:len(path)-4], 0, 0)
		_ = os.Chmod(path[0:len(path)-4], 0755)
		return command, err
	case "c++":
		cmd := exec.Command("g++", "-o", path[0:len(path)-5], "-Wall", "-O2", path)
		err := cmd.Run()
		os.Chown(path[0:len(path)-5], 0, 0)
		_ = os.Chmod(path[0:len(path)-5], 0755)
		command := []string{"/root/go/oj/" + path[0:len(path)-4]}
		return command, err
	case "python":
		command := []string{"python3", path}
		return command, nil
	}
	return nil, fmt.Errorf("语言类型不匹配")
}

//func run(path string, input, language string) (string, float64, error) {
//
//	cmd := exec.Command("./main1")
//	//cmd.SysProcAttr = &syscall.SysProcAttr{
//	//	// Limit the process to read-only access to the file system.
//	//	// Ensures that the process cannot write to disk.
//	//	Chroot: "/sandbox",
//	//	// Mount a tmpfs filesystem at /sandbox, ensuring that the process cannot read from disk.
//	//	// Ensures that the process cannot read from disk.
//	//	MountNamespace: true,
//	//	// Limit the process to read-only access to /dev, /proc, and /sys.
//	//	// Ensures that the process cannot see host's kernel information.
//	//	CloneFlags: syscall.CLONE_NEWNS | syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID,
//	//}
//	// 使用管道将输入数据传递给进程。
//	stdin, err := cmd.StdinPipe()
//	if err != nil {
//		return "", 0, err
//	}
//	// 使用管道捕获进程的输出数据。
//	stdout, err := cmd.StdoutPipe()
//	if err != nil {
//		return "", 0, err
//	}
//	// 启动流程。
//	start := time.Now()
//	if err := cmd.Start(); err != nil {
//		return "", 0, err
//	}
//
//	// 将输入数据写入进程的stdin
//	io.WriteString(stdin, input)
//	stdin.Close()
//	elapsed := time.Since(start)
//	// Use bufio.Scanner to read output data.
//	output, err := io.ReadAll(stdout)
//	err = cmd.Wait()
//	if err != nil {
//		return "", 0, err
//	}
//	return string(output), elapsed.Seconds(), nil
//}

func getUsage(pid int) (string, error) {
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "rss=")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}
	str := string(stdout.Bytes())
	str, _ = strings.CutSuffix(str, "\n")
	return str, nil
}

func JudgeHandler(vo Entity.ReceiveCodeVo) (int, interface{}) {
	CodePath, err := saveCode(vo.Code, vo.Language)
	if err != nil {
		return Entity.Response{}.Fail(err.Error())
	}
	defer deleteCode(CodePath)
	log.Println("create file: " + CodePath)
	command, err := compile(CodePath, vo.Language)
	if err != nil {
		return Entity.Response{}.Fail(err.Error())
	}
	return runCode(vo.TestPoints, command)
}

func deleteCode(path string) {
	exec.Command("rm", path)
}

func execCode(input string, command []string) RunResult {
	name := command[0]
	// ./文件名
	cmd := exec.Command(name, command[1:]...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return RunResult{Exception: err.Error()}
	}
	// 处理获取标准错误输出失败的情况
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return RunResult{Exception: err.Error()}
	}
	// 使用管道捕获进程的输出数据。
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return RunResult{Exception: err.Error()}
	}
	// 启动流程。
	start := time.Now()
	if err := cmd.Start(); err != nil {
		return RunResult{Exception: err.Error()}
	}
	memory, _ := getUsage(cmd.Process.Pid)
	// 将输入数据写入进程的stdin
	_, _ = io.WriteString(stdin, input)
	_ = stdin.Close()
	var buf bytes.Buffer
	// 处理 获取标准错误输出失败的情况
	if _, err := io.Copy(&buf, stderr); err != nil {
		return RunResult{Exception: err.Error()}
	}
	elapsed := time.Since(start)
	output, err := io.ReadAll(stdout)
	// 处理执行代码失败
	if err := cmd.Wait(); err != nil {
		// 处理py异常信息
		if buf.Len() > 0 {
			return RunResult{Exception: buf.String()}
		}
	}
	result := RunResult{
		Output: string(output),
		Time:   elapsed.Seconds(),
		Memory: memory,
	}
	return result
}
func runCode(testPoints []string, command []string) (int, interface{}) {
	var ans Entity.Response
	var wg sync.WaitGroup
	resList := make([]RunResult, 0)
	for _, val := range testPoints {
		wg.Add(1)
		go func(val string, group *sync.WaitGroup) {
			defer group.Done()
			res := execCode(val, command)
			resList = append(resList, res)
		}(val, &wg)
	}
	wg.Wait()
	ans.Data = resList
	return 200, ans
}
