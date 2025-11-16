## 目标
- 在 `config/.github/` 下新增独立的 CI/Lint/Dependabot/Issue 模板，使该组件作为独立仓库时即可生效；不依赖根目录全局配置。

## 添加内容
- `config/.github/workflows/ci.yml`：push/PR 触发，Go 1.21/1.22 矩阵，`go test ./... -race -coverprofile`，上传覆盖率。
- `config/.github/workflows/lint.yml`：push/PR 触发，使用 `golangci-lint-action` 进行静态检查。
- `config/.github/dependabot.yml`：每周检查该组件的 `gomod` 依赖更新（目录 `/`）。
- `config/.github/ISSUE_TEMPLATE/{bug_report.yml, feature_request.yml, config.yml}`：标准化问题与需求收集。

## 注意
- 这些文件在独立仓库下生效；当前 monorepo 环境中它们不会影响根仓库的自动化行为。

## 成功标准
- `config` 目录包含完整 `.github` 配置；在独立仓库场景下可直接运行 CI/Lint，并按周创建依赖更新 PR。