[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE.md)
[![macOS](https://img.shields.io/badge/macOS-supported-brightgreen?logo=apple)](https://github.com/yamd1/dotfiles)
[![Linux](https://img.shields.io/badge/Linux-supported-brightgreen?logo=linux&logoColor=white)](https://github.com/yamd1/dotfiles)
[![WSL2](https://img.shields.io/badge/WSL2-supported-brightgreen?logo=windows&logoColor=white)](https://github.com/yamd1/dotfiles)

# dotfiles

> Personal development environment for macOS / Linux / WSL2

---

## Stack

| Category | Tool |
| :--- | :--- |
| Shell | [zsh](https://www.zsh.org/) + [sheldon](https://github.com/rossmacarthur/sheldon) |
| Prompt | [Starship](https://starship.rs/) |
| Terminal | [Ghostty](https://ghostty.org/) / [WezTerm](https://wezfurlong.org/wezterm/) |
| Editor | [Neovim](https://neovim.io/) |
| Git TUI | [lazygit](https://github.com/jesseduffield/lazygit) |
| Files | [eza](https://github.com/eza-community/eza) + [fd](https://github.com/sharkdp/fd) |
| Search | [ripgrep](https://github.com/BurntSushi/ripgrep) + [fzf](https://github.com/junegunn/fzf) |
| Pager | [bat](https://github.com/sharkdp/bat) + [delta](https://github.com/dandavison/delta) |
| Runtime | [mise](https://mise.jdx.dev/) |
| CLI tools | [aqua](https://aquaproj.github.io/) |
| Key remap | [Karabiner-Elements](https://karabiner-elements.pqrs.org/) |

---

## Install

### macOS / Linux

```sh
git clone https://github.com/yamd1/dotfiles.git $HOME/dotfiles
./dotfiles/install.sh
```

### Windows (WSL2)

Install WezTerm first:

```powershell
winget install wez.wezterm
```

Then inside WSL:

```sh
git clone https://github.com/yamd1/dotfiles.git $HOME/dotfiles
./dotfiles/install.sh wsl
```

---

## Structure

```
dotfiles/
├── files/
│   ├── .config/          # XDG config (symlinked to ~/.config)
│   │   ├── nvim/
│   │   ├── zsh/
│   │   ├── starship/
│   │   ├── wezterm/
│   │   ├── ghostty/
│   │   ├── lazygit/
│   │   ├── git/
│   │   └── ...
│   └── .local/           # XDG local (symlinked to ~/.local)
├── .bin/                 # Install scripts
└── install.sh
```

---

## License

[MIT](./LICENSE.md)
