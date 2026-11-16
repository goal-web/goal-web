# Goal-Web 日志系统（FR-012）实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 为 goal-web 框架交付 zap 驱动、多通道、配置化的日志组件，并让 8+ 个现存组件无感迁移。

**Architecture:** contracts 增量接口（不破坏现有 Logger）→ 新建 `log` 独立组件（驱动+工厂+服务提供者）→ `supports/logs` 改为代理兼容层 → goal 骨架接入。设计文档：`docs/plans/2026-11-17-logging-system-design.md`。

**Tech Stack:** Go 1.25 / zap v1.x / testify（测试）/ TOML 配置（goal 惯例）

---

## 全局约定（每个任务都适用）

- **仓库布局**: meta-repo。组件源码只在主 checkout `/Users/qbhy/GolandProjects/go/goal-web/<组件>`；Agent Hub worktree 里组件目录是空的，**不要在里面操作**
- **Shell**: 一律用 `git -C <路径>` / `go -C <路径>`，不 cd
- **组件分支**: 全部在 `1.0` 分支开发（doc=main、goal-cli/views=master 除外）
- **推送身份**: `export GIT_SSH_COMMAND="ssh -i $HOME/.ssh/id_rsa -o IdentitiesOnly=yes -o BatchMode=yes"`（xjl-jhs 身份对 goal-web 只读，qbhy 才有写权限）
- **测试命令**: `go -C /Users/qbhy/GolandProjects/go/goal-web/log test ./tests -v`（新组件）；已加入根 go.work 的模块可从仓库根构建
- **提交信息**: 中文 conventional commits + 尾行 `Co-Authored-By: Claude Code <noreply@anthropic.com>`
- **发版**: `git describe --tags --abbrev=0` 取最新 tag，patch+1（对应 update.sh 的 increment_tag 惯例），`git tag && git push origin <分支> --tags`
- **双语注释**: 参照 cache 组件，公开符号用 `// 英文\n// 中文` 成对注释

## 关键参照文件（动手前先读）

| 文件 | 用途 |
|---|---|
| `contracts/logger.go` | 现有 Logger 接口（39 行，不改） |
| `contracts/basic.go` | `Fields` 类型定义 |
| `cache/service_provider.go` | 服务提供者写法样板 |
| `cache/config.go` | 强类型 Config 样板（`Stores map[string]contracts.Fields`） |
| `supports/logs/{logger,apex}.go` | 被代理的现有 API |
| `goal/config/cache.go` | 骨架配置访问器样板（init() + configs map） |

---

### Task 1: contracts 增量日志接口

**Files:**
- Modify: `contracts/logger.go`（文件末尾追加，不动现有 Logger）

**Step 1: 追加接口定义**

在 `contracts/logger.go` 末尾追加：

```go
// LogLevel represents the severity of a log record.
// LogLevel 表示日志记录的严重级别。
type LogLevel int

const (
	// DebugLevel is for verbose diagnostic output.
	// DebugLevel 用于详细的诊断输出。
	DebugLevel LogLevel = iota
	// InfoLevel is for general operational messages.
	// InfoLevel 用于常规运行信息。
	InfoLevel
	// WarnLevel is for abnormal but recoverable conditions.
	// WarnLevel 用于异常但可恢复的情况。
	WarnLevel
	// ErrorLevel is for errors that should be addressed.
	// ErrorLevel 用于需要处理的错误。
	ErrorLevel
	// FatalLevel is for unrecoverable errors, typically followed by process exit.
	// FatalLevel 用于不可恢复的错误，通常随后进程退出。
	FatalLevel
)

// ChannelLogger extends Logger with level-based dynamic logging.
// ChannelLogger 在 Logger 基础上支持按级别动态记录。
type ChannelLogger interface {
	Logger
	// Log adds a log record at the given level.
	// Log 按给定级别添加日志记录。
	Log(level LogLevel, msg string)
}

// LogManager resolves named log channels and registers custom drivers.
// LogManager 按名称解析日志通道，并注册自定义驱动。
type LogManager interface {
	// Channel returns the logger for the given channel; no name means the default channel.
	// Channel 返回给定通道的 logger；不传名称时返回默认通道。
	Channel(name ...string) ChannelLogger
	// Driver registers a custom log driver under the given name.
	// Driver 以给定名称注册自定义日志驱动。
	Driver(name string, driver LogDriver)
}

// LogDriver builds a Logger from a channel's configuration.
// LogDriver 按通道配置构建 Logger。
type LogDriver func(channelName string, config Fields) (Logger, error)
```

