# igmeek

`igmeek` 是一个面向 [Gmeek](https://github.com/Meekdai/Gmeek) 博客工作流的本地优先命令行工具，用来在本地管理 GitHub Issues 和 Labels。

你可以在本地编辑 Markdown 文件，然后通过 CLI 创建、更新、关闭、重开 Issue，并维护 Gmeek 依赖的标签体系，尽量不需要频繁打开 GitHub 网页界面。

## 项目定位

Gmeek 的核心发布模型依赖 GitHub Issues 与 Labels。`igmeek` 提供了一套本地工作流：

- 在本地写文章 Markdown
- 用命令创建或更新 GitHub Issue
- 用标签控制文章分类、发布与草稿状态
- 把远端仓库的 issue 和 label 同步到本地缓存
- 在一套配置里切换多个 Gmeek 仓库

## 功能特性

- 本地优先：在任意目录编辑 Markdown，再通过 CLI 发布到 GitHub
- 多仓库管理：支持添加、切换、列出、删除多个 `owner/repo`
- 初始化即同步：`igmeek init` 保存 Token 后会立即同步目标仓库
- 添加仓库即同步：`igmeek repo add` 验证仓库权限后会立即同步
- 全量同步：同步当前所选远端仓库全部 issue 和 label 到本地缓存
- 新建文章：从 Markdown 文件创建 GitHub Issue
- 更新文章：按文件路径或 issue 编号更新 issue 内容
- 标签维护：支持新增标签、列出全部标签、增删改 issue 标签
- 关闭与恢复：关闭 issue 或重新打开已关闭 issue
- 文件名歧义提示：当同名文件对应多个 issue 时，会提示候选项并要求显式指定 issue 编号
- 跨目录使用：不要求你在 Gmeek 仓库根目录执行命令
- 跨平台配置目录：使用系统标准用户配置目录保存数据

## 安装

### 从源码构建

```bash
git clone https://github.com/CJSen/igmeek.git
cd igmeek
go build -o igmeek .
```

构建完成后，可以把二进制加入 `PATH`，或直接在当前目录执行。

### 从 Release 下载

仓库在推送 tag 时会自动创建 GitHub Release，并生成以下平台产物：

- macOS `amd64` / `arm64`
- Linux `amd64` / `arm64`
- Windows `amd64` / `arm64`

发布的可执行文件统一命名为 `igmeek`，Windows 包内为 `igmeek.exe`。

## GitHub Token 配置

`igmeek` 需要 GitHub Personal Access Token 才能访问仓库、读写 Issues 和 Labels。

### 获取方式

在 GitHub 网页中按以下路径操作：

1. 打开 GitHub。
2. 进入 `Settings`。
3. 进入 `Developer settings`。
4. 打开 `Personal access tokens`。
5. 选择 `Fine-grained tokens`。
6. 点击 `Generate new token`。
7. 按上面的权限配置生成 token。
8. 复制生成后的 token，后续用于 `igmeek init` 或环境变量配置。

### 配置方式

你可以通过交互式初始化保存 token：

```bash
igmeek init
```

也可以通过环境变量覆盖配置文件中的 token：

```bash
export IMGEEK_GITHUB_TOKEN=your_token_here
```

注意：环境变量 `IMGEEK_GITHUB_TOKEN` 的优先级高于本地配置文件。

## 快速开始

### 1. 初始化

`igmeek init` 会交互式要求输入：

- GitHub Token
- 仓库地址，支持 `owner/repo` 或 GitHub URL

例如：

```bash
igmeek init
```

执行后会：

- 创建全局配置目录
- 保存 token 到配置文件
- 把当前仓库加入仓库列表
- 将该仓库设为当前仓库
- 立即执行一次全量同步

### 2. 添加仓库

如果已经初始化过 token，可以继续添加其他仓库：

```bash
igmeek repo add owner/repo
```

也支持直接传 GitHub URL：

```bash
igmeek repo add https://github.com/owner/repo
```

如果不带参数，会进入交互式输入。

添加仓库时会：

- 校验当前 token 是否能访问该仓库
- 保存仓库配置
- 在本地创建该仓库的数据目录
- 自动同步该仓库的 issues 和 labels

### 3. 切换当前仓库

```bash
igmeek repo use
```

如果只配置了一个仓库，会直接选中；如果有多个仓库，会提示你选择。

### 4. 同步远端数据

```bash
igmeek sync
```

该命令会拉取当前仓库：

- 全部 open 和 closed issues
- 全部 labels

并更新本地缓存文件。

### 5. 新建文章 Issue

```bash
igmeek new post.md --tag blog,tech
```

或创建一个不带标签的草稿 issue：

```bash
igmeek new draft.md --notag
```

实际行为说明：

- `new` 必须使用 `--tag` 或 `--notag` 二选一
- `--tag` 支持英文逗号 `,` 和中文逗号 `，` 分隔多个标签
- issue 标题默认取自文件名，不是 Markdown 第一行标题
- issue 正文取整个 Markdown 文件内容
- 创建时会直接带上 labels，一次性创建 issue
- 如果指定的标签在远端仓库中不存在，会直接报错，并列出缺失标签；此时不会创建 issue
- 创建成功后会把 issue 信息写入本地索引

### 6. 更新已有 Issue

按文件更新：

```bash
igmeek update post.md
```

按 issue 编号和文件更新：

```bash
igmeek update 42 post.md
```

更新标签：

```bash
igmeek update post.md --add-tag new-tag
igmeek update post.md --remove-tag old-tag
igmeek update post.md --set-tag tag1,tag2
```

说明：

- `update <file>` 会尝试在本地索引里按文件路径匹配对应 issue
- 如果同名文件匹配到多个 issue，会提示候选 issue，并要求使用 `igmeek update <num> <file>`
- 如果找不到映射，会提示先执行 `sync` 或显式传入 issue 编号
- 更新正文时，标题优先取 Markdown 首个一级标题 `# `；如果没有一级标题，则退回文件名
- `--add-tag`、`--remove-tag`、`--set-tag` 都支持英文逗号 `,` 和中文逗号 `，`
- 标签修改基于远端当前标签集合进行增删改
- 当同时更新正文和标签时，会一次性把最终内容和最终标签一起提交，避免 Gmeek workflow 在中间状态触发

### 7. 关闭或恢复 Issue

关闭 issue：

```bash
igmeek del 42
```

恢复 issue：

```bash
igmeek undel 42
```

关闭时不会删除本地 Markdown 文件，只会修改 GitHub issue 状态，并同步更新本地索引状态。

### 8. 管理标签

列出当前仓库标签：

```bash
igmeek label list
```

新增标签：

```bash
igmeek label add blog tech draft
```

说明：

- `label list` 会从 GitHub 拉取标签并刷新本地标签缓存
- `label add` 支持一次创建多个标签
- 新创建标签默认颜色是 `ededed`

### 9. 管理仓库列表

列出当前已配置仓库：

```bash
igmeek repo list
```

删除一个仓库配置：

```bash
igmeek repo del
```

说明：

- 如果只配置了一个仓库，会直接删除它
- 如果配置了多个仓库，会提示你选择要删除的仓库
- 删除仓库时会同时删除该仓库在本地的缓存目录
- 如果删除的是当前仓库，会自动切换到剩余仓库中的第一个；如果没有剩余仓库，则清空当前仓库设置

## 命令一览

### 根命令

| 命令 | 说明 |
|------|------|
| `igmeek init` | 初始化 token 和仓库，并立即同步 |
| `igmeek sync` | 全量同步当前仓库的 issues 与 labels |
| `igmeek new <file> --tag <tags>` | 从 Markdown 文件创建带标签的 issue；缺失标签时直接报错且不创建 |
| `igmeek new <file> --notag` | 从 Markdown 文件创建不带标签的 issue |
| `igmeek update <file>` | 按本地文件映射一次性更新 issue 的正文和标签 |
| `igmeek update <num> <file>` | 按 issue 编号一次性更新指定 issue 的正文和标签 |
| `igmeek del <num>` | 关闭 issue |
| `igmeek undel <num>` | 重新打开 issue |
| `igmeek repo ...` | 管理仓库配置 |
| `igmeek label ...` | 管理标签 |

### 仓库子命令

| 命令 | 说明 |
|------|------|
| `igmeek repo add [owner/repo]` | 添加仓库并自动同步 |
| `igmeek repo list` | 列出所有已配置仓库 |
| `igmeek repo use` | 选择当前活动仓库 |
| `igmeek repo del` | 删除仓库配置和本地缓存 |

### 标签子命令

| 命令 | 说明 |
|------|------|
| `igmeek label list` | 拉取并列出当前仓库所有标签 |
| `igmeek label add <tags...>` | 创建一个或多个标签 |

## Markdown 处理规则

`igmeek` 目前和 Markdown 文件交互时，标题处理规则分两种：

- `igmeek new`：issue 标题直接取文件名去掉扩展名后的结果
- `igmeek update`：优先读取 Markdown 文件中的首个一级标题 `# 标题`；如果没有，则回退为文件名

issue 正文使用文件完整内容原样提交。

## 本地数据目录

`igmeek` 使用系统标准用户配置目录保存全局配置与仓库缓存。

目录结构：

```text
<config-dir>/igmeek/
├── config.json
└── repos/
    └── <owner_repo>/
        ├── repo.json
        ├── issues_num_name.json
        └── tags.json
```

各文件用途：

- `config.json`：全局配置，保存 token、仓库列表、当前仓库
- `repo.json`：仓库元数据
- `issues_num_name.json`：issue 本地索引缓存
- `tags.json`：标签缓存

不同平台默认位置：

| 平台 | 路径 |
|------|------|
| macOS | `~/Library/Application Support/igmeek/` |
| Linux | `~/.config/igmeek/` |
| Windows | `%APPDATA%\igmeek\` |

## 环境变量

| 变量名 | 说明 |
|--------|------|
| `IMGEEK_GITHUB_TOKEN` | GitHub Token，优先级高于配置文件 |

## 错误与退出码

程序定义了以下退出码：

| 退出码 | 含义 |
|--------|------|
| `0` | 成功 |
| `1` | 通用错误 |
| `2` | 参数错误 |
| `3` | 认证错误 |
| `4` | 网络错误 |

其中最常见的认证错误是未配置 token，此时可以：

- 设置环境变量 `IMGEEK_GITHUB_TOKEN`
- 或执行 `igmeek init`

## 使用建议

- 第一次使用时优先执行 `igmeek init`
- 管理多个博客仓库时，先用 `igmeek repo add` 添加，再用 `igmeek repo use` 切换
- 修改远端 issue 或 labels 后，建议执行一次 `igmeek sync`
- 当按文件更新失败时，优先改用 `igmeek update <issue_number> <file>`
- Gmeek 的文章发布行为取决于你的标签策略，建议先统一维护仓库标签集合

## License

MIT
