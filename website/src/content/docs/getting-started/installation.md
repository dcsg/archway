---
title: Installation
description: Install Archway in under a minute
---

import { Aside, Steps, Tabs, TabItem } from '@astrojs/starlight/components';

## Install

<Tabs>
  <TabItem label="Go Install">

```bash
go install github.com/dcsg/archway/cmd/archway@latest
```

Requires **Go 1.23** or later.

  </TabItem>
  <TabItem label="From Source">

```bash
git clone https://github.com/dcsg/archway.git
cd archway
go build -o archway ./cmd/archway
```

  </TabItem>
</Tabs>

## Verify

```bash
archway --version
```

## Shell Completions

<Tabs>
  <TabItem label="Bash">

```bash
archway completion bash > /etc/bash_completion.d/archway
```

  </TabItem>
  <TabItem label="Zsh">

```bash
archway completion zsh > "${fpath[1]}/_archway"
```

  </TabItem>
  <TabItem label="Fish">

```bash
archway completion fish > ~/.config/fish/completions/archway.fish
```

  </TabItem>
</Tabs>

<Aside type="tip">
After installing completions, restart your shell or run `source ~/.bashrc` (or equivalent) for them to take effect.
</Aside>

## Next Steps

You're ready to scaffold your first service. Head to the [Quick Start](/archway/getting-started/quickstart/).