**Step 2: 验证编译**

Run: `go -C /Users/qbhy/GolandProjects/go/goal-web build ./contracts/...`
Expected: 无输出（成功）

**Step 3: 全 workspace 回归（接口是增量，不能破坏任何组件）**

Run: `go -C /Users/qbhy/GolandProjects/go/goal-web build ./...`
Expected: 成功

**Step 4: Commit**

```bash
git -C /Users/qbhy/GolandProjects/go/goal-web/contracts add logger.go
git -C /Users/qbhy/GolandProjects/go/goal-web/contracts commit -m "feat: 新增 ChannelLogger/LogManager/LogDriver 日志接口" -m "Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

**Step 5: 发版 contracts（当前 v0.5.4 → v0.5.5）**

```bash
export GIT_SSH_COMMAND="ssh -i $HOME/.ssh/id_rsa -o IdentitiesOnly=yes -o BatchMode=yes"
git -C /Users/qbhy/GolandProjects/go/goal-web/contracts tag v0.5.5
git -C /Users/qbhy/GolandProjects/go/goal-web/contracts push origin 1.0 --tags
```

---

### Task 2: log 组件脚手架

**Files:**
- Create: 远程仓库 `goal-web/log`（GitHub，公开）
- Create: `/Users/qbhy/GolandProjects/go/goal-web/log/go.mod`
- Create: `log/.gitignore`、`log/.github/workflows/{ci.yml,int.yml→lint.yml}`、`log/README.md`、`log/README.cn.md`
- Modify: 根 `go.work`（use 列表加 `./log`）

**Step 1: 创建 GitHub 仓库并 init 本地**

```bash
export GIT_SSH_COMMAND="ssh -i $HOME/.ssh/id_rsa -o IdentitiesOnly=yes -o BatchMode=yes"
gh repo create goal-web/log --public --description "Goal-Web logging component powered by zap"
mkdir -p /Users/qbhy/GolandProjects/go/goal-web/log && cd 不允许 → 直接在目标目录 git init:
git -C /Users/qbhy/GolandProjects/go/goal-web/log init -b 1.0
git -C /Users/qbhy/GolandProjects/go/goal-web/log remote add origin git@github.com:goal-web/log.git
```

**Step 2: go.mod**

```
module github.com/goal-web/log

go 1.25.0

require (
	github.com/goal-web/contracts v0.5.5
	github.com/stretchr/testify v1.9.0
	go.uber.org/zap v1.27.0
)
```

然后 `go -C .../log mod tidy`（拉取补全 go.sum；apex/log 是间接依赖随 supports 引入不需要）。
注意：contracts v0.5.5 是 Task 1 刚发的 tag。

**Step 3: 根 go.work 加 `./log`**

在 `/Users/qbhy/GolandProjects/go/goal-web/go.work` 的 use 块中按字母序插入 `./log`。验证：`go -C /Users/qbhy/GolandProjects/go/goal-web build ./log/...`

**Step 4: 工程配置文件**

`.gitignore`:

```
.idea
.DS_Store
```

`.github/workflows/ci.yml` 与 `lint.yml`：从 `container/.github/workflows/` 复制模板，但**删除 on.push.branches 与 on.pull_request.branches 过滤**（昨天修过的坑：触发器不能写死分支）。

`README.md` / `README.cn.md`：暂写一句话占位（Task 8 补全）。

**Step 5: Commit + push（先推空骨架，确认 CI 接通）**

```bash
git -C .../log add -A
git -C .../log commit -m "chore: 初始化 log 组件脚手架" -m "Co-Authored-By: Claude Code <noreply@anthropic.com>"
git -C .../log push -u origin 1.0
```

同时提交根 go.work（外层仓库）：

```bash
git -C /Users/qbhy/GolandProjects/go/goal-web add go.work
git -C /Users/qbhy/GolandProjects/go/goal-web commit -m "chore: go.work 加入 log 组件" -m "Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

### Task 3: zap 适配层（TDD）

**Files:**
- Create: `log/zap.go`（zapLogger：contracts.Logger/ChannelLogger 的 zap 实现 + encoder/helper）
- Test: `log/tests/zap_test.go`

**Step 1: 写失败测试** `log/tests/zap_test.go`

