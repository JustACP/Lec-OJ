package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"oj/Entity"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
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
	pwd := exec.Command("pwd")
	stdout, _ := pwd.StdoutPipe()
	_ = pwd.Start()
	locatae, _ := io.ReadAll(stdout)
	locate := string(locatae)
	locate, _ = strings.CutSuffix(locate, "\n")
	locate += "/"
	switch language {
	case "go":
		command := []string{locate + path[0:len(path)-3]}
		cmd := exec.Command("go", "build", path)
		err := cmd.Run()
		os.Chown(path[0:len(path)-4], 0, 0)
		_ = os.Chmod(path[0:len(path)-3], 0755)
		return command, err
	case "c++":
		cmd := exec.Command("g++", "-o", path[0:len(path)-4], "-Wall", "-O2", path)
		err := cmd.Run()
		os.Chown(path[0:len(path)-5], 0, 0)
		_ = os.Chmod(path[0:len(path)-5], 0755)
		command := []string{locate + path[0:len(path)-4]}
		return command, err
	case "python":
		command := []string{"python3", path}
		return command, nil
	}
	return nil, fmt.Errorf("语言类型不匹配")
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

func execCode(input string, command []string, wg *sync.WaitGroup, ch *chan RunResult, num int) {
	name := command[0]
	cmd := exec.Command(name, command[1:]...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		*ch <- RunResult{Exception: err.Error()}
		return
	}
	// 处理获取标准错误输出失败的情况
	stderr, err := cmd.StderrPipe()
	if err != nil {
		*ch <- RunResult{Exception: err.Error()}
		return
	}
	// 使用管道捕获进程的输出数据。
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		*ch <- RunResult{Exception: err.Error()}
		return
	}
	// 启动流程。
	start := time.Now()
	if err := cmd.Start(); err != nil {
		*ch <- RunResult{Exception: err.Error()}
		return
	}

	memory, _ := getUsage(cmd.Process.Pid)
	// 将输入数据写入进程的stdin
	_, _ = io.WriteString(stdin, input)
	_ = stdin.Close()
	var buf bytes.Buffer
	// 处理 获取标准错误输出失败的情况
	if _, err := io.Copy(&buf, stderr); err != nil {
		*ch <- RunResult{Exception: err.Error()}
		return
	}
	elapsed := time.Since(start)
	output, err := io.ReadAll(stdout)
	// 处理执行代码失败
	if err := cmd.Wait(); err != nil {
		// 处理异常信息
		if buf.Len() > 0 {
			*ch <- RunResult{Exception: err.Error()}
			return
		}
	}
	result := RunResult{
		Output: string(output),
		Time:   elapsed.Seconds(),
		Memory: memory,
	}
	result.Number = num
	*ch <- result
	return
}
func runCode(testPoints []string, command []string) (int, interface{}) {
	log.Println("run code")
	var ans Entity.Response
	var wg sync.WaitGroup
	resList := make([]RunResult, 0)
	ch := make(chan RunResult, len(testPoints))
	for num, val := range testPoints {
		wg.Add(1)
		go run(val, num, &wg, &ch, command)
	}
	wg.Wait()
	close(ch)
	for i := range ch {
		resList = append(resList, i)
		// log.Println(resList)
	}
	ans.Data = resList
	return 200, ans
}
func run(val string, num int, wg *sync.WaitGroup, ch *chan RunResult, command []string) {
	defer wg.Done()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	done := make(chan RunResult, 2)
	go execCode(val, command, wg, &done, num)
	go MemoryMonitor(&mem, &done)
	select {
	case <-ctx.Done():
		*ch <- RunResult{Exception: "TLE", Number: num}
		log.Println(strconv.Itoa(num) + "TLE")
		return
	case faq := <-done:
		faq.Number = num
		*ch <- faq
		log.Println(strconv.Itoa(num) + faq.Output)
		return
	}
}

func MemoryMonitor(r *runtime.MemStats, done *chan RunResult) {
	be := time.Now()
	for true {
		if time.Since(be).Seconds() < 1.00 {
			if r.Alloc >= 5120000 {
				*done <- RunResult{Exception: "mle"}
				return
			}
		} else {
			return
		}
	}

}
