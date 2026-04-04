# igmeek-cli 项目上下文与进度历史

> 更新时间：2026-04-04 20:15
> 用途：跨 Session 恢复上下文，新对话可直接基于此文件继续

---

## 一、项目背景

### 痛点
用户基于 Gmeek 框架搭建博客，核心机制是：
- 每篇文章 = 一个 GitHub Issue
- **必须至少有一个 Label 才会被识别为可发布文章**
- Issue 编辑/标签变更后，GitHub Actions 自动触发生成静态页并部署

现有工作流痛点：
- 只能在 GitHub Issues 网页上编辑，无法用本地编辑器
- 标签管理繁琐，依赖网页操作
- 无本地草稿概念
- 图片管理不便
- 无法本地预览

### 目标
构建 `igmeek` CLI，实现：
- 本地任意路径编辑 Markdown
- 命令行创建/更新/关闭/重开 Issue
- 通过标签管理适配 Gmeek 发布流程
- 不依赖 Gmeek 仓库根目录
- 不依赖 `backup/` 目录
- 跨平台（Win/Mac/Linux）

---

## 二、产品定位

`igmeek` 是一个本地优先的 GitHub Issue / Tag 管理 CLI，对 Gmeek 的"标签驱动发布"流程做专门适配。

不是通用 GitHub 客户端，也不是 Gmeek 私有脚本，而是介于两者之间。

---

## 三、已确定的设计决策

### 技术栈
- **语言**: Go（编译为单一二进制文件，跨平台分发方便）
- **CLI 框架**: cobra
- **GitHub API**: go-github/v68 + golang.org/x/oauth2
- **配置读写**: 标准库 encoding/json

### 工作区模型
- **不依赖** `backup/` 目录
- **不依赖** Gmeek 博客仓库根目录
- 文章文件允许任意相对路径或绝对路径
- 主键是 `issue_number`
- 本地通过索引维护 `file_path <-> issue_number` 映射

### 同步策略
- 默认**全量同步**
- 第一版不做增量同步
- 第一版不做 diff/冲突工作流
- 用户自行决定覆盖行为

### 认证
- Token 优先从环境变量 `IMGEEK_GITHUB_TOKEN` 读取
- 若无，进入初始化交互流程（`igmeek init`）
- 保存到全局 `config.json`
- 保存后明确提示"已保存到配置文件"

### 全局数据目录
- 跨平台标准用户数据目录（`os.UserConfigDir()`）
- 按仓库隔离存储
- 结构：
  ```
  <global-data-dir>/config.json
  <global-data-dir>/repos/<owner_repo>/repo.json
  <global-data-dir>/repos/<owner_repo>/issues_num_name.json
  <global-data-dir>/repos/<owner_repo>/tags.json
  ```

### 索引文件
`issues_num_name.json` 采用完整型，字段：
- `issue_number`
- `file_path`
- `title`
- `labels`
- `state`
- `created_at`
- `updated_at`
- `closed_at`
- `url`
- `html_url`

### 标签规则
- Gmeek 发布依赖至少一个标签
- `new` 必须带 `--tag`（至少一个）或 `--notag`，二选一
- `update` 允许 `--add-tag` / `--remove-tag` / `--set-tag`
- `label add` 只创建标签名称，不处理颜色

### 退出码
| 退出码 | 含义 |
|--------|------|
| 0 | 成功 |
| 1 | 通用错误 |
| 2 | 参数错误 |
| 3 | 认证失败 |
| 4 | 网络错误 |

### 第一版明确不做
- `publish` 独立命令
- `diff` 命令
- 冲突检测/冲突合并
- 增量同步
- 图片上传
- 桌面应用
- Markdown 元数据高级编辑器

---

## 四、命令集（已确认）

| 命令 | 说明 |
|------|------|
| `igmeek init` | 初始化 GitHub Token（交互式） |
| `igmeek sync` | 全量同步远端 issue 和标签到本地缓存 |
| `igmeek new <file> --tag <tags>` | 创建新 Issue，`--tag` 与 `--notag` 二选一 |
| `igmeek new <file> --notag` | 创建不带标签的 Issue（草稿） |
| `igmeek update <file> [--add-tag] [--remove-tag] [--set-tag]` | 按索引找到 issue 更新正文，可改标签 |
| `igmeek update <num> <file> [--add-tag] [--remove-tag] [--set-tag]` | 显式指定编号更新 |
| `igmeek del <num>` | 关闭 Issue，不删本地文件 |
| `igmeek undel <num>` | 重新打开 Issue |
| `igmeek label list` | 列出仓库标签 |
| `igmeek label add <tags>` | 创建仓库标签，仅名称 |
| `igmeek repo add [owner/repo]` | 添加仓库，支持参数或交互 |
| `igmeek repo del` | 删除仓库配置 |
| `igmeek repo list` | 列出已绑定仓库 |
| `igmeek repo use` | 选择当前操作仓库 |

---

## 五、项目结构

```
.                                # 项目根目录（也是 CLI 代码根目录）
├── main.go                      # 入口，调用 cmd.Execute()
├── go.mod                       # github.com/CJSen/igmeek/cli
├── go.sum
├── .gitignore
├── cmd/
│   ├── root.go                  # 根命令，TokenError，GetToken()，退出码常量
│   ├── init.go                  # igmeek init
│   ├── sync.go                  # igmeek sync
│   ├── new.go                   # igmeek new (--tag/--notag)
│   ├── update.go                # igmeek update (--add-tag/--remove-tag/--set-tag)
│   ├── del.go                   # igmeek del
│   ├── undel.go                 # igmeek undel
│   ├── label.go                 # label 父命令
│   ├── label_list.go            # label list
│   ├── label_add.go             # label add
│   ├── repo.go                  # repo 父命令
│   ├── repo_add.go              # repo add
│   ├── repo_del.go              # repo del
│   ├── repo_list.go             # repo list
│   └── repo_use.go              # repo use
├── internal/
│   ├── config/
│   │   ├── config.go            # 全局配置 + 仓库配置 JSON 读写
│   │   └── config_test.go       # 4 tests PASS
│   ├── api/
│   │   ├── github.go            # go-github 封装：Issues/Labels CRUD
│   │   └── github_test.go       # 2 tests PASS
│   ├── index/
│   │   ├── index.go             # issues_num_name.json + tags.json 读写
│   │   └── index_test.go        # 5 tests PASS
│   ├── markdown/
│   │   ├── reader.go            # Markdown 读取 + 路径归一化 + 标题提取
│   │   └── reader_test.go       # 8 tests PASS
│   └── sync/
│       ├── sync.go              # 全量同步逻辑 + 数据转换
│       ├── sync_test.go         # 2 tests PASS
│       └── parse.go             # ParseOwnerRepo 工具函数
├── README.md                    # 完整使用说明文档
└── docs/
    ├── igmeek-cli-history.md    # 项目上下文与进度历史
    ├── gmeek-need-doc.md        # 需求文档
    └── superpowers/             # 设计与计划
```

---

## 六、当前进度

### 已完成 (Task 1-14) ✅

| Task | 内容 | 提交 SHA | 状态 |
|------|------|----------|------|
| 1 | 初始化 Go 模块与 cobra 根命令 | `c48a5da` | ✅ |
| 2 | 跨平台全局数据目录与全局配置 | `a389134` | ✅ |
| 3 | Token 解析与初始化流程 | `da12efb` | ✅ |
| 4 | GitHub API 适配层 | `2f3bf8c` | ✅ |
| 5 | 仓库级数据文件读写（索引模块） | `7e305e6` | ✅ |
| 6 | Markdown 文件读取与路径归一化 | `182633c` | ✅ |
| 7 | sync 全量同步 | `5afd3ad` | ✅ |
| 8 | repo add/del/list/use 命令 | `dacfcab` | ✅ |
| 9 | new 命令 (--tag/--notag) | `67a7795` | ✅ |
| 10 | update 命令 (标签操作) | `4128a69` | ✅ |
| 11 | del 与 undel 命令 | `6c756cd` | ✅ |
| 12 | label list 与 label add 命令 | `bf8a51b` | ✅ |
| 13 | 补齐帮助信息与错误提示 | | ✅ |
| 14 | README + 全量测试 | | ✅ |

### 全部完成

14/14 个 Task 全部完成。

---

## 七、关键文件路径参考

### 博客仓库
- 本地路径：`/Users/css/dev/igmeek/cjsen.github.io/`
- 远程仓库：`CJSen/cjsen.github.io`
- Gmeek 源码：`/Users/css/dev/igmeek/Gmeek/`（分支 `myself-use`）

### 需求文档
- `/Users/css/dev/igmeek/docs/gmeek-need-doc.md`

### 设计与计划
- Spec: `docs/superpowers/specs/2026-04-04-igmeek-cli-design.md`
- Plan: `docs/superpowers/plans/2026-04-04-igmeek-cli-implementation.md`

### 工作区
- `/Users/css/dev/igmeek/`（也是 CLI 代码根目录）

---

## 八、新 Session 启动建议

1. 读取本文件恢复上下文
2. 所有 14 个 Task 已完成，第一版功能齐全
3. 可选：在真实博客仓库上端到端测试
4. 可选：考虑后续功能（增量同步、diff、图片上传等）
