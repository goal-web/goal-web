## 目标
- 将自动化配置严格按组件作用域管理，不在仓库根目录做全局触发；以 `config` 为例优化现有设置。

## 调整内容
- Dependabot：移除根目录 `gomod` 更新项，仅保留 `/config` 目录的检查，避免全仓库级升级。
- Actions 工作流：继续放置在 `.github/workflows/`（GitHub 要求），但通过 `paths: 'config/**'` 严格限定只对 `config` 组件触发。
- 说明：Issue 模板与工作流文件必须位于仓库根的 `.github/` 下才会被 GitHub 识别；不建议迁移到组件目录，否则不会生效。若需要只针对某组件使用模板，可保留模板但在说明中引导到对应组件路径。

## 实施步骤
- 更新 `.github/dependabot.yml`：删除根 `/` 条目，保留 `/config`。
- 保持现有 `config-ci.yml` 与 `config-lint.yml` 的路径过滤配置，无需额外改动。

## 后续扩展
- 为其他组件复制同样模式：新增 `<component>-ci.yml` 与 `<component>-lint.yml` 并限定各自路径；Dependabot 增加对应子目录条目。

## 成功标准
- 仅 `config/**` 的改动会触发 CI/Lint；Dependabot 仅对 `config` 模块提出依赖更新；根目录不再有全仓库范围的自动化影响。