# igmeek

Local-first GitHub Issue/Tag management CLI for [Gmeek](https://github.com/CJSen/Gmeek) blogs.

## Overview

`igmeek` is a command-line tool that lets you manage GitHub Issues and Tags from your local terminal, specifically tailored for blogs built with the Gmeek framework.

Write Markdown in your favorite local editor, then use `igmeek` to create, update, close, and reopen Issues — all without touching the GitHub web interface.

## Features

- **Local-first editing**: Write Markdown anywhere, publish from the CLI
- **Label management**: Add, remove, and replace labels to control Gmeek's label-driven publishing
- **Full sync**: Sync all remote issues and labels to a local index
- **No project dependency**: Works from any directory, no need to be in the Gmeek repo root
- **Cross-platform**: Single binary for Windows, macOS, and Linux

## Installation

### From Source

```bash
git clone https://github.com/CJSen/igmeek.git
cd igmeek
go build -o igmeek .
```

Move the binary to your `PATH`, or use it directly.

## Quick Start

### 1. Initialize

Set your GitHub Personal Access Token (requires `repo` scope):

```bash
igmeek init
```

Or set the environment variable:

```bash
export IMGEEK_GITHUB_TOKEN=ghp_xxxx
```

### 2. Add a Repository

```bash
igmeek repo add owner/repo
```

### 3. Sync

Download all existing issues and labels to your local index:

```bash
igmeek sync
```

### 4. Create an Issue

```bash
igmeek new my-post.md --tag blog,tech
```

Or create a draft without labels:

```bash
igmeek new my-draft.md --notag
```

### 5. Update an Issue

```bash
igmeek update my-post.md
```

Add or change labels:

```bash
igmeek update my-post.md --add-tag new-tag
igmeek update my-post.md --remove-tag old-tag
igmeek update my-post.md --set-tag tag1,tag2
```

Or update by issue number:

```bash
igmeek update 42 my-post.md
```

### 6. Close / Reopen

```bash
igmeek del 42
igmeek undel 42
```

## Commands

| Command | Description |
|---------|-------------|
| `igmeek init` | Initialize GitHub Token (interactive) |
| `igmeek sync` | Full sync of remote issues and labels to local cache |
| `igmeek new <file> --tag <tags>` | Create a new Issue with labels |
| `igmeek new <file> --notag` | Create a new Issue without labels (draft) |
| `igmeek update <file>` | Update an Issue linked to a file |
| `igmeek update <num> <file>` | Update a specific Issue by number |
| `igmeek del <num>` | Close an Issue (preserves local file) |
| `igmeek undel <num>` | Reopen a closed Issue |
| `igmeek label list` | List repository labels |
| `igmeek label add <tags>` | Create repository labels |
| `igmeek repo add [owner/repo]` | Add a repository configuration |
| `igmeek repo del` | Remove a repository configuration |
| `igmeek repo list` | List configured repositories |
| `igmeek repo use` | Switch the active repository |

## Configuration

### Global Data Directory

Configuration and index files are stored in the standard user config directory:

```
<config-dir>/igmeek/
├── config.json                    # Global config (token, repos, current repo)
└── repos/
    └── <owner_repo>/
        ├── repo.json              # Repository metadata
        ├── issues_num_name.json   # Issue index
        └── tags.json              # Tag cache
```

Locations by platform:

| Platform | Path |
|----------|------|
| macOS | `~/Library/Application Support/igmeek/` |
| Linux | `~/.config/igmeek/` |
| Windows | `%APPDATA%\igmeek\` |

### Environment Variables

| Variable | Description |
|----------|-------------|
| `IMGEEK_GITHUB_TOKEN` | GitHub Personal Access Token (overrides config file) |

## Exit Codes

| Code | Meaning |
|------|--------|
| 0 | Success |
| 1 | General error |
| 2 | Parameter error |
| 3 | Authentication error |
| 4 | Network error |

## License

MIT
