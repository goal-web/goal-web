package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// RunPipeline 用于执行批量更新操作
func RunPipeline(excludeComponents []string, includeComponents []string, commitMessage string, version string, customScript string, mode string) {
	// 记录脚本开始时间
	startTime := time.Now()

	if mode == "update" {
		// 1. 遍历所有组件，并确定需要操作的组件列表
		allComponents := getAllComponents(excludeComponents, includeComponents, false)
		fmt.Printf("需要操作的组件: %v\n", allComponents)

		for _, component := range allComponents {
			fmt.Printf("正在更新组件: %s\n", component)
			err := runCommandInComponent(component, "git", "push", "--all")
			if err != nil {
				logError(component, "--all 推送", err)
			}
			err = runCommandInComponent(component, "git", "push", "--tag")
			if err != nil {
				logError(component, "--tag 推送", err)
			}
		}
		return
	}

	// 在检查组件之前，先对所有组件执行 pull 操作
	if mode != "check" {
		allComponentsForPull := getAllComponents(excludeComponents, includeComponents, false)
		fmt.Printf("正在更新所有组件的远程代码: %v\n", allComponentsForPull)
		for _, component := range allComponentsForPull {
			// 获取当前分支名称并拉取远程更改
			branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
			branchCmd.Dir = component
			branchOutput, err := branchCmd.Output()
			if err != nil {
				fmt.Printf("组件 %s 获取当前分支失败: %v，跳过 pull\n", component, err)
				continue
			}
			currentBranch := strings.TrimSpace(string(branchOutput))
			
			err = runCommandInComponent(component, "git", "pull", "--rebase", "origin", currentBranch)
			if err != nil {
				fmt.Printf("组件 %s 拉取远程分支 %s 失败: %v\n", component, currentBranch, err)
			}
		}
	}

	// 1. 遍历所有组件，并确定需要操作的组件列表
	allComponents := getAllComponents(excludeComponents, includeComponents, mode == "submit")
	fmt.Printf("需要操作的组件: %v\n", allComponents)

	// 如果指定了版本号，检查版本号的 tag 是否已存在
	if version != "" {
		checkTags(allComponents, version)
	}

	// 记录提交版本更新情况
	versionUpdates := map[string]string{}

	// 2. 遍历所有组件并检查代码（golangci-lint）
	pwd, _ := os.Getwd()
	for _, component := range allComponents {
		fmt.Printf("正在检查组件: %s\n", component)
		err := runCommandInComponent(component, "golangci-lint", "run", "--disable=typecheck", filepath.Join(pwd, component))
		if err != nil {
			logError(component, "代码检查失败", err)
		}
	}

	// 3. 遍历所有组件并执行测试用例（如果存在）
	for _, component := range allComponents {
		fmt.Printf("正在更新组件 %s 的 go.mod 文件，删除 `github.com/goal-web` 的依赖...\n", component)
		err := updateGoMod(component)
		if err != nil {
			logError(component, "更新 go.mod 失败", err)
			continue
		}

		fmt.Printf("正在测试组件: %s\n", component)
		if hasTests(component) {
			err := runCommandInComponent(component, "go", "test", "./tests")
			if err != nil {
				logError(component, "测试用例执行失败", err)
			}
		} else {
			fmt.Printf("组件 %s 没有测试用例，跳过。\n", component)
		}
	}

	// 4. 如果存在自定义脚本，则遍历所有组件并执行自定义脚本
	if customScript != "" {
		for _, component := range allComponents {
			fmt.Printf("正在组件 %s 中执行自定义脚本: %s\n", component, customScript)
			err := runCommandInComponent(component, "sh", "-c", customScript)
			if err != nil {
				logError(component, "自定义脚本执行失败", err)
			}
		}
	}

	// 如果是 `check` 模式，则到此为止
	if mode == "check" {
		fmt.Println("检查模式已完成。")
		fmt.Printf("所有操作完成，总耗时: %s\n", time.Since(startTime))
		return
	}

	// 5. 遍历所有组件并提交代码，打 tag 并 push
	for _, component := range allComponents {
		newVersion := version
		if newVersion == "" {
			newVersion = getNextVersion(component)
		}

		oldVersion := getCurrentVersion(component)

		fmt.Printf("正在为组件 %s 提交代码并打标签: %s\n", component, newVersion)
		
		// Fetch remote tags to ensure we have the latest tags before creating a new one
		err := runCommandInComponent(component, "git", "fetch", "--tags")
		if err != nil {
			fmt.Printf("组件 %s 获取远程标签失败: %v，将继续执行后续操作\n", component, err)
		}
		
		err = runCommandInComponent(component, "git", "add", ".")
		if err != nil {
			logError(component, "Git 添加变更失败", err)
		}

		err = runCommandInComponent(component, "git", "commit", "-m", commitMessage)
		if err != nil {
			logError(component, "Git 提交失败", err)
		}

		err = runCommandInComponent(component, "git", "tag", newVersion)
		if err != nil {
			fmt.Printf("组件 %s 发生错误: %s，错误内容: %v\n", component, "Git 打标签失败，尝试直接推送所有代码", err)
		}

		// 获取当前分支名称并拉取远程更改
		branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		branchCmd.Dir = component
		branchOutput, err := branchCmd.Output()
		if err != nil {
			fmt.Printf("组件 %s 获取当前分支失败: %v，继续推送...\n", component, err)
		} else {
			currentBranch := strings.TrimSpace(string(branchOutput))
			
			err = runCommandInComponent(component, "git", "pull", "--rebase", "origin", currentBranch)
			if err != nil {
				fmt.Printf("组件 %s 拉取远程分支 %s 失败: %v，继续推送...\n", component, currentBranch, err)
			}
		}

		err = runCommandInComponent(component, "git", "push", "--all")
		if err != nil {
			logError(component, "Git 推送失败", err)
		}

		err = runCommandInComponent(component, "git", "push", "origin", newVersion)
		if err != nil {
			logError(component, "Git 推送标签失败", err)
		}

		versionUpdates[component] = fmt.Sprintf("%s -> %s", oldVersion, newVersion)
		fmt.Printf("组件 %s 已成功更新到版本 %s\n", component, newVersion)
	}

	// 输出版本更新情况
	fmt.Println("所有组件的版本更新情况如下：")
	for component, update := range versionUpdates {
		fmt.Printf("组件 %s: %s\n", component, update)
	}

	fmt.Printf("所有操作完成，总耗时: %s\n", time.Since(startTime))
}

