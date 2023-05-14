package service

import (
	"bytes"
	"fmt"
	"io"
	"oj/Entity"
	"oj/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type RunResult struct {
	Status string  `json:"status"`
	Output string  `json:"output"`
	Time   float64 `json:"time"`
	Memory float64 `json:"memory"`
}

func getPrefix(language string) string {
	switch language {
	case "c++":
		return "cpp"
	case "python":
		return "py"
	case "java":
		return "java"

	}
	return ""
}

func saveCode(code, language string) (string, error) {
	language = getPrefix(language)
	filename := fmt.Sprintf("%s.%s", "main1", language)
	path := filepath.Join("./", filename)
	err := os.WriteFile(path, []byte(code), 0644)
	if err != nil {
		return "", err
	}
	return path, nil
}

func compile(path, language string) (*exec.Cmd, error) {
	switch language {
	case "go":
		cmd := exec.Command("go", "build", path)
		return cmd, cmd.Run()
	case "c++":
		cmd := exec.Command("g++", "main", path, "-o")
		return cmd, cmd.Run()
	}
}
func run(path string, input, language string) (string, float64, error) {

	switch language {
	case "python":

	}
	cmd := exec.Command(path, "run", "main1."+language)
	//cmd.SysProcAttr = &syscall.SysProcAttr{
	//	//// Limit the process to read-only access to the file system.
	//	//// Ensures that the process cannot write to disk.
	//	//Chroot: "/sandbox",
	//	//// Mount a tmpfs filesystem at /sandbox, ensuring that the process cannot read from disk.
	//	//// Ensures that the process cannot read from disk.
	//	//MountNamespace: true,
	//	//// Limit the process to read-only access to /dev, /proc, and /sys.
	//	//// Ensures that the process cannot see host's kernel information.
	//	//CloneFlags: syscall.CLONE_NEWNS | syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID,
	//}
	// 使用管道将输入数据传递给进程。
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", 0, err
	}
	// 使用管道捕获进程的输出数据。
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", 0, err
	}
	// 启动流程。
	start := time.Now()
	if err := cmd.Start(); err != nil {
		return "", 0, err
	}

	// 将输入数据写入进程的stdin
	io.WriteString(stdin, input)
	stdin.Close()
	elapsed := time.Since(start)
	// Use bufio.Scanner to read output data.
	output, err := io.ReadAll(stdout)
	err = cmd.Wait()
	if err != nil {
		return "", 0, err
	}
	return string(output), elapsed.Seconds(), nil
}

func getUsage(pid int) (float64, error) {
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "%cpu,%mem")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return 0, err
	}
	output := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(output) < 2 {
		return 0, fmt.Errorf("cannot get usage for pid %d", pid)
	}
	values := strings.Fields(output[1])
	if len(values) < 2 {
		return 0, fmt.Errorf("cannot get usage for pid %d", pid)
	}
	_, err := strconv.ParseFloat(values[0], 64)
	if err != nil {
		return 0, fmt.Errorf("cannot parse CPU usage for pid %d: %v", pid, err)
	}
	mem, err := strconv.ParseFloat(values[1], 64)
	if err != nil {
		return 0, fmt.Errorf("cannot parse memory usage for pid %d: %v", pid, err)
	}
	return mem, nil
}

func runCode(path, input, language string) (*RunResult, error) {

	ans, seconds, err := run(path, input, language)
	if err != nil {
		return nil, err
	}
	result := &RunResult{
		Status: "correct",
		Output: ans,
		Time:   seconds,
	}
	return result, nil
}

func Judge(vo Entity.ReceiveCodeVo) (int, interface{}) {
	switch vo.Language {
	case "c++":
		return cPlusJuegeMent(vo)
	case "go":
		return goJudgeMent(vo)
	case "python":
		return pythonJudgeMent(vo)
	default:
		return 500, fmt.Errorf("代码类型错误")
	}
}

func pythonJudgeMent(vo Entity.ReceiveCodeVo) (int, interface{}) {
	path, err := saveCode(vo.Code, vo.Language)
	if err != nil {
		return 500, err
	}
	ans, err := runPython(path, vo.TestPoints)
	if err != nil {
		return 500, err
	}
	return 200, ans
}

func runPython(path string, input []string) (interface{}, error) {
	var res Entity.Response
	cmd := exec.Command("pythton", path)
	for _, val := range input {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return "", err
		}
		// 使用管道捕获进程的输出数据。
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return "", err
		}
		// 启动流程。
		start := time.Now()
		if err := cmd.Start(); err != nil {
			return "", err
		}

		// 将输入数据写入进程的stdin
		io.WriteString(stdin, val)
		stdin.Close()
		elapsed := time.Since(start)
		output, err := io.ReadAll(stdout)
		err = cmd.Wait()
		if err != nil {
			return "", err
		}
		res.Code = 200
		res.Data = RunResult{
			Status: "correct",
			Output: string(output),
			Time:   elapsed.Seconds(),
			Memory: 0,
		}
	}
}

func goJudgeMent(vo Entity.ReceiveCodeVo) (int, interface{}) {
	var response Entity.Response
	response.Code = 200
	var ans []RunResult
	path, err := saveCode(vo.Code, vo.Language)

	if err != nil {
		return utils.Fail(err.Error())
	}

	// 编译代码
	cmd, err := compile(path, vo.Language)
	if err != nil {
		return utils.Fail(err.Error())
	}
	// 跑每一个数据点

	for _, val := range vo.TestPoints {
		code, err := runCode(cmd.Path, val, getPrefix(vo.Language))
		if err != nil {
			return utils.Fail(err.Error())
		}
		ans = append(ans, *code)
	}
	response.Data = ans
	return response.Code, response

}

func cPlusJuegeMent(vo Entity.ReceiveCodeVo) (int, interface{}) {

}
