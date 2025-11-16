## 目标
- 新增 GitHub Actions 工作流与相关 GitHub 配置文件，实现对 `config` 组件的构建、测试、覆盖率与静态检查，并建立依赖更新与基本模板，形成现代化工程基础设施。

## 工作流文件
- `.github/workflows/config-ci.yml`
  - 触发：`push` 与 `pull_request`，限定路径 `config/**`
  - 环境：`ubuntu-latest`，`go: [1.21, 1.22]` 矩阵
  - 步骤：`actions/checkout` → `actions/setup-go`（cache 开启） → `working-directory: config` → `go test ./... -race -coverprofile=coverage.out` → 上传覆盖率 artifact
- `.github/workflows/config-lint.yml`
  - 触发：`push` 与 `pull_request`，限定路径 `config/**`
  - 步骤：`actions/checkout` → `actions/setup-go` → 使用 `golangci/golangci-lint-action@v4` 在 `config` 目录运行静态检查
- 可选：`.github/workflows/monorepo-ci.yml`
  - 针对整个单仓库模块集运行 `go test`；初始版本可仅覆盖 `config`，后续按需扩展模块矩阵
- 可选：`.github/workflows/codeql.yml`
  - 安全扫描（Go 语言）
- 可选：`.github/workflows/release.yml`
  - 如需要二进制/库发布，结合 `goreleaser`；当前为库，可暂不启用

## 依赖与自动更新
- `.github/dependabot.yml`
  - 每周检查 Go 模块更新，覆盖根目录与 `config/` 子模块
  - 按路径配置多个包管理器条目，以 monorepo 兼容

## Issue 模板（可选）
- `.github/ISSUE_TEMPLATE/bug_report.yml`
- `.github/ISSUE_TEMPLATE/feature_request.yml`
- `.github/ISSUE_TEMPLATE/config.yml`（禁用空白 issue / 指向讨论区）

## 内容草案（核心工作流）
- config-ci.yml（示例）
```
name: config CI
on:
  push:
    paths:
      - 'config/**'
  pull_request:
    paths:
      - 'config/**'
jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: [ '1.21', '1.22' ]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}
          cache: true
      - name: Test
        working-directory: config
        run: go test ./... -race -coverprofile=coverage.out
      - name: Upload coverage
        uses: actions/upload-artifact@v4
        with:
          name: coverage-${{ matrix.go-version }}
          path: config/coverage.out
```
- config-lint.yml（示例）
```
name: config Lint
on:
  push:
    paths:
      - 'config/**'
  pull_request:
    paths:
      - 'config/**'
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest
          args: --out-format=github-actions
          working-directory: config
```
- dependabot.yml（示例）
```
version: 2
updates:
  - package-ecosystem: gomod
    directory: "/"
    schedule:
      interval: weekly
  - package-ecosystem: gomod
    directory: "/config"
    schedule:
      interval: weekly
```

## README 徽章（可选增强）
- CI 状态徽章：`https://img.shields.io/github/actions/workflow/status/goal-web/config/config-ci.yml?branch=main`
- 覆盖率徽章：如接入 Codecov 可添加 `https://img.shields.io/codecov/c/github/goal-web/config`（需要后续配置）

## 验证与兼容
- 使用路径过滤确保仅在 `config/` 变更时触发，避免 monorepo 全量构建。
- 如需全仓库测试，后续增加模块矩阵或遍历 `*/go.mod` 的脚本步骤。
- 初始版本不依赖外部密钥；后续接入 Codecov 或发布流程再补充密钥配置。

## 交付与成功标准
- 新增文件位于 `.github/` 目录，合并后在 PR 中自动执行测试与静态检查。
- README 增加 CI 徽章后可展示状态；Dependabot PR 按计划周期产生。