```go
package tests

import (
	"bytes"
	"testing"

	"github.com/goal-web/contracts"
	"github.com/goal-web/log"
	"github.com/stretchr/testify/assert"
)

// newBufferChannel 通过自定义驱动拿到写入 buffer 的 logger，避免依赖真实 stdout
func newBufferChannel(t *testing.T, format string) (*bytes.Buffer, contracts.ChannelLogger) {
	buf := &bytes.Buffer{}
	mgr := log.NewFactory(log.Config{
		Default: "test",
		Channels: map[string]contracts.Fields{
			"test": {
				"driver": "buffer",
				"level":  "debug",
				"format": format,
				"writer": buf,
			},
		},
	})
	mgr.Driver("buffer", func(name string, cfg contracts.Fields) (contracts.Logger, error) {
		return log.NewBufferLogger(cfg) // Task 3 提供的测试辅助构造器（导出，仅供测试场景注入 writer）
	})
	return buf, mgr.Channel()
}
```

再加三个用例（同一文件）：

```go
func TestLevelMethods(t *testing.T) {
	buf, logger := newBufferChannel(t, "console")
	logger.Debug("d"); logger.Info("i"); logger.Warn("w"); logger.Error("e")
	out := buf.String()
	for _, s := range []string{"d", "i", "w", "e"} {
		assert.Contains(t, out, s)
	}
	assert.NotContains(t, out, "FATAL")
}

func TestWithChain(t *testing.T) {
	buf, logger := newBufferChannel(t, "json")
	logger.WithField("k1", "v1").WithFields(map[string]any{"k2": 2}).WithError(assert.AnError).Info("chained")
	out := buf.String()
	assert.Contains(t, out, `"k1":"v1"`)
	assert.Contains(t, out, `"k2":2`)
	assert.Contains(t, out, "chained")
}

func TestLogByLevel(t *testing.T) {
	buf, logger := newBufferChannel(t, "console")
	logger.Log(contracts.WarnLevel, "by-level")
	assert.Contains(t, buf.String(), "by-level")
}

func TestLevelFilter(t *testing.T) {
	mgr := log.NewFactory(log.Config{
		Default: "test",
		Channels: map[string]contracts.Fields{
			"test": {"driver": "buffer", "level": "warn", "format": "console", "writer": &bytes.Buffer{}},
		},
	})
	mgr.Driver("buffer", func(name string, cfg contracts.Fields) (contracts.Logger, error) {
		return log.NewBufferLogger(cfg)
	})
	logger := mgr.Channel()
	logger.Info("dropped"); logger.Warn("kept")
	// 通过 buffer writer 无法直接读（在工厂里闭包了），改用第二个用例方式断言——见下
}
```

> 注意 TestLevelFilter 的断言方式：把 buffer 提到用例作用域再断言 `Contains "kept"` `NotContains "dropped"`。写代码时直接内联，不照抄上面的示意。

**Step 2: 跑测试确认失败**

Run: `go -C .../log test ./tests -run TestLevel -v`
Expected: FAIL（`log.NewFactory` / `log.NewBufferLogger` 未定义）

**Step 3: 实现 `log/zap.go`**

要点（完整实现）：
- `type zapLogger struct { l *zap.Logger }`；`WithFields/WithField/WithError/WithException` 基于 `l.With(...)` 返回新实例（zap 不可变语义）
- 级别映射 `contracts.LogLevel → zapcore.Level`
- `Log(level, msg)` 按 level 分派到 `l.Debug/Info/Warn/Error/Fatal`
- 导出两个构造入口：
  - `NewBufferLogger(cfg contracts.Fields) (contracts.Logger, error)`：读 `writer io.Writer`、`level`、`format`，组装 encoder+core——**同时是自定义驱动/测试的标准构造器**
  - `newChannelLogger(writer zapcore.WriteSyncer, levelStr, format string) *zapLogger`（内部）
- encoder：`console` → `zapcore.NewConsoleEncoder`（时间 `2006-01-02 15:04:05`、大写级别、message 优先）；`json` → `zapcore.NewJSONEncoder`
- `Sync() error` 语义：不进接口，文件驱动内部处理

**Step 4: 跑测试确认通过**

Run: `go -C .../log test ./tests -v`（此时 4 个用例全绿）

**Step 5: Commit** `feat: zap 适配层与通道 logger 实现`

---

### Task 4: config 与零配置兜底（TDD）

**Files:**
- Create: `log/config.go`
- Test: `log/tests/config_test.go`

**Step 1: 失败测试**

