## 目标
- 修复 `staticcheck` 的 QF1008 提示，保持行为不变。

## 修改内容
- 将嵌入字段成员访问由 `provider.BaseFields.Provider = provider` 简化为 `provider.Provider = provider`（成员提升）。
- 涉及三处：
  - `config/dotenv.go:23`
  - `config/toml.go:22`
  - `config/yaml.go:22`

## 原理与影响
- `supports.BaseFields` 作为嵌入字段，其成员（`Provider`）提升到外层结构体命名空间；简化语法不影响赋值逻辑或运行时行为。

## 验证
- 变更仅为语法简化，无逻辑改动；`staticcheck` 的 QF1008 将消除。