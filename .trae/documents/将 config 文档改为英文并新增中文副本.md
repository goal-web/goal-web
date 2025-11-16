## 目标
- 将 `config/README.md` 改为英文版主文档，并把现有中文内容复制为 `config/README.cn.md`；在英文版顶部提供中文链接。

## 修改项
- 新增 `config/README.cn.md`：完整复制当前中文内容（包含已加入的徽标、目录、CLI 加密、Star History 等）。
- 重写 `config/README.md` 为英文：
  - 保留徽章矩阵与快速链接（Docs/Issues/Releases/CLI Encryption + 中文版链接）。
  - 将各章节改为英文：Highlights、Compatibility、Installation、Quick Start、Supported Formats、Config Priority、Key Naming、Examples、Reading Config、Optional Getters、Singleton Usage、Combining Sources、Env Override、Error Handling & Diagnostics、CLI Encryption、Star History、Contributing、Roadmap、Changelog、Security、FAQ。

## 实施步骤
- 读取当前 `config/README.md` 内容作为中文源。
- 创建新文件 `config/README.cn.md`，写入中文内容。
- 用英文版内容替换 `config/README.md` 并添加到顶部的语言切换链接。

## 成功标准
- `config` 目录下同时存在英文版与中文版 README；英文版主页结构现代化、包含中文文档链接；示例与锚点可用。