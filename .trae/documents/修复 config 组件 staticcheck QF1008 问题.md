## 问题说明
- staticcheck 报错 QF1008：提示不需要通过嵌入字段选择器访问成员（示例：`provider.BaseFields.Provider = provider`）。因为 `supports.BaseFields` 作为嵌入字段，其成员已提升，直接访问即可。

## 改动方案
- 将以下三处赋值由嵌入字段选择器改为直接使用提升后的字段：
  - `config/dotenv.go:23` 替换为：`provider.Provider = provider`
  - `config/toml.go:22` 替换为：`provider.Provider = provider`
  - `config/yaml.go:22` 替换为：`provider.Provider = provider`

## 预期效果
- 消除 QF1008 静态检查问题；不影响运行时行为，保持字段赋值语义一致。

## 可选补充
- 若仍触发其他静态检查（取决于工作流使用的 golangci-lint 配置），在修复后可重新启用 `staticcheck` 以确保质量检查通过。