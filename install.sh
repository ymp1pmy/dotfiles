#!/bin/bash
set -ue

DOTFILES_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Install packages for Linux
install_linux_packages() {
    echo "Installing Linux packages..."
    sudo apt-get update
    sudo apt-get install -y zsh build-essential unzip xdg-utils
    xdg-settings set default-web-browser file-protocol-handler.desktop
}

# Install packages for macOS
install_macos_packages() {
    echo "Installing macOS packages..."
    BREW_CMD=""
    if command -v brew &> /dev/null; then
        BREW_CMD="brew"
    elif [[ -x "/opt/homebrew/bin/brew" ]]; then
        BREW_CMD="/opt/homebrew/bin/brew"
    elif [[ -x "/usr/local/bin/brew" ]]; then
        BREW_CMD="/usr/local/bin/brew"
    fi

    if [[ -z "$BREW_CMD" ]]; then
        echo "Homebrew is not installed. Installing Homebrew..."
        echo "Press Enter to continue or Ctrl+C to cancel..."
        read -r
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
        if [[ -x "/opt/homebrew/bin/brew" ]]; then
            BREW_CMD="/opt/homebrew/bin/brew"
        elif [[ -x "/usr/local/bin/brew" ]]; then
            BREW_CMD="/usr/local/bin/brew"
        fi
    else
        echo "Homebrew found at: $BREW_CMD"
    fi

    $BREW_CMD install zsh coreutils unzip openssl@3
    echo "macOS package installation completed"
}

# Show daemon start instructions
show_daemon_instructions() {
    echo "=== Daemon Start Instructions ==="
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "To start BrowserPipe service on macOS, run:"
        echo "launchctl load ~/Library/LaunchAgents/com.user.browserpipe.plist"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo "To start BrowserPipe service on Linux, run:"
        echo "systemctl --user enable BrowserPipe"
        echo "systemctl --user start BrowserPipe"
    fi
    echo "================================"
}

# Main
cd "$DOTFILES_DIR"
echo "Starting dotfiles installation..."

git submodule update --init --recursive
"$DOTFILES_DIR/dotbot/bin/dotbot" -d "$DOTFILES_DIR" -c install.conf.yaml "$@"

if [[ "$OSTYPE" == "darwin"* ]]; then
    install_macos_packages
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    install_linux_packages
fi

show_daemon_instructions
echo "Installation completed successfully!"