```go
func TestZeroConfigFallback(t *testing.T) {
	mgr := log.NewFactory(log.Config{}) // 全空
	logger := mgr.Channel()             // 不 panic，返回 console/debug
	assert.NotNil(t, logger)
}

func TestParseConfig(t *testing.T) {
	cfg := log.ParseConfig(contracts.Fields{
		"default": "file",
		"channels": contracts.Fields{
			"file": contracts.Fields{"driver": "file", "level": "info"},
		},
	})
	assert.Equal(t, "file", cfg.Default)
	assert.Len(t, cfg.Channels, 1)
}
```

**Step 2: 确认失败**（ParseConfig 不存在）

**Step 3: 实现 `log/config.go`**

```go
// Config holds logging configuration.
// Config 保存日志配置。
type Config struct {
	Default  string                        // default channel name 默认通道名
	Channels map[string]contracts.Fields   // channel configs 通道配置
}

// ParseConfig converts raw config fields into Config with fallback defaults.
// ParseConfig 将原始配置字段转换为 Config，并应用兜底默认值。
func ParseConfig(raw any) Config
```

规则：raw 非 Fields / 为空 → 返回零配置；`ParseConfig` 不注入兜底，兜底在 factory（`NewFactory` 收到 Default=="" 且 Channels 空 → 自动生成 `console/debug` 并设 default）。`default` 指向不存在的通道 → factory 侧 fallback + Warn。

**Step 4: 通过 → Step 5: Commit** `feat: 日志配置结构与零配置兜底`

---

### Task 5: console 与 file 驱动（TDD）

**Files:**
- Create: `log/drivers/console.go`、`log/drivers/file.go`
- Test: `log/tests/file_test.go`

**Step 1: 失败测试** `log/tests/file_test.go`

```go
func TestFileDriverDailySplit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	now := time.Now()
	day1 := now
	day2 := now.AddDate(0, 0, 1)

	w := log.NewDailyFileWriter(path, func() time.Time { return day1 })
	w.Write([]byte("day-one\n"))

	w2 := log.NewDailyFileWriter(path, func() time.Time { return day2 })
	w2.Write([]byte("day-two\n"))

	files, _ := filepath.Glob(filepath.Join(dir, "app-*.log"))
	assert.Len(t, files, 2) // app-20261117.log / app-20261118.log（按注入时间）
}
```

（`NewDailyFileWriter` 导出以便测试注入时钟；驱动内部调用它。）

**Step 2: 确认失败 → Step 3: 实现**

`drivers/file.go`：
- `type DailyFileWriter struct { mu, path, day, file, nowFn }`，`NewDailyFileWriter(path string, now func() time.Time) *DailyFileWriter`
- `Write`：加锁；`day != now().Format("20060102")` 时关闭旧文件，打开 `TrimSuffix(path, ext) + "-" + day + ext`（`os.MkdirAll(dir, 0755)`）；写失败时向 stderr 兜底输出一条
- 实现 `zapcore.WriteSyncer`（Write + Sync）
- `drivers/console.go`：`NewConsoleWriter(cfg) zapcore.WriteSyncer`——`output` 字段 `stdout`(默认)/`stderr`
- factory 内置驱动注册：`"console"`（用 console writer + level + format 组 core）、`"file"`（用 DailyFileWriter，now 用 time.Now）

**Step 4: 通过 → Step 5: Commit** `feat: console/file 驱动与按天切分写入器`

---

### Task 6: stack 通道与 Manager（TDD）

**Files:**
- Create: `log/manager.go`
- Modify: `log/factory.go`（make 流程：stack 特殊处理 + 懒构建缓存 + fallback）
- Test: `log/tests/manager_test.go`

**Step 1: 失败测试**

```go
func TestChannelCache(t *testing.T) {
	mgr := log.NewFactory(validConfig())
	assert.Same(t, mgr.Channel("c1"), mgr.Channel("c1")) // 同实例
}

func TestStackFanout(t *testing.T) {
	buf1, buf2 := &bytes.Buffer{}, &bytes.Buffer{}
	mgr := log.NewFactory(log.Config{
		Default: "stack",
		Channels: map[string]contracts.Fields{
			"a": {"driver": "buffer", "writer": buf1, "level": "debug", "format": "console"},
			"b": {"driver": "buffer", "writer": buf2, "level": "debug", "format": "console"},
			"stack": {"driver": "stack", "channels": []string{"a", "b"}, "level": "debug"},
		},
	})
	mgr.Channel().Info("fan")
	assert.Contains(t, buf1.String(), "fan")
	assert.Contains(t, buf2.String(), "fan")
}

func TestCustomDriver(t *testing.T) {
	mgr := log.NewFactory(log.Config{
		Default:  "mem",
		Channels: map[string]contracts.Fields{"mem": {"driver": "mem"}},
	})
	mgr.Driver("mem", func(name string, cfg contracts.Fields) (contracts.Logger, error) {
		return &nopLogger{}, nil
	})
	assert.NotNil(t, mgr.Channel())
}

func TestUnknownChannelFallback(t *testing.T) {
	mgr := log.NewFactory(log.Config{Default: "nope"}) // default 指向不存在通道
	assert.NotNil(t, mgr.Channel())                    // 兜底 console/debug，不 panic
}
```

