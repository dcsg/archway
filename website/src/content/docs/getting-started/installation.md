---
title: Installation
description: How to install Archway
---

## Using Go

```bash
go install github.com/dcsg/archway/cmd/archway@latest
```

Requires Go 1.23 or later.

## From Source

```bash
git clone https://github.com/dcsg/archway.git
cd archway
go build -o archway ./cmd/archway
```

## Verify

```bash
archway --version
```

## Shell Completions

Archway supports shell completions for bash, zsh, fish, and PowerShell:

```bash
# Bash
archway completion bash > /etc/bash_completion.d/archway

# Zsh
archway completion zsh > "${fpath[1]}/_archway"

# Fish
archway completion fish > ~/.config/fish/completions/archway.fish
```
