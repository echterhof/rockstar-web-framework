---
title: "Installation"
description: "Install the Rockstar Web Framework and set up your development environment"
category: "guide"
tags: ["installation", "setup", "getting-started"]
version: "1.0.0"
last_updated: "2025-11-28"
related:
  - "GETTING_STARTED.md"
---

# Installation

This guide will help you install Go and set up the Rockstar Web Framework on your system.

## Prerequisites

The Rockstar Web Framework requires:

- **Go 1.25 or higher** - The framework is built with Go and requires a recent version
- **Git** - For cloning the repository and managing dependencies
- **C compiler** (optional) - Required only if using SQLite database support

## Installing Go

### Windows

1. Download the Windows installer from the [official Go downloads page](https://go.dev/dl/)
2. Run the MSI installer and follow the installation wizard
3. The installer will automatically add Go to your PATH

Verify the installation:

```powershell
go version
```

You should see output like: `go version go1.25.0 windows/amd64`

### macOS

**Using Homebrew (recommended):**

```bash
brew install go
```

**Using the official installer:**

1. Download the macOS package from the [official Go downloads page](https://go.dev/dl/)
2. Open the package file and follow the installation prompts
3. Go will be installed to `/usr/local/go`

Verify the installation:

```bash
go version
```

You should see output like: `go version go1.25.0 darwin/amd64`

### Linux

**Ubuntu/Debian:**

```bash
# Remove any previous Go installation
sudo rm -rf /usr/local/go

# Download and extract Go (replace VERSION with the latest version)
wget https://go.dev/dl/go1.25.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.25.0.linux-amd64.tar.gz

# Add Go to PATH (add this to ~/.profile or ~/.bashrc)
export PATH=$PATH:/usr/local/go/bin
```

**Fedora/RHEL/CentOS:**

```bash
# Using dnf
sudo dnf install golang

# Or download from official site
wget https://go.dev/dl/go1.25.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.25.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

**Arch Linux:**

```bash
sudo pacman -S go
```

Verify the installation:

```bash
go version
```

You should see output like: `go version go1.25.0 linux/amd64`

## Installing the Rockstar Web Framework

### Option 1: Using Go Modules (Recommended)

The easiest way to use Rockstar is to add it as a dependency to your Go project:

1. Create a new Go project:

```bash
mkdir my-rockstar-app
cd my-rockstar-app
go mod init my-rockstar-app
```

2. Add Rockstar as a dependency:

```bash
go get github.com/echterhof/rockstar-web-framework/pkg
```

3. The framework and all its dependencies will be automatically downloaded.

### Option 2: Cloning the Repository

If you want to explore the examples or contribute to the framework:

1. Clone the repository:

```bash
git clone https://github.com/echterhof/rockstar-web-framework.git
cd rockstar-web-framework
```

2. Download dependencies:

```bash
go mod download
```

3. Build the framework:

```bash
go build ./...
```

## Installing Database Drivers (Optional)

The framework supports multiple databases. Install the drivers you need:

**SQLite (included by default):**
```bash
# SQLite requires a C compiler (gcc on Linux/macOS, MinGW on Windows)
go get github.com/mattn/go-sqlite3
```

**PostgreSQL:**
```bash
go get github.com/lib/pq
```

**MySQL:**
```bash
go get github.com/go-sql-driver/mysql
```

**Microsoft SQL Server:**
```bash
go get github.com/microsoft/go-mssqldb
```

## Verifying Your Installation

Create a simple test file to verify everything is working:

**test.go:**
```go
package main

import (
    "fmt"
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
    config := pkg.FrameworkConfig{
        ServerConfig: pkg.ServerConfig{
            EnableHTTP1: true,
        },
    }
    
    app, err := pkg.New(config)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("✅ Rockstar Web Framework installed successfully!")
}
```

Run the test:

```bash
go run test.go
```

You should see:
```
✅ Rockstar Web Framework installed successfully!
```

## Troubleshooting

### "go: command not found"

**Problem:** Go is not in your system PATH.

**Solution:**
- **Windows:** Restart your terminal or computer after installation
- **macOS/Linux:** Add Go to your PATH by adding this line to `~/.profile` or `~/.bashrc`:
  ```bash
  export PATH=$PATH:/usr/local/go/bin
  ```
  Then run: `source ~/.profile` or `source ~/.bashrc`

### "gcc: command not found" (when using SQLite)

**Problem:** SQLite driver requires a C compiler.

**Solution:**
- **Windows:** Install [MinGW-w64](https://www.mingw-w64.org/) or [TDM-GCC](https://jmeubank.github.io/tdm-gcc/)
- **macOS:** Install Xcode Command Line Tools: `xcode-select --install`
- **Linux:** Install gcc: `sudo apt install build-essential` (Ubuntu/Debian) or `sudo dnf install gcc` (Fedora)

### "cannot find package"

**Problem:** Dependencies are not downloaded.

**Solution:**
```bash
go mod download
go mod tidy
```

### Version Mismatch

**Problem:** Your Go version is too old.

**Solution:** Upgrade to Go 1.25 or higher using the installation instructions above.

## Next Steps

Now that you have Rockstar installed, you're ready to build your first application!

- [Getting Started Tutorial →](GETTING_STARTED.md) - Build your first Rockstar application
- [Configuration Guide →](guides/configuration.md) - Learn about configuration options
- [Examples →](examples/README.md) - Explore complete example applications

## Navigation

- [← Back to Documentation Home](README.md)
- [Next: Getting Started →](GETTING_STARTED.md)
