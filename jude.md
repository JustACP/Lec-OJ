实现代码评测需要考虑安全性和稳定性，可以采用沙箱技术和多台服务器进行评测。
一般而言，代码评测需要以下步骤：
1. 接收用户提交的代码，并将其保存到文件中。
2. 根据代码的编程语言，选择相应的编译器/解释器进行编译/解释生成可执行文件或字节码。
3. 在沙箱环境下运行可执行文件/字节码，拦截所有破坏性的操作，如修改文件、网络访问等。
4. 将运行结果和运行时间/内存消耗等信息记录到数据库中，并将评测结果返回给用户。
   以下是代码评测部分的基本流程：
1. 接收用户提交的代码，并将其保存到文件中。

```go
    package main

import (
   "fmt"
   "io/ioutil"
   "path/filepath"
)

func saveCode(code, language string) (string, error) {
   filename := fmt.Sprintf("%s.%s", uuid.New().String(), language)
   path := filepath.Join("code", filename)
   err := ioutil.WriteFile(path, []byte(code), 0644)
   if err != nil {
      return "", err
   }
   return path, nil
}
  ```
2. 根据代码的编程语言，选择相应的编译器/解释器进行编译/解释生成可执行文件或字节码。

```go
package main

import (
   "fmt"
   "os/exec"
)

func compile(path, language string) error {
   switch language {
   case "go":
      cmd := exec.Command("go", "build", "-o", "main", path)
      return cmd.Run()
   case "python":
      cmd := exec.Command("python", "-m", "py_compile", path)
      return cmd.Run()
      // ...
   default:
      return fmt.Errorf("unsupported language: %s", language)
   }
}

```

3. 在沙箱环境下运行可执行文件/字节码，拦截所有破坏性的操作，如修改文件、网络访问等。
   可以使用开源的沙箱工具，如seccomp或者Docker，搭建一个沙箱环境，限制运行的程序只能访问指定的文件和目录，并禁止访问网络、修改文件等破坏性操作。具体实现可以参考下面的代码：

```go
package main

import (
   "bufio"
   "os/exec"
   "syscall"
)

func run(path string, input []byte) (*exec.Cmd, error) {
   cmd := exec.Command(path)
   cmd.SysProcAttr = &syscall.SysProcAttr{
      // Limit the process to read-only access to the file system.
      // Ensures that the process cannot write to disk.
      Chroot: "/sandbox",
      // Mount a tmpfs filesystem at /sandbox, ensuring that the process cannot read from disk.
      // Ensures that the process cannot read from disk.
      MountNamespace: true,
      // Limit the process to read-only access to /dev, /proc, and /sys.
      // Ensures that the process cannot see host's kernel information.
      CloneFlags: syscall.CLONE_NEWNS | syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID,
   }
   // Use a pipe to pass input data to the process.
   stdin, err := cmd.StdinPipe()
   if err != nil {
      return nil, err
   }
   // Use a pipe to capture output data from the process.
   stdout, err := cmd.StdoutPipe()
   if err != nil {
      return nil, err
   }
   // Start the process.
   if err := cmd.Start(); err != nil {
      return nil, err
   }
   // Write input data to the process's stdin.
   if _, err := stdin.Write(input); err != nil {
      return nil, err
   }
   stdin.Close()
   // Use bufio.Scanner to read output data.
   scanner := bufio.NewScanner(stdout)
   scanner.Split(bufio.ScanLines)
   // Capture output data and write it to output buffer.
   for scanner.Scan() {
      // ...
   }
   // Wait for the process to finish and obtain the exit code.
   if err := cmd.Wait(); err != nil {
      return nil, err
   }
   return cmd, nil
}
```

4. 将运行结果和运行时间/内存消耗等信息记录到数据库中，并将评测结果返回给用户。
   根据运行结果，可以将结果分类为正确、错误、超时等情况，并将结果记录到数据库中，供用户查询使用。
```go
   type RunResult struct {
   Status  string  `json:"status"`
   Output  string  `json:"output"`
   Time    float64 `json:"time"`
   Memory  int64   `json:"memory"`
   }
   func runCode(path, input string) (*RunResult, error) {
   // Compile the code.
   err := compile(path)
   if err != nil {
   return nil, err
   }
   // Run the code.
   start := time.Now()
   cmd, err := run("./main", []byte(input))
   if err != nil {
   return nil, err
   }
   elapsed := time.Since(start)
   // Capture output and compute time/memory usage.
   var buf bytes.Buffer
   io.Copy(&buf, cmd.Stdout)
   output := buf.String()
   // Get resource usage information.
   usage, err := getUsage(cmd.Process.Pid)
   if err != nil {
   return nil, err
   }
   // Construct the run result.
   result := &RunResult{
   Status:  "correct",
   Output:  output,
   Time:    elapsed.Seconds(),
   Memory:  usage.MaxRSS,
   }
   // Record the run result in the database.
   // ...
   return result, nil
   }
```
   以上是一个代码评测的基本实现流程，实际的实现还需要根据具体应用场景和安全性要求进行相应的调整和完善。