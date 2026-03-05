# Installation Guide

This guide covers system requirements, dependencies, and platform-specific setup instructions for clibot.

## Table of Contents

- [System Requirements](#system-requirements)
- [Dependencies](#dependencies)
- [Platform-Specific Setup](#platform-specific-setup)
  - [Linux](#linux)
  - [macOS](#macos)
  - [Windows (WSL2)](#windows-wsl2)

## System Requirements

### Supported Platforms

| Platform | Status | Notes |
|----------|--------|-------|
| **Linux** | ✅ Fully Supported | Ubuntu, Debian, Fedora, CentOS, Arch, etc. |
| **macOS** | ✅ Fully Supported | 10.15+ |
| **Windows** | ⚠️ WSL2 Only | Windows Subsystem for Linux 2 required |

### Why Not Windows Native?

clibot's **Hook Mode** requires `tmux` for session management, which is not available natively on Windows.

**Workarounds:**
1. **ACP Mode** (Recommended): No tmux required, works on all platforms
2. **WSL2**: Run clibot in Windows Subsystem for Linux

## Dependencies

### Required Software

| Software | Version | Purpose |
|----------|---------|---------|
| **Go** | 1.24+ | Build and run clibot |
| **Git** | Any | Clone repository (if installing from source) |
| **tmux** | Any | Session management (Hook Mode only, not required for ACP Mode) |

### Installing Go

**Linux (Ubuntu/Debian):**
```bash
# Option 1: From repository (may be older version)
sudo apt install golang-go

# Option 2: Latest version (recommended)
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

**macOS:**
```bash
brew install go
```

**Verify installation:**
```bash
go version
# Should output: go version go1.24.0 ...
```

### Installing tmux (Hook Mode Only)

**ACP Mode does NOT require tmux. Skip this section if using ACP Mode.**

**Linux (Ubuntu/Debian):**
```bash
sudo apt-get install tmux
```

**macOS:**
```bash
brew install tmux
```

**Fedora/CentOS/RHEL:**
```bash
sudo dnf install tmux
```

**Arch Linux:**
```bash
sudo pacman -S tmux
```

**Verify installation:**
```bash
tmux -V
# Should output: tmux 3.x or similar
```

### Installing Git

**Linux:**
```bash
sudo apt-get install git  # Ubuntu/Debian
sudo dnf install git      # Fedora/CentOS
sudo pacman -S git        # Arch
```

**macOS:**
```bash
brew install git
```

## Platform-Specific Setup

### Linux

clibot runs natively on Linux with full feature support.

#### Install clibot

**Option 1: Install from source (recommended):**
```bash
go install github.com/keepmind9/clibot@latest
```

**Option 2: Build from repository:**
```bash
git clone https://github.com/keepmind9/clibot.git
cd clibot
make build
sudo make install
```

The binary will be installed at `~/go/bin/clibot`. Make sure it's in your PATH:

```bash
export PATH=$PATH:~/go/bin
```

#### Configure clibot

```bash
# Create config directory
mkdir -p ~/.config/clibot

# Copy configuration template
cp configs/config.mini.yaml ~/.config/clibot/config.yaml

# Edit configuration
nano ~/.config/clibot/config.yaml
```

#### Run clibot

```bash
clibot serve --config ~/.config/clibot/config.yaml
```

### macOS

clibot runs natively on macOS with full feature support.

#### Install Homebrew (if not installed)

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

#### Install dependencies

```bash
brew install go tmux git
```

#### Install clibot

```bash
go install github.com/keepmind9/clibot@latest
```

#### Configure and run

Same as Linux instructions above.

### Windows (WSL2)

clibot can run on Windows using WSL2 (Windows Subsystem for Linux).

#### Step 1: Install WSL2

Open PowerShell or Command Prompt as Administrator:

```powershell
# Enable WSL
wsl --install

# Restart your computer when prompted
```

#### Step 2: Set WSL2 as default

```powershell
wsl --set-default-version 2
```

#### Step 3: Install Ubuntu (recommended)

```powershell
# View available distributions
wsl --list --online

# Install Ubuntu
wsl --install -d Ubuntu
```

#### Step 4: Complete Ubuntu setup

1. Launch Ubuntu from Start menu
2. Create a username and password
3. Update packages:

```bash
# Inside WSL Ubuntu terminal
sudo apt update && sudo apt upgrade -y
```

#### Step 5: Install required tools in WSL

```bash
# Install Go
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Install tmux (for Hook Mode)
sudo apt install tmux -y

# Install Git
sudo apt install git -y
```

#### Step 6: Install and run clibot

```bash
# Inside WSL Ubuntu terminal
go install github.com/keepmind9/clibot@latest

# Configure
mkdir -p ~/.config/clibot
cp /mnt/c/path/to/clibot/configs/config.yaml ~/.config/clibot/config.yaml
nano ~/.config/clibot/config.yaml

# Run clibot
clibot serve --config ~/.config/clibot/config.yaml
```

#### Windows + WSL2 Tips

- **Access Windows files from WSL**: `/mnt/c/Users/YourName/...`
- **Access WSL files from Windows**: `\\wsl$\Ubuntu\home\yourname\...`
- **Run as background service**: Use systemd inside WSL
- **No firewall needed**: Bots use long-connections (outbound only)

#### Limitations

- Clipboard integration may not work seamlessly
- File path conversions needed (WSL ↔ Windows)
- Performance slightly lower than native Linux

## Next Steps

After installation:

1. **Configure clibot**:
   - Edit `~/.config/clibot/config.yaml`
   - Add your bot credentials
   - Add your user ID to whitelist

2. **Choose your mode**:
   - **ACP Mode** (Recommended): No tmux required
   - **Hook Mode**: Requires tmux + CLI hook configuration

3. **Start clibot**:
   ```bash
   clibot serve --config ~/.config/clibot/config.yaml
   ```

For detailed configuration and usage, see [README.md](README.md).
