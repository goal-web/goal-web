## 目标
- 以现代化开源项目主页风格美化 `config/README.md`，提升可读性与传播性：徽标矩阵、星标趋势、快速链接、结构化目录与亮点版块。

## 徽标矩阵（置顶）
- 添加一行徽标，覆盖：
  - Go Reference：`[![Go Reference](https://pkg.go.dev/badge/github.com/goal-web/config.svg)](https://pkg.go.dev/github.com/goal-web/config)`
  - Go Report Card：`[![Go Report Card](https://goreportcard.com/badge/github.com/goal-web/config)](https://goreportcard.com/report/github.com/goal-web/config)`
  - License（MIT）：`[![License](https://img.shields.io/badge/license-MIT-green.svg)](./LICENSE)`
  - Stars：`![GitHub Stars](https://img.shields.io/github/stars/goal-web/config?style=social)`
  - 最新版本：`![Release](https://img.shields.io/github/v/release/goal-web/config?include_prereleases)`
  - Go 版本：`![Go Version](https://img.shields.io/badge/go-%3E=%201.20-00ADD8?logo=go)`
  - 可选：CI 状态（若存在工作流）：`![CI](https://img.shields.io/github/actions/workflow/status/goal-web/config/ci.yml?branch=main)`；覆盖率（若接入 Codecov）：`![Coverage](https://img.shields.io/codecov/c/github/goal-web/config?logo=codecov&label=coverage)`；PRs 欢迎：`![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)`

## 快速链接（置顶）
- 链接区：Docs（pkg.go.dev）、Issues、Discussions（如开启）、Releases、示例与 CLI 加密章节锚点。
- 示例：
  - `[Docs](https://pkg.go.dev/github.com/goal-web/config) · [Issues](https://github.com/goal-web/config/issues) · [Releases](https://github.com/goal-web/config/releases) · [CLI 加密](#cli-加密命令)`

## 结构化目录与版块
- 在现有目录基础上补充：
  - **亮点**：一句话价值主张 + 关键能力（多源合并、环境覆盖、类型安全、线程安全、可插拔）。
  - **兼容性**：支持的 Go 版本；模块化导入路径。
  - **Roadmap**（简要列点，指向里程碑或 Issues 标签）。
  - **Changelog**（指向 Releases）。
  - **Contributing**（贡献指南与 PRs 欢迎说明）。
  - **Security**（安全提示：不提交密钥；使用 CLI 加密；环境变量优先级）。
  - **FAQ**（常见错误与排查入口，复用“错误处理与诊断”要点）。

## 星标趋势（独立章节）
- 添加“Star History”图表与链接：
  - `![Star History](https://starchart.cc/goal-web/config.svg)` 或
  - `<a href="https://star-history.com/#goal-web/config&Date"><img src="https://api.star-history.com/svg?repos=goal-web/config&type=Date" alt="Star History Chart"/></a>`

## 社交预览与头图（可选）
- 建议在仓库设置中添加 Social Preview Banner（含项目名与一句话描述），README 顶部可放置同款头图（若有）。

## 文案与内容优化
- 精炼“快速开始”与“单例用法”，统一示例风格与注释；为 CLI 加密命令增加“风险与场景”提示。
- 在“注意事项”中明确 `config.Reload()` 的生效语义与适用场景。

## 实施步骤
- 编辑 `config/README.md`：
  - 在标题下插入徽标矩阵与快速链接。
  - 增加“亮点”“兼容性”“Star History”“Contributing”“Roadmap”“Changelog”“Security”“FAQ”版块。
  - 保留并优化现有示例与目录，确保锚点一致。
- 本地预览与链接校验，确认所有徽标与跳转正常可用。

## 验证与交付
- 校验徽标图与链接可达；核对 LICENSE（MIT）一致。
- 保持中文主 README；如需要，后续提供英文版 `README.en.md`（可选）。
- 不改动代码，仅更新文档；与现有内容无冲突。

## 成功标准
- README 顶部有完整徽标与快速链接；星标趋势图可见；新增版块结构清晰；外链均正确；示例可复制运行。