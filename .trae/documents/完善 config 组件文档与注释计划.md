## 总体目标
- 系统性提升 `config` 组件的可发现性、可读性与可维护性：补齐 GoDoc 注释、完善 README 与示例、形成各业务模块配置文档矩阵，并统一约定与最佳实践。

## 覆盖范围
- 核心包：`/config/*`（环境源、单例访问、服务注册、CLI 命令）。
- 业务提供器：`/goal/config/*` 与 `/goal-cli/config/*`（app/cache/http/database/redis/queue/session/filesystem/encryption/serialization/views/websocket/auth/bloomfilter/mail 等）。

## 结构化文档改进
- 重构 `/config/README.md`，形成清晰结构：
  - 快速开始与最小示例（包含 `application.Default()` 与服务注册链路）。
  - 配置加载链路与优先级图（环境变量 > 文件源 > URL 源 > 默认值）。
  - 键命名规则与展平策略（`a.b.c` → `A_B_C`），包含冲突与覆盖说明。
  - 类型化读取与 Optional 用法示例（`GetString`/`IntOptional` 等）。
  - 单例与线程安全（写入互斥、并发读取），解释 `Reload()` 语义。
  - 组合数据源实践（TOML/YAML/DotEnv + OS 环境覆盖）。
  - 错误处理与诊断（缺键、类型不匹配、文件不可读、URL 请求失败）。
  - CLI 加密命令使用指南与常见问题。
- 在 README 顶部加入目录与模块索引，交叉链接至各子模块 README。

## GoDoc 注释补齐
- 为 `config` 包所有导出 API 按 GoDoc 风格撰写简洁注释，聚焦用途、输入、行为与并发语义：
  - `config/singleton.go:13` `Default()`：返回配置单例与懒加载行为。
  - `config/singleton.go:21` `Get(key string)`：键路径规则与返回类型。
  - `config/singleton.go:25` `Set(key string, value any)`：写锁与覆盖策略。
  - `config/singleton.go:29` `Reload()`：重算各模块配置的触发场景与成本。
  - `config/config.go:9` `New(env, providers)`：从环境派生模块配置、不可变视图。
  - `config/service_provider.go:13` `NewService(env, config)`：`Register/Boot` 行为与与 `application` 的集成。
  - `config/dotenv.go:17` `NewDotEnv`、`config/toml.go:16` `NewToml`、`config/yaml.go:16` `NewYaml`：文件/URL 源、解析与展平策略。
  - `config/common.go:13` `File(path)`、`config/common.go:23` `Url(url)`：Provider 失败时机与重试建议。
  - `config/encryption.go:12` `EncryptionCommand()`：命令用途、驱动、参数与输出。
- 为关键非导出类型添加简要概览注释（例如 `config` 结构体字段含义与并发策略）。

## 子模块文档矩阵
- 为 `/goal/config/*` 与 `/goal-cli/config/*` 每个模块新增一个 README，内容模板统一：
  - 模块简介与职责边界。
  - 配置键表（键名、类型、默认值、是否必需、示例）。
  - 与环境变量对应关系（`a.b.c` → `A_B_C`）。
  - 与其他模块的依赖与交互（如 cache 依赖 redis）。
  - 常见问题（端口占用、证书、超时、连接池饱和、路径权限）。
  - 最小可运行示例与示例配置文件（TOML/YAML/DotEnv）。
- 优先从以下模块开始产出：`app`、`http`、`database`、`redis`、`cache`、`filesystem`、`queue`、`session`，随后覆盖`encryption`、`serialization`、`views`、`websocket`、`auth`、`bloomfilter`、`mail`。

## 示例与最佳实践
- 新增示例配置文件集：`examples/config/{toml,yaml,dotenv}/`，覆盖常见组合与覆盖场景。
- 新增最小应用示例：演示 `application.Default()` + `NewService` 注册、读取配置、动态 `Reload()`。
- 提供生产/开发环境分层示例与差异对比（feature flag、端点、超时、日志）。

## CLI 加密命令文档
- 在 `/config/README.md` 增加单独章节：
  - 命令用途、支持驱动与默认行为。
  - 使用示例：加密/解密文件、指定密钥/自动生成、强制覆盖、错误输出。
  - 与 `goal-cli` 的协作方式与典型流水线（预处理配置文件）。

## 一致性规范与贡献指南
- 统一键命名与默认值约定，给出 lint 规则（人工校验清单）。
- 新增贡献指南小节：添加新模块配置提供器时的文档/注释要求与示例模板。

## 自动化校验
- 为 `config` 包新增轻量测试（不改变现有行为）：
  - 优先级正确性（OS 环境覆盖文件值）。
  - 类型化读取与 Optional 默认值。
  - 组合源的展平与覆盖。
- 为示例配置做加载冒烟测试，保证 README 中示例可运行。

## 交付物与目录结构
- `/config/README.md`（重构后主文档）。
- `/config/*` GoDoc 注释完善。
- `/goal/config/<module>/README.md` 与示例配置文件。
- `/goal-cli/config/<module>/README.md`（强调 CLI 环境差异）。
- `examples/config/*` 示例工程与配置集。

## 里程碑
- M1（文档骨架）：重构主 README、输出统一模板与目录、完成 `app/http/database/redis` 四个模块 README。
- M2（注释与示例）：补齐 `config` 包 GoDoc、给出最小应用与示例配置、再覆盖 `cache/filesystem/queue/session` 四模块。
- M3（长尾与校验）：其余模块文档、CLI 加密章节完善、补充测试与冒烟验证、统一交叉链接与目录。

## 成功标准
- 关键导出 API 全部具备 GoDoc；主 README 有清晰目录与链路图；每个模块至少有一份 README 与示例配置；示例工程通过冒烟测试；新增贡献规范可指导后续扩展。