**Step 2: 确认失败 → Step 3: 实现 manager.go**

- `type factory struct { config Config; mu sync.Mutex; channels map[string]contracts.ChannelLogger; drivers map[string]contracts.LogDriver }`
- `NewFactory(Config) LogManager`（空配置 → 注入 console/debug 兜底）
- `Channel(name ...string)`：默认名解析 → 缓存命中 → `make` → 失败 fallback+Warn → 缓存
- `make`：`driver == "stack"` → 递归构建子通道 → `zapcore.NewTee` 组合；否则查 drivers 表（内置 console/file + 注册的自定义）
- 返回类型守卫：driver 产物实现 ChannelLogger 直接用，否则包一层 levelRouter（把 Log(level,msg) 分派到 Logger 方法）——兼容第三方返回值
- `Driver(name, driver)`：注册到 drivers 表（不影响已缓存通道）

**Step 4: 通过 → Step 5: Commit** `feat: LogManager 懒构建缓存、stack 组合通道与自定义驱动`

---

### Task 7: service_provider、exception、双语 README

**Files:**
- Create: `log/service_provider.go`、`log/exception.go`、`log/examples/`（示例代码）
- Modify: `log/README.md`、`log/README.cn.md`

**Step 1: service_provider.go（严格对照 cache/service_provider.go 的结构）**

```go
func (provider serviceProvider) Register(container contracts.Application) {
	container.Singleton("log", func(config contracts.Config) contracts.LogManager {
		return NewFactory(ParseConfig(config.Get("log")))
	})
}
```

Start/Stop 返回空实现。注意：`logs.SetDefault` 的接线**不在组件内做**（会与 supports 形成循环依赖），由 goal 骨架在服务注册后接线（Task 12）——对设计文档的一处修正，理由已记录。

**Step 2: exception.go** —— `ErrDriverNotFound` / `ErrChannelNotFound` 哨兵错误 + `ConfigException`（对齐 cache/exception.go 惯例，先读该文件再对齐命名风格）。

**Step 3: README 双语**：安装（go get）、TOML 配置示例（设计文档第 2 节）、用法（骨架用例）、自定义驱动示例、兜底行为说明。

**Step 4: Commit** `docs+feat: 服务提供者、异常定义与双语 README`

---

### Task 8: benchmark（SC-003 首份数据）

**Files:**
- Test: `log/tests/benchmark_test.go`

**Step 1:** `BenchmarkZapLogger`（buffer writer，Info 级别 1 字段）与 `BenchmarkApexLogger`（`supports/logs` 的 `WithField(...).Info`，引 apex/log）各跑一组。

**Step 2:** Run: `go -C .../log test ./tests -bench . -benchmem`
Expected: 输出两组 ns/op 数据，记录到 README 性能小节。

**Step 3: Commit** `test: zap 与 apex 日志基准对比`

---

### Task 9: log 组件发版

```bash
export GIT_SSH_COMMAND="ssh -i $HOME/.ssh/id_rsa -o IdentitiesOnly=yes -o BatchMode=yes"
git -C .../log tag v0.5.0
git -C .../log push origin 1.0 --tags
```

验证：`go list -m -versions github.com/goal-web/log`（稍后能看到 v0.5.0；pkg.go.dev 索引有延迟，不阻塞）。

---

### Task 10: supports/logs 改代理 + 回归

**Files:**
- Modify: `supports/logs/logger.go`、`supports/logs/apex.go`（改为代理实现）
- Modify: `supports/go.mod`（依赖 `github.com/goal-web/log v0.5.0`；apex/log 保留为测试用或移除）
- Test: 回归 `supports/logs/apex_test.go`

**Step 1: 先跑现有测试记录基线**

