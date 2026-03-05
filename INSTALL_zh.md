# 安装指南

本指南涵盖 clibot 的系统要求、依赖项和平台特定设置说明。

## 目录

- [系统要求](#系统要求)
- [依赖项](#依赖项)
- [平台特定设置](#平台特定设置)
  - [Linux](#linux)
  - [macOS](#macos)
  - [Windows (WSL2)](#windows-wsl2)

## 系统要求

### 支持的平台

| 平台 | 状态 | 说明 |
|----------|--------|-------|
| **Linux** | ✅ 完全支持 | Ubuntu、Debian、Fedora、CentOS、Arch 等 |
| **macOS** | ✅ 完全支持 | 10.15+ |
| **Windows** | ⚠️ 仅 WSL2 | 需要Windows 子系统 for Linux 2 |

### 为什么不支持 Windows 原生？

clibot 的 **Hook 模式**需要 `tmux` 进行会话管理，而 Windows 原生不支持 tmux。

**解决方案：**
1. **ACP 模式**（推荐）：无需 tmux，适用于所有平台
2. **WSL2**：在 Windows 子系统 for Linux 中运行 clibot

## 依赖项

### 必需软件

| 软件 | 版本 | 用途 |
|----------|---------|---------|
| **Go** | 1.24+ | 构建和运行 clibot |
| **Git** | 任意 | 克隆仓库（如果从源码安装） |
| **tmux** | 任意 | 会话管理（仅 Hook 模式需要，ACP 模式不需要） |

### 安装 Go

**Linux (Ubuntu/Debian):**
```bash
# 方式 1：从仓库安装（可能是较旧版本）
sudo apt install golang-go

# 方式 2：最新版本（推荐）
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

**macOS:**
```bash
brew install go
```

**验证安装：**
```bash
go version
# 应输出: go version go1.24.0 ...
```

### 安装 tmux（仅 Hook 模式）

**ACP 模式不需要 tmux。如果使用 ACP 模式，请跳过本节。**

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

**验证安装：**
```bash
tmux -V
# 应输出: tmux 3.x 或类似
```

### 安装 Git

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

## 平台特定设置

### Linux

clibot 在 Linux 上原生运行，具有完整功能支持。

#### 安装 clibot

**方式 1：从源码安装（推荐）：**
```bash
go install github.com/keepmind9/clibot@latest
```

**方式 2：从仓库构建：**
```bash
git clone https://github.com/keepmind9/clibot.git
cd clibot
make build
sudo make install
```

二进制文件将安装到 `~/go/bin/clibot`。确保它在你的 PATH 中：

```bash
export PATH=$PATH:~/go/bin
```

#### 配置 clibot

```bash
# 创建配置目录
mkdir -p ~/.config/clibot

# 复制配置模板
cp configs/config.mini.yaml ~/.config/clibot/config.yaml

# 编辑配置
nano ~/.config/clibot/config.yaml
```

#### 运行 clibot

```bash
clibot serve --config ~/.config/clibot/config.yaml
```

### macOS

clibot 在 macOS 上原生运行，具有完整功能支持。

#### 安装 Homebrew（如果未安装）

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

#### 安装依赖项

```bash
brew install go tmux git
```

#### 安装 clibot

```bash
go install github.com/keepmind9/clibot@latest
```

#### 配置和运行

与 Linux 说明相同。

### Windows (WSL2)

clibot 可以使用 WSL2（Windows 子系统 for Linux）在 Windows 上运行。

#### 步骤 1：安装 WSL2

以管理员身份打开 PowerShell 或命令提示符：

```powershell
# 启用 WSL
wsl --install

# 出现提示时重启计算机
```

#### 步骤 2：将 WSL2 设置为默认

```powershell
wsl --set-default-version 2
```

#### 步骤 3：安装 Ubuntu（推荐）

```powershell
# 查看可用的发行版
wsl --list --online

# 安装 Ubuntu
wsl --install -d Ubuntu
```

#### 步骤 4：完成 Ubuntu 设置

1. 从开始菜单启动 Ubuntu
2. 创建用户名和密码
3. 更新软件包：

```bash
# 在 WSL Ubuntu 终端中
sudo apt update && sudo apt upgrade -y
```

#### 步骤 5：在 WSL 中安装所需工具

```bash
# 安装 Go
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# 安装 tmux（用于 Hook 模式）
sudo apt install tmux -y

# 安装 Git
sudo apt install git -y
```

#### 步骤 6：安装和运行 clibot

```bash
# 在 WSL Ubuntu 终端中
go install github.com/keepmind9/clibot@latest

# 配置
mkdir -p ~/.config/clibot
cp /mnt/c/path/to/clibot/configs/config.yaml ~/.config/clibot/config.yaml
nano ~/.config/clibot/config.yaml

# 运行 clibot
clibot serve --config ~/.config/clibot/config.yaml
```

#### Windows + WSL2 提示

- **从 WSL 访问 Windows 文件**：`/mnt/c/Users/你的名字/...`
- **从 Windows 访问 WSL 文件**：`\\wsl$\Ubuntu\home\你的名字\...`
- **作为后台服务运行**：在 WSL 内使用 systemd
- **无需防火墙配置**：机器人使用长连接（仅出站）

#### 限制

- 剪贴板集成可能无法正常工作
- 需要文件路径转换（WSL ↔ Windows）
- 性能略低于原生 Linux

## 后续步骤

安装完成后：

1. **配置 clibot**：
   - 编辑 `~/.config/clibot/config.yaml`
   - 添加你的机器人凭据
   - 将你的用户 ID 添加到白名单

2. **选择你的模式**：
   - **ACP 模式**（推荐）：无需 tmux
   - **Hook 模式**：需要 tmux + CLI hook 配置

3. **启动 clibot**：
   ```bash
   clibot serve --config ~/.config/clibot/config.yaml
   ```

详细配置和使用说明，请参阅 [README_zh.md](README_zh.md)。