// 判断组件目录中是否存在未提交的更改（使用 `git status` 或 `git diff`）
func hasGitChanges(component string) bool {
	cmd := exec.Command("git", "diff", "--quiet", "HEAD", "--")
	cmd.Dir = component
	err := cmd.Run()
	// 如果 `git diff --quiet` 返回非零状态码，表示有变更
	return err != nil
}

// 获取所有需要操作的组件（根据 exclude 和 include 参数过滤）
func getAllComponents(excludeComponents []string, includeComponents []string, ignoreEmptyChange bool) []string {
	components := []string{}
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && path != "." {
			if hasGoMod(path) {
				component := filepath.Base(path)
				// 过滤排除的组件
				if contains(excludeComponents, component) {
					return nil
				}
				// 如果指定了包含的组件，则只添加指定的组件
				if len(includeComponents) == 0 || contains(includeComponents, component) {
					// 检查组件是否有修改
					if ignoreEmptyChange && !hasGitChanges(component) {
						fmt.Printf("组件 %s 没有任何修改，已过滤掉。\n", component)
						return nil
					}

					components = append(components, component)
				}
			}
		}
		return nil
	})
	return components
}

// 检查组件目录中是否存在 go.mod 文件
func hasGoMod(path string) bool {
	_, err := os.Stat(filepath.Join(path, "go.mod"))
	return !os.IsNotExist(err)
}

// 检查组件中是否存在 tests 目录且文件夹不为空
func hasTests(component string) bool {
	// 构建 tests 目录的路径
	testsDir := filepath.Join(component, "tests")

	// 检查 tests 目录是否存在
	info, err := os.Stat(testsDir)
	if err != nil || !info.IsDir() {
		// 如果目录不存在或路径不是一个目录，则返回 false
		return false
	}

	// 检查目录中是否有文件
	files, err := os.ReadDir(testsDir)
	if err != nil || len(files) == 0 {
		// 如果无法读取目录或目录为空，则返回 false
		return false
	}

	// tests 目录存在且不为空
	return true
}

// 在指定组件目录下执行命令
func runCommandInComponent(component string, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Dir = component
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 设置超时时间为 60 秒
	timer := time.AfterFunc(60 * time.Second, func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	})
	defer timer.Stop()

	return cmd.Run()
}

// 获取组件当前最新的 tag
func getCurrentVersion(component string) string {
	// Fetch remote tags to ensure we have the latest tags
	fetchCmd := exec.Command("git", "fetch", "--tags")
	fetchCmd.Dir = component
	err := fetchCmd.Run()
	if err != nil {
		fmt.Printf("获取 %s 组件的远程标签失败: %v，将继续获取本地标签\n", component, err)
	}
	
	// 获取所有标签
	cmd := exec.Command("git", "tag", "-l")
	cmd.Dir = component
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("获取 %s 组件的标签失败: %v\n", component, err)
		return ""
	}

	// 分割标签列表, 用 \n 分隔并过滤空值
	tags := strings.Split(strings.TrimSpace(string(output)), "\n")
	
	// Filter out empty tags (this handles the case where the output ends with a newline)
	var validTags []string
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			validTags = append(validTags, tag)
		}
	}
	
	if len(validTags) == 0 {
		fmt.Printf("组件 %s 没有找到任何标签\n", component)
		return ""
	}

	// Parse tags into Version structs and sort them
	var versions []*Version
	for _, tag := range validTags {
		version, err := ParseVersion(tag)
		if err != nil {
			fmt.Printf("无法解析版本 %s: %v，跳过\n", tag, err)
			continue
		}
		versions = append(versions, version)
	}

	// Sort versions in descending order (latest first)
	sort.Slice(versions, func(i, j int) bool {
		return versions[j].Less(*versions[i]) // Descending order: larger versions first
	})

	if len(versions) == 0 {
		fmt.Printf("组件 %s 没有有效的语义化版本标签\n", component)
		return ""
	}

	latestTag := versions[0].String()
	fmt.Printf("组件 %s 的最新标签版本号: %s\n", component, latestTag)
	return latestTag
}

// 从 Git 中获取下一个版本号
func getNextVersion(component string) string {
	currentVersion := getCurrentVersion(component)
	fmt.Printf("当前版本：%s\n", currentVersion)
	if currentVersion == "" {
		return "v0.1.0"
	}
	
	// Parse the current version
	version, err := ParseVersion(currentVersion)
	if err != nil {
		fmt.Printf("无法解析当前版本 %s: %v，使用默认版本 v0.1.0\n", currentVersion, err)
		return "v0.1.0"
	}
	
	// Increment the patch version
	version.Patch++
	
	return version.String()
}

// 增加字符串中的数字（用于版本号升级）
func incrementString(s string) (int, error) {
	var number int
	_, err := fmt.Sscanf(s, "%d", &number)
	return number + 1, err
}

// 检查指定组件是否已存在目标版本号的 tag
func checkTags(components []string, version string) {
	for _, component := range components {
		// Fetch remote tags to ensure we have the latest tags
		fetchCmd := exec.Command("git", "fetch", "--tags")
		fetchCmd.Dir = component
		err := fetchCmd.Run()
		if err != nil {
			fmt.Printf("获取 %s 组件的远程标签失败: %v，将继续检查本地标签\n", component, err)
		}
		
		cmd := exec.Command("git", "tag", "-l")
		cmd.Dir = component
		output, err := cmd.Output()
		if err == nil && strings.Contains(string(output), version) {
			fmt.Printf("组件 %s 中的 tag %s 已存在，操作中止。\n", component, version)
			os.Exit(1)
		}
	}
}

// 判断字符串数组中是否包含指定元素
func contains(array []string, item string) bool {
	for _, v := range array {
		if v == item {
			return true
		}
	}
	return false
}

// 记录错误信息并打印
func logError(component, context string, err error) {
	fmt.Printf("组件 %s 发生错误: %s，错误内容: %v\n", component, context, err)
	log.Fatal(err)
}

// Version represents a semantic version
type Version struct {
	Major int
	Minor int
	Patch int
}

// ParseVersion parses a version string like "v0.5.11" into a Version struct
func ParseVersion(versionStr string) (*Version, error) {
	// Remove the 'v' prefix if present
	versionStr = strings.TrimPrefix(versionStr, "v")

	// Use regex to parse the version parts
	re := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+).*$`)
	matches := re.FindStringSubmatch(versionStr)
	if len(matches) < 4 {
		return nil, fmt.Errorf("invalid version format: %s", versionStr)
	}

	var v Version
	_, err := fmt.Sscanf(matches[1], "%d", &v.Major)
	if err != nil {
		return nil, err
	}
	_, err = fmt.Sscanf(matches[2], "%d", &v.Minor)
	if err != nil {
		return nil, err
	}
	_, err = fmt.Sscanf(matches[3], "%d", &v.Patch)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

// Less returns true if this version is less than the other version
func (v Version) Less(other Version) bool {
	if v.Major != other.Major {
		return v.Major < other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor < other.Minor
	}
	return v.Patch < other.Patch
}

// String returns the string representation of the version
func (v Version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func trim(paths ...string) []string {
	var newPaths []string
	for _, path := range paths {
		if path != "" {
			newPaths = append(newPaths, path)
		}
	}
	return newPaths

}

// 更新组件中的 go.mod 文件，删除所有 `github.com/goal-web` 依赖
func updateGoMod(component string) error {
	if component == "goal-cli" || component == "goal" {
		return nil
	}

	cmd := exec.Command("go", "list", "-m", "all")
	cmd.Dir = component
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("无法列出模块依赖: %v", err)
	}

	// 查找 `github.com/goal-web` 相关的依赖
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "github.com/goal-web") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				module := parts[0]
				// 删除 `github.com/goal-web` 依赖项
				err := runCommandInComponent(component, "go", "mod", "edit", "-droprequire", module)
				if err != nil {
					return fmt.Errorf("无法删除依赖 %s: %v", module, err)
				}
			}
		}
	}

	// 整理 go.mod 文件
	//err = runCommandInComponent(component, "go", "mod", "tidy")
	//if err != nil {
	//	return fmt.Errorf("go mod tidy 失败: %v", err)
	//}

	return nil
}

func main() {
	// 定义命令行参数
	exclude := flag.String("e", "goal,goal-cli", "排除的组件列表（用逗号分隔）")
	include := flag.String("i", "", "包含的组件列表（用逗号分隔）")
	message := flag.String("m", "更新组件版本", "Git 提交消息")
	version := flag.String("v", "", "指定的目标版本号")
	script := flag.String("script", "", "自定义脚本路径")
	mode := flag.String("mode", "submit", "模式: check (默认) 或 submit")

	flag.Parse()

	// 解析参数
	excludeComponents := trim(strings.Split(*exclude, ",")...)
	includeComponents := trim(strings.Split(*include, ",")...)

	if *message == "" && *mode == "submit" {
		log.Fatal("请输入提交备注（-m）")
	}

	// 执行批量更新操作
	RunPipeline(excludeComponents, includeComponents, *message, *version, *script, *mode)
}
