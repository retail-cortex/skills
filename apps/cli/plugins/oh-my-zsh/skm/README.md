# SKM Oh My Zsh Plugin

Official [Oh My Zsh](https://ohmyz.sh/) plugin for **`skm`** (*Skill Manager*).

Provides rich subcommand and flag autocompletion alongside productivity aliases (`skma`, `skmv`, `skml`, `skms`, `skmc`, `skmi`).

## Installation

### 1. Copy or Symlink Plugin Directory

```bash
mkdir -p ~/.oh-my-zsh/custom/plugins
cp -r apps/cli/plugins/oh-my-zsh/skm ~/.oh-my-zsh/custom/plugins/
```

### 2. Enable Plugin in `~/.zshrc`

Add `skm` to your plugins list in `~/.zshrc`:

```zsh
plugins=(
  git
  skm
)
```

### 3. Reload Zsh Configuration

```bash
source ~/.zshrc
```

## Useful Aliases

| Alias | Command | Description |
| :--- | :--- | :--- |
| `skma` | `skm add` | Add skills from URI |
| `skmv` | `skm validate` | Audit skills against 5-point SDLC |
| `skml` | `skm list` | List registered skills |
| `skms` | `skm search` | Search skills by keyword |
| `skmc` | `skm compile` | Pre-compile skills manifest JSON |
| `skmi` | `skm init` | Scaffold new skill directory |
