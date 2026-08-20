# Castor CLI (cstr) Oh My Zsh Plugin

Official [Oh My Zsh](https://ohmyz.sh/) plugin for **`cstr`** (*Castor CLI*).

Provides rich subcommand and flag autocompletion alongside productivity aliases (`cstra`, `cstrv`, `cstrl`, `cstrs`, `cstrc`, `cstri`).

## Installation

### 1. Copy or Symlink Plugin Directory

```bash
mkdir -p ~/.oh-my-zsh/custom/plugins
cp -r cmd/cstr/plugins/oh-my-zsh/cstr ~/.oh-my-zsh/custom/plugins/
```

### 2. Enable Plugin in `~/.zshrc`

Add `cstr` to your plugins list in `~/.zshrc`:

```zsh
plugins=(
  git
  cstr
)
```

### 3. Reload Zsh Configuration

```bash
source ~/.zshrc
```

## Useful Aliases

| Alias | Command | Description |
| :--- | :--- | :--- |
| `cstra` | `cstr add` | Add skills from URI |
| `cstrver` | `cstr verify` | Verify installed skills against lockfile |
| `cstrv` | `cstr validate` | Audit skills against 5-point SDLC |
| `cstrl` | `cstr list` | List registered skills |
| `cstrs` | `cstr search` | Search skills by keyword |
| `cstrc` | `cstr compile` | Pre-compile skills manifest JSON |
| `cstri` | `cstr init` | Scaffold new skill directory |