Run: `go -C .../supports test ./logs/... -v`
Expected: PASS（记录输出格式基线）

**Step 2: 重写为代理**

`logger.go` 保留全部导出 API（`WithField/WithFields/WithError/WithException/Default/WithAny/Debug var`），新增：

```go
// SetDefault installs a provider returning the configured default logger.
// SetDefault 安装返回已配置默认 logger 的提供者（由应用引导时调用）。
func SetDefault(provider func() contracts.Logger)

func current() contracts.Logger {
	if defaultProvider != nil {
		if l := defaultProvider(); l != nil {
			return l
		}
	}
	return log.NewFactory(Config{}式兜底).Channel()  // 内置 console/debug 兜底
}
```

`apex.go` 的 ApexLogger 实现删除或保留为 fallback？**删除**（兜底由新组件提供），`Debug` 变量语义保留：`Debug()` 方法先判 `Debug` 再委托。去掉 apex/log 依赖（若 apex_test.go 依赖 apex 断言，则改写该测试用 testify + 自定义 SetDefault 注入断言）。

**Step 3: 全部测试通过**（改写后的用例 + 其他依赖 supports 的组件 workspace 构建无恙：`go -C 根 build ./...`）

**Step 4: Commit + 发版**（supports 当前 v0.5.7 → v0.5.8，`push origin 1.0 --tags`）

---

### Task 11: goal 骨架接入

**前置（需用户决策后执行）**: goal 与 origin/1.0 分叉（本地领先 2 / 落后 1）。先解决再动手。

**Files:**
- Create: `goal/config/log.go`、`goal/config/log_example.toml`
- Modify: `goal/bootstrap/core/application.go`（服务列表 + logs.SetDefault 接线）
- Modify: `goal/go.mod`（`go get github.com/goal-web/log@v0.5.0` + 升级 contracts/supports）
- Modify: `goal/go.work`（本地 gitignored 文件，加 `../log`）

**Step 1: `config/log.go`**（对照 goal/config/cache.go 的 init() 模式）：

```go
func init() {
	configs["log"] = func(env contracts.Env) any {
		return log.Config{
			Default: utils.StringOr(env.GetString("log.default"), "stack"),
			Channels: map[string]contracts.Fields{
				"console": {"driver": "console", "level": env.GetString("log.level"), "format": "console"},
				"file":    {"driver": "file", "level": "info", "format": "json", "path": env.GetString("log.path")},
				"stack":   {"driver": "stack", "level": "debug", "channels": []string{"console", "file"}},
			},
		}
	}
}
```

**Step 2: application.go**：`RegisterServices(logs.NewService(), ...)` 加入列表；注册后接线：

```go
logs.SetDefault(func() contracts.Logger {
	return contracts.LogManager(app.Get("log")).Channel()
})
```

（`app.Get` 的具体取用方式以 application.go 现有代码为准，先读再写。）

**Step 3: 骨架冒烟**：`go run main.go run`（或 `go -C .../goal build ./...` 至少编译）→ 确认启动日志走 zap console 格式、storage/logs 下生成按天文件。

**Step 4: Commit + 发版 goal**（此步也是分叉处理后的第一次推送，方式取决于 Step 0 的决策）

---

### Task 12: 文档与收尾

**Files:**
- Create: `doc/guide/...`（VitePress 新结构）或 `doc/digging-deeper/logging.md`（旧结构）——以 doc 仓库当前实际结构为准
- Modify: `docs/plans/2026-11-17-logging-system-design.md`（记录 Task 7 的 SetDefault 接线位置修正）

**Step 1:** 写 `logging.md`：配置、通道、自定义驱动、兜底行为（内容可从 log README 扩写）。
**Step 2:** 外层仓库全量验证：`go -C 根 build ./...` + 组件状态 sweep（确认无脏遗留）。
**Step 3:** 更新记忆（FR-012 完成、log 组件 v0.5.0、supports v0.5.8）。
**Step 4:** doc 仓库 commit + push（main 分支）。

---

## 完成定义（Definition of Done）

1. `contracts` v0.5.5 / `log` v0.5.0 / `supports` v0.5.8 / goal 已接入，全部已推送
2. `go -C 根 build ./...` 通过；log 组件 8 组测试 + benchmark 全绿
3. `supports/logs` 8+ 依赖组件零改动（workspace 构建即证明）
4. goal 启动后：console 出 zap 格式日志、`storage/logs/app-YYYYMMDD.log` 按天生成
5. `docs/plans/` 与 doc 站文档就位
