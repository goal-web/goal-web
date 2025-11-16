## 目标
- 为 `supports` 组件提供现代化的工程与文档体系：独立 .github 工作流、依赖更新、Issue 模板、完善测试、README（中英文）、示例与设计说明，统一 Go 版本到 `>=1.25.0`。

## 目录现状与缺口（概览）
- 现有包：`supports` 根包（BaseFields）、`logs`、`utils`、`class`、`exceptions`、`signal`、`commands`。
- 测试现状：`supports/tests` 覆盖 utils/logs/class/compare/file/rand/error 等；缺 `commands`、`exceptions`、`signal` 专属测试。
- 文档缺失：未见 README 或使用文档。

## .github（组件内独立）
- `supports/.github/workflows/ci.yml`
  - on: push/pull_request
  - matrix: `go: ['1.25.0']`
  - steps: checkout → setup-go (cache) → `go test ./... -race -coverprofile=coverage.out` → upload artifact
- `supports/.github/workflows/lint.yml`
  - on: push/pull_request
  - steps: checkout → setup-go@v5 (`go-version: '1.25.0'`, `GOTOOLCHAIN: go1.25.0`) → golangci-lint-action@v8
  - with: `version: v2.6.2`, `install-mode: goinstall`
  - args：初始启用 `staticcheck, govet, revive, misspell`，避免类型信息不全问题；必要时临时降级为仅 `misspell` 以保障通过
- `supports/.github/dependabot.yml`
  - 每周检查 `gomod` 更新（目录 `/`）
- `supports/.github/ISSUE_TEMPLATE/*`
  - `bug_report.yml`、`feature_request.yml`、`config.yml`

## README 与文档
- `supports/README.md`（英文）与 `README.cn.md`（中文）：
  - 顶部徽标矩阵：Go Reference、Go Report Card、MIT、Stars、Release、Go 版本（>=1.25.0）
  - 快速链接：Docs、Issues、Releases、Examples、中文文档
  - Highlights：BaseFields 可选读取、类型安全、日志封装、异常与命令签名解析、工具集合
  - Compatibility：Go 版本与模块路径
  - Quick Start：
    - BaseFields 与 OptionalGetter 示例（对接 `contracts.FieldsProvider` 的典型用法）
    - logs 使用示例（`logs.Default().WithField().Info()`）
    - utils 示例（`Flatten/ToFields`、`RandStr`、`ExistsPath`）
    - commands 签名解析与参数注入示例（`ParseSignature`、`InjectArguments`）
  - Design：模块职责与协作关系（BaseFields 提升成员、logs 的 Debug 逻辑、commands 的 DSL）
  - Star History：图表链接
  - Contributing / Roadmap / Changelog / Security / FAQ
- `supports/docs/`（可选）：简短设计说明与扩展指南

## 测试用例完善
- 新增：`supports/commands/command_test.go`、`supports/commands/signature_test.go`
  - 覆盖 `{name}`、`{name?}`、`{name=default}`、`{--opt}`、`{--opt=default}` 解析与 InjectArguments 的必填/默认/可选行为
- 新增：`supports/exceptions/exception_test.go`、`handler_test.go`
  - 构造异常、处理函数断言、`service_provider` 注册集成基本冒烟测试
- 新增：`supports/signal/service_provider_test.go`
  - 校验服务提供者的注册/启动行为（可简化为接口调用的冒烟）
- 扩展：`supports/tests/utils_test.go` 添加 `ToFields/Flatten` 边界用例（嵌套 map、混合键类型）
- 扩展：`supports/tests/logs_test.go` 增加 Debug 开关行为断言

## Go 版本与模块
- `supports/go.mod`：设置 `go 1.25.0`，依赖版本审查与更新（不改语义，保持兼容）。
- 确保测试与 lint 在组件内独立执行，无需根仓库全局配置。

## README 徽章（可选增强）
- CI 状态：`supports/.github/workflows/ci.yml` 对应徽章
- 覆盖率（后续接入 Codecov 时启用）

## 里程碑
- M1：.github 工作流与模板、英文 README 骨架与中文副本、commands 测试补齐
- M2：exceptions/signal 测试、utils/logs 扩展用例、Star History 与 FAQ 完善
- M3：设计文档与 Roadmap、贡献指南细化、（可选）Codecov 接入与徽章

## 成功标准
- supports 组件具备独立的 CI/Lint/Dependabot/Issue 模板；README 现代化；关键模块（commands/exceptions/signal）具备基础测试并通过；Go 版本与工作流统一到 >=1.25.0。