# Goal-Web 日志系统设计（FR-012）

**日期**: 2026-11-17
**状态**: 设计已确认
**对应规格**: specs/001-golang-fullstack-framework/spec.md FR-012（日志系统，MUST）
**相关背景**: contracts/logger.go 已有 Logger 接口；supports/logs 为 apex/log（已停止维护）全局薄包装，被 8+ 组件引用；当前无任何日志配置

## 已确认的决策

| 决策点 | 结论 | 理由 |
|---|---|---|
| 底层日志库 | **zap** | 性能优先（spec SC-003/SC-008 高性能目标），结构化强 |
| 组件形态 | **新建 `github.com/goal-web/log` 独立组件** + `supports/logs` 改代理 | 遵循组件化惯例；8+ 依赖旧包的组件零改动 |
| v1 范围 | console + file（按天切分）+ 自定义驱动；多通道 + stack 组合；text/json；每通道级别 | YAGNI：不做 syslog/sentry、日志事件、http 中间件（列入后续演进） |

## 一、接口与组件结构

### contracts 增量（不动现有接口，全部向前兼容）

```go
type LogLevel int // Debug / Info / Warn / Error / Fatal

// ChannelLogger = 现有 Logger 能力 + 按级别动态调用
type ChannelLogger interface {
    Logger
    Log(level LogLevel, msg string)
}

// LogManager：容器服务，通道入口
type LogManager interface {
    Channel(name ...string) ChannelLogger // 无参 → default 通道
    Driver(name string, driver LogDriver) // 运行时注册自定义驱动
}

// LogDriver：驱动工厂，从通道配置构建 logger
type LogDriver func(channelName string, config Fields) (Logger, error)
```

### 新组件 `log/` 结构（对齐 cache 组件惯例）

```
log/
├── config.go            # Config / ChannelConfig 结构
├── factory.go           # 按配置构建通道，管理驱动注册表
├── drivers/console.go   # stdout / stderr
├── drivers/file.go      # 单文件 + 按天切分（日期后缀，不引轮转依赖）
├── manager.go           # LogManager 实现（懒构建 + 缓存）
├── zap.go               # zap 适配层：接口映射、级别映射、encoder
├── service_provider.go  # 注册 "log" 服务单例
├── exception.go
└── tests/
```

zap 适配要点：`WithFields` → zap fields；`WithException` → error + previous 链；级别过滤走 `zap.Check()` 在格式化前短路；`Fatal` 保持 zap 语义（写完 `os.Exit(1)`）。

## 二、配置模型

```toml
[log]
default = "stack"

[log.channels.console]
driver = "console"          # console | file | stack | 自定义
level = "debug"             # debug | info | warn | error | fatal
format = "console"          # console | json

[log.channels.file]
driver = "file"
level = "info"
format = "json"
path = "storage/logs/app.log"  # 落盘为 app-20261117.log，按天切分

[log.channels.stack]
driver = "stack"
level = "debug"
channels = ["console", "file"]
```

运行时机制：

- **懒构建 + 缓存**：通道首次访问构建 zap 实例后复用；stack 用 `zapcore.NewTee` 扇出，自身级别过滤在 tee 后统一做
- **零配置兜底**：`[log]` 缺失时自动生成 `console/debug` 单通道（spec 边缘用例：配置缺失优雅降级）
- **自定义驱动**：`manager.Driver("name", fn)` 注册后配置里 `driver = "name"` 即用
- **级别过滤短路**：低于通道级别的日志不产生格式化开销

## 三、goal 骨架集成与兼容代理

### 骨架接入

- `goal/config/log.go`：`[log]` 配置节访问器（对齐 config/cache.go 写法）；`config/log_example.toml` 示例
- `bootstrap/core/application.go`：注册 `log.NewService(...)`
- 业务用法（依赖注入）：

```go
func Api(router contracts.HttpRouter, log contracts.LogManager) {
    router.Get("/order/{id}", func(request contracts.HttpRequest) any {
        log.Channel().WithField("order_id", request.Param("id")).Info("查询订单")
        log.Channel("file").Error("落盘错误日志")
        return ...
    })
}
```

### supports/logs 兼容代理

- 包 API 原样保留（WithField/WithFields/WithError/WithException/Default/WithAny），内部代理到新组件默认通道
- 依赖方向：`supports/logs → log → contracts`，log 不 import supports，无循环依赖
- 引导时机：log 服务提供者启动时 `logs.SetDefault(...)` 安装 LogManager；启动前自动落内置 console/debug 兜底
- `logs.Debug` 开关变量保留，语义不变（false 时 Debug 静默）
- 唯一可感知变化：兜底输出格式从 apex 默认变为 zap console 格式

## 四、错误处理

- 配置错误（未知驱动、非法级别、stack 引用不存在通道）：构建通道返回异常（log/exception.go，对齐 cache 惯例），**fallback console/debug + Warn**，不 panic、应用照常启动
- file 驱动：目录不存在自动 MkdirAll；写失败降级 stderr 输出
- Fatal：写完日志 os.Exit(1)

## 五、测试

- `log/tests/`：级别过滤短路、With 链字段传递、stack 扇出、文件按天切分、console/json 格式断言、自定义驱动注册调用、零配置兜底、配置错误 fallback
- `supports/logs` 回归：现有 apex_test.go 原样跑过
- goal 骨架冒烟：bootstrap 启动 → 配置加载 → 双通道各写一条
- benchmark：zap 通道 vs 原 apex 包装（SC-003 首份数据）

## 六、交付步骤（依赖链从底到顶）

1. `contracts` 增量接口 → 提交发版
2. 新组件 `log`（新仓库、`1.0` 分支、CI + go.sum + 双语 README）→ 发版 v0.5.0
3. `supports` 日志改代理 → 发版
4. `goal` 骨架接入 → **前置：先解决 goal 与 origin/1.0 分叉（force 或 merge，需决策）**
5. spec 状态更新 FR-012 + `doc/digging-deeper/logging.md` 文档

实施方式：直接在主 checkout 各组件仓库的 `1.0` 分支上操作（meta-repo 结构下 worktree 无意义）；推送使用 qbhy 身份（`GIT_SSH_COMMAND="ssh -i ~/.ssh/id_rsa -o IdentitiesOnly=yes"`）。

## 后续演进（明确不在 v1）

http 请求日志中间件、日志事件、远程驱动（syslog/sentry）、双写告警、生产级轮转（lumberjack）
