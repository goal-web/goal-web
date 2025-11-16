## 范围
- 组件：serialization、bloomfilter

## 工程化配置（每个组件）
- 添加独立 .github 配置：
  - CI：push/PR，Go 1.25.0，`go test ./... -race -coverprofile=coverage.out`，上传覆盖率。
  - Lint：`actions/setup-go@v5`（`go-version: '1.25.0'`、`GOTOOLCHAIN: go1.25.0`），`golangci-lint-action@v8`（`version: v2.6.2`、`install-mode: goinstall`），初始 `--disable-all --enable=misspell`；依赖可解析后再扩展 `staticcheck/govet`。
  - Dependabot：每周 gomod 更新。
  - Issue 模板：`bug_report.yml`、`feature_request.yml`、`config.yml`。
- go.mod：统一 `go 1.25.0`，如需在 monorepo 下解析内部依赖则添加 `replace`。

## 文档（每个组件）
- README.md（英文）与 README.cn.md（中文）：
  - 徽章矩阵（Go Reference、Go Report Card、MIT、Stars、Release、Go 版本）
  - 快速链接：Docs、Issues、Releases、Examples、中文文档
  - Highlights：
    - serialization：JSON/XML/GOB serializers、class serializer、service provider
    - bloomfilter：file/redis drivers、hash helpers、factory
  - Quick Start：最小示例（不依赖外部服务）
  - Star History、Contributing、Roadmap、Changelog、Security、FAQ

## 测试用例（最小补充）
- serialization：已有 `tests/serialize_test.go`；保持现状并确保 CI 收集覆盖率。
- bloomfilter：已有 `tests/bloomfilter_test.go`；保持现状，并保证工作流稳定运行。

## 验证与成功标准
- 两组件在独立目录下具备 CI/Lint/Dependabot/Issue 模板；README 现代化；Go 版本统一为 >=1.25.0；现有测试运行并生成覆盖率。

## 执行顺序
1) 添加 .github/ 工作流与模板
2) 添加双语 README
3) 审查 `go.mod` 版本与必要 replace（仅 monorepo 场景）
4) 运行 CI/Lint 验证（本地不执行，按配置完成文件更新）