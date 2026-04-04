# igmeek CLI 设计规格文档

> 创建时间：2026-04-04
> 技术栈：Go + cobra + go-github
> 状态：待审阅

---

## 一、项目概述

### 1.1 背景

基于 Gmeek 框架的博客系统以 GitHub Issue 为文章载体，通过标签（Label）驱动发布。现有工作流强制依赖 GitHub Issues 网页端操作，存在编辑体验差、标签管理繁琐、无本地草稿、无法本地预览等痛点。

### 1.2 目标

构建 `igmeek` CLI，实现本地优先的 GitHub Issue/Tag 管理，专门适配 Gmeek 的标签驱动发布流程。

### 1.3 产品定位

介于通用 GitHub 客户端和 Gmeek 私有脚本之间的专用 CLI 工具。

### 1.4 非目标（第一版不做）

- `publish` 独立命令
- `diff` 命令
- 冲突检测/冲突合并
- 增量同步
- 图片上传
- 桌面应用
- Markdown 元数据高级编辑器

---

## 二、技术架构

### 2.1 技术选型

| 组件 | 选择 | 理由 |
|------|------|------|
| 语言 | Go | 编译为单一二进制文件，跨平台分发方便 |
| CLI 框架 | cobra | Go 生态最成熟的 CLI 框架，支持子命令、flag、帮助生成 |
| GitHub API | go-github | 官方维护的 Go SDK，类型安全 |
| 配置读写 | 标准库 encoding/json | 无需额外依赖 |
| 交互输入 | 标准库 + 第三方 prompt 库 | Token 初始化等交互场景 |

### 2.2 项目结构

```
igmeek/
├── cmd/
│   ├── root.go           # 根命令，全局 flag
│   ├── sync.go           # igmeek sync
│   ├── new.go            # igmeek new <file> --tag <tags>
│   ├── update.go         # igmeek update <file|num> [file]
│   ├── del.go            # igmeek del <num>
│   ├── undel.go          # igmeek undel <num>
│   ├── label.go          # igmeek label (父命令)
│   ├── label_list.go     # igmeek label list
│   ├── label_add.go      # igmeek label add <tags>
│   ├── repo.go           # igmeek repo (父命令)
│   ├── repo_add.go       # igmeek repo add [owner/repo]
│   ├── repo_del.go       # igmeek repo del
│   ├── repo_list.go      # igmeek repo list
│   └── repo_use.go       # igmeek repo use
├── internal/
│   ├── api/
│   │   └── github.go     # GitHub API 封装层
│   ├── config/
│   │   └── config.go     # 全局配置读写
│   ├── index/
│   │   └── index.go      # 仓库级索引文件读写
│   ├── markdown/
│   │   └── reader.go     # Markdown 文件读取与路径归一化
│   └── sync/
│   │   └── sync.go       # 全量同步逻辑
├── go.mod
├── go.sum
└── main.go
```

### 2.3 模块职责

| 模块 | 职责 | 输入 | 输出 | 依赖 |
|------|------|------|------|------|
| `cmd/*` | 命令解析、参数校验、调用业务层 | CLI 参数 | 终端输出 | internal/* |
| `api/github` | 封装 GitHub REST API 调用 | Token, 仓库信息 | API 响应结构体 | go-github |
| `config` | 全局配置（Token、当前仓库）读写 | 配置文件路径 | 配置结构体 | encoding/json |
| `index` | 仓库级索引文件（issues_num_name.json, tags.json）读写 | 仓库目录 | 索引结构体 | encoding/json |
| `markdown` | 读取 Markdown 文件内容，归一化路径 | 文件路径 | 文件内容 + 归一化路径 | os, filepath |
| `sync` | 全量同步远端 Issue 和标签到本地 | Token, 仓库信息 | 更新的索引文件 | api, index |

---

## 三、数据模型

### 3.1 全局配置（config.json）

存储位置：跨平台标准用户数据目录

```json
{
  "token": "ghp_xxxx",
  "current_repo": "CJSen/cjsen.github.io",
  "repos": ["CJSen/cjsen.github.io"]
}
```

### 3.2 仓库配置（repo.json）

存储位置：`<global-data-dir>/repos/<owner_repo>/repo.json`

```json
{
  "owner": "CJSen",
  "repo": "cjsen.github.io",
  "full_name": "CJSen/cjsen.github.io"
}
```

### 3.3 Issue 索引（issues_num_name.json）

存储位置：`<global-data-dir>/repos/<owner_repo>/issues_num_name.json`

```json
[
  {
    "issue_number": 1,
    "file_path": "/absolute/path/to/post.md",
    "title": "文章标题",
    "labels": ["tech", "blog"],
    "state": "open",
    "created_at": "2026-04-01T00:00:00Z",
    "updated_at": "2026-04-01T00:00:00Z",
    "closed_at": null,
    "url": "https://api.github.com/repos/...",
    "html_url": "https://github.com/.../issues/1"
  }
]
```

### 3.4 标签缓存（tags.json）

存储位置：`<global-data-dir>/repos/<owner_repo>/tags.json`

```json
[
  { "name": "tech", "color": "0075ca" },
  { "name": "blog", "color": "0075ca" }
]
```

### 3.5 路径归一化规则

- 相对路径：相对于当前工作目录解析为绝对路径
- 绝对路径：直接使用
- 所有路径在索引中统一存储为绝对路径
- 路径分隔符统一使用 OS 原生分隔符（Go 的 `filepath` 包自动处理）

---

## 四、命令规格

### 4.1 认证与初始化

**Token 获取优先级：**

1. 环境变量 `IMGEEK_GITHUB_TOKEN`
2. 全局配置文件 `config.json`
3. 交互式输入（引导用户创建 Token 并保存）

**初始化流程：**

当首次运行任何命令且无 Token 时，提示用户输入 GitHub Personal Access Token（需 `repo` 权限），保存到全局配置后提示"已保存到配置文件"。

### 4.2 `igmeek sync`

**功能：** 全量同步远端 Issue 和标签到本地缓存

**行为：**
1. 读取当前仓库配置
2. 调用 GitHub API 获取所有 Issue（分页）
3. 调用 GitHub API 获取所有标签
4. 更新 `issues_num_name.json` 和 `tags.json`
5. 输出同步结果统计

**输出示例：**
```
Synced 42 issues, 8 labels from CJSen/cjsen.github.io
```

### 4.3 `igmeek new <file> --tag <tags>`

**功能：** 创建新 Issue

**参数：**
- `<file>`：Markdown 文件路径（必填）
- `--tag <tags>`：标签列表，逗号分隔（与 `--notag` 二选一）
- `--notag`：创建不带标签的 Issue（与 `--tag` 二选一）

**行为：**
1. 读取 Markdown 文件内容
2. 使用文件名（不含扩展名）作为 Issue 标题
3. 调用 GitHub API 创建 Issue
4. 如果指定了 `--tag`，为 Issue 添加指定标签；如果指定了 `--notag`，不添加任何标签
5. 更新本地索引
6. 输出创建的 Issue 编号和 URL

**约束：** `--tag` 和 `--notag` 必须指定其一，`--tag` 至少一个标签

**输出示例：**
```
Created issue #43: 文章标题
URL: https://github.com/CJSen/cjsen.github.io/issues/43
```

### 4.4 `igmeek update <file|num> [file]`

**功能：** 更新已有 Issue

**参数形式：**
- `igmeek update <file> [--add-tag] [--remove-tag] [--set-tag]`
- `igmeek update <num> <file> [--add-tag] [--remove-tag] [--set-tag]`

**行为：**
1. 通过文件路径或 Issue 编号在索引中查找对应 Issue
2. 读取 Markdown 文件内容
3. 调用 GitHub API 更新 Issue 正文
4. 处理标签变更（追加/移除/替换）
5. 更新本地索引

**标签操作：**
- `--add-tag <tags>`：追加标签
- `--remove-tag <tags>`：移除标签
- `--set-tag <tags>`：替换全部标签

**输出示例：**
```
Updated issue #43: 文章标题
```

### 4.5 `igmeek del <num>`

**功能：** 关闭 Issue

**行为：**
1. 在索引中查找 Issue
2. 调用 GitHub API 关闭 Issue
3. 更新本地索引中的 `state` 和 `closed_at`
4. 不删除本地文件

**输出示例：**
```
Closed issue #43: 文章标题
```

### 4.6 `igmeek undel <num>`

**功能：** 重新打开 Issue

**行为：**
1. 在索引中查找 Issue
2. 调用 GitHub API 重新打开 Issue
3. 更新本地索引中的 `state`

**输出示例：**
```
Reopened issue #43: 文章标题
```

### 4.7 `igmeek label list`

**功能：** 列出仓库所有标签

**行为：**
1. 优先从本地缓存读取 `tags.json`
2. 调用 GitHub API 获取最新标签列表
3. 更新本地缓存
4. 格式化输出

**输出示例：**
```
Labels in CJSen/cjsen.github.io:
  tech
  blog
  life
  tutorial
```

### 4.8 `igmeek label add <tags>`

**功能：** 创建仓库标签

**参数：** 一个或多个标签名称

**行为：**
1. 调用 GitHub API 创建标签
2. 仅创建标签名称，不处理颜色
3. 更新本地缓存

**输出示例：**
```
Created labels: tech, blog
```

### 4.9 `igmeek repo add [owner/repo]`

**功能：** 添加仓库配置

**参数：**
- `[owner/repo]`：可选，如未提供则进入交互模式

**行为：**
1. 如果提供了参数，直接使用
2. 如果未提供，交互式输入 owner 和 repo
3. 验证仓库可访问性（调用 API 检查）
4. 创建仓库目录和配置文件
5. 添加到全局配置的 repos 列表
6. 如果是第一个仓库，自动设为 current_repo

**输出示例：**
```
Added repository: CJSen/cjsen.github.io
```

### 4.10 `igmeek repo del`

**功能：** 删除仓库配置

**行为：**
1. 如果只有一个仓库，直接删除
2. 如果有多个，交互式选择要删除的
3. 删除仓库目录和配置文件
4. 从全局配置中移除
5. 如果删除的是 current_repo，清空 current_repo 或选择下一个

**输出示例：**
```
Removed repository: CJSen/cjsen.github.io
```

### 4.11 `igmeek repo list`

**功能：** 列出已绑定的仓库

**输出示例：**
```
Configured repositories:
* CJSen/cjsen.github.io (current)
  owner/repo2
```

### 4.12 `igmeek repo use`

**功能：** 选择当前操作仓库

**行为：**
1. 列出所有已绑定的仓库
2. 交互式选择
3. 更新全局配置中的 current_repo

**输出示例：**
```
Switched to: CJSen/cjsen.github.io
```

---

## 五、错误处理

### 5.1 错误分类

| 错误类型 | 处理方式 | 示例 |
|----------|----------|------|
| 认证失败 | 提示重新配置 Token | Token 过期、权限不足 |
| 网络错误 | 提示检查网络，返回非零退出码 | 超时、DNS 解析失败 |
| 参数错误 | 显示帮助信息，返回非零退出码 | 缺少必填参数 |
| 仓库未配置 | 引导用户执行 `repo add` | 首次使用 |
| Issue 未找到 | 提示执行 `sync` 刷新索引 | 索引过期 |
| 文件不存在 | 提示检查文件路径 | 路径错误 |

### 5.2 退出码

| 退出码 | 含义 |
|--------|------|
| 0 | 成功 |
| 1 | 通用错误 |
| 2 | 参数错误 |
| 3 | 认证失败 |
| 4 | 网络错误 |

---

## 六、实施计划

### Task 列表

1. **初始化 CLI 工程与测试骨架** — `go mod init`、cobra 根命令、基础测试框架
2. **实现跨平台全局数据目录与全局配置** — 使用 `os.UserConfigDir()` 或 `os.UserDataDir()`
3. **实现 Token 解析与初始化流程** — 环境变量优先，交互式回退
4. **实现仓库注册与当前仓库选择** — `repo add/del/list/use` 命令
5. **实现 GitHub API 适配层** — go-github 封装，统一错误处理
6. **实现仓库级数据文件读写** — `repo.json`、`issues_num_name.json`、`tags.json`
7. **实现 `sync` 全量同步** — 分页获取 Issues 和 Labels
8. **实现 Markdown 文件读取与路径归一化** — 读取文件内容，路径处理
9. **实现 `new <file> --tag <tags>`** — 创建 Issue + 打标签
10. **实现 `update <file>` 与 `update <num> <file>`** — 更新 Issue 正文和标签
11. **实现 `del` 与 `undel`** — 关闭/重开 Issue
12. **实现 `label list` 与 `label add`** — 标签管理
13. **补齐帮助信息与错误提示** — cobra 自动生成的帮助 + 自定义错误消息
14. **手工联调与最小文档** — 端到端测试 + README

---

## 七、测试策略

### 7.1 单元测试

- `config` 模块：配置读写、跨平台路径
- `index` 模块：索引文件读写、查询
- `markdown` 模块：文件读取、路径归一化
- `api` 模块：使用 `go-github` 的 mock 客户端

### 7.2 集成测试

- 使用 GitHub API 的测试模式或测试仓库
- 验证完整的创建→更新→关闭→重开流程

### 7.3 手工测试

- 在真实博客仓库上执行完整工作流
- 验证 Issue 创建后 Gmeek Actions 能正常触发

---

## 八、约束与假设

### 约束

- Token 需要 `repo` 权限
- 依赖网络连接访问 GitHub API
- 第一版仅支持全量同步

### 假设

- 用户已有 Gmeek 博客仓库并配置好 GitHub Actions
- 用户了解 Gmeek 的基本工作原理
- Markdown 文件编码为 UTF-8
