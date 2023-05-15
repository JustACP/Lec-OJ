package service

import (
	"bytes"
	"fmt"
	"io"
	"oj/Entity"
	"os"
	"os/exec"
	"path/filepath"
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
	}
	return ""
}

func saveCode(code, language string) (string, error) {
	language = getPrefix(language)
	filename := fmt.Sprintf("%s.%s", "main1", language)
	path := filepath.Join("./", filename)
	path += "go"
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
	}
	return nil, fmt.Errorf("语言类型错误")
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
	//if len(output) < 2 {
	//	return 0, fmt.Errorf("cannot get usage for pid %d", pid)
	//}
	//values := strings.Fields(output[1])
	//mem, err := strconv.ParseFloat(values[1], 64)
	//if err != nil {
	//	return 0, fmt.Errorf("cannot parse memory usage for pid %d: %v", pid, err)
	//}
	//return mem, nil
}

func runCode(path, input, language string) (*RunResult, error) {

	ans, seconds, err := run(path, input, language)
	if err != nil {
		return nil, err
	}
	result := &RunResult{
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
		return Entity.Response{}.Fail(err.Error())
	}
	ans, err := runPython(path, vo.TestPoints)
	if err != nil {
		return Entity.Response{}.Fail(err.Error())
	}
	return Entity.Response{}.Success(ans)
}

func execPy(path string, input string) RunResult {
	cmd := exec.Command("python3", path)
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

func runPython(path string, input []string) (interface{}, error) {
	var res Entity.Response
	var ans []RunResult
	var wg sync.WaitGroup
	for num, val := range input {
		val := val
		wg.Add(1)
		num := num
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			res := execPy(path, val)
			res.Number = num
			ans = append(ans, res)
		}(&wg)
	}
	wg.Wait()
	res.Data = ans
	return res, nil
}

func goJudgeMent(vo Entity.ReceiveCodeVo) (int, interface{}) {
	var ans []RunResult
	path, err := saveCode(vo.Code, vo.Language)

	if err != nil {
		return Entity.Response{}.Fail(err.Error())
	}

	// 编译代码
	cmd, err := compile(path, vo.Language)
	if err != nil {
		return Entity.Response{}.Fail(err.Error())
	}
	// 跑每一个数据点

	for _, val := range vo.TestPoints {
		code, err := runCode(cmd.Path, val, getPrefix(vo.Language))
		if err != nil {
			return Entity.Response{}.Fail(err.Error())
		}
		ans = append(ans, *code)
	}
	return Entity.Response{}.Success(ans)

}

func cPlusJuegeMent(vo Entity.ReceiveCodeVo) (int, interface{}) {
	return 0, nil
}
