---
title: "Platform-Specific Optimization Guide"
description: "Platform-specific optimizations for Unix/Linux and Windows deployments"
category: "guide"
tags: ["platform", "optimization", "unix", "linux", "windows", "performance"]
version: "1.0.0"
last_updated: "2025-11-29"
---

# Platform-Specific Optimization Guide

## Table of Contents

- [Overview](#overview)
- [Unix/Linux Optimizations](#unixlinux-optimizations)
- [Windows Optimizations](#windows-optimizations)
- [Platform-Specific Listener Differences](#platform-specific-listener-differences)
- [Platform-Specific Network Tuning](#platform-specific-network-tuning)
- [Platform-Specific Diagnostics](#platform-specific-diagnostics)
- [Cross-Platform Best Practices](#cross-platform-best-practices)

## Overview

The Rockstar Web Framework provides platform-specific optimizations to maximize performance on different operating systems. This guide covers Unix/Linux and Windows-specific configurations, tuning parameters, and diagnostic approaches.

### Platform Detection

The framework automatically detects the platform and applies appropriate optimizations:

```go
import "github.com/echterhof/rockstar-web-framework/pkg"

// Get platform information
info := pkg.GetPlatformInfo()
fmt.Printf("OS: %s, Arch: %s, CPUs: %d\n", info.OS, info.Arch, info.NumCPU)
fmt.Printf("Supports Prefork: %v\n", info.SupportsPrefork)
fmt.Printf("Supports ReusePort: %v\n", info.SupportsReusePort)
```

### Key Platform Differences

| Feature | Unix/Linux | Windows |
|---------|-----------|---------|
| Prefork Mode | ✅ Supported (fork) | ✅ Supported (spawn) |
| SO_REUSEPORT | ✅ Supported | ❌ Not supported |
| SO_REUSEADDR | ✅ Supported | ✅ Supported (different behavior) |
| Signal Handling | POSIX signals | Windows events |
| File Descriptors | ulimit configuration | Handle limits |
| I/O Model | epoll/kqueue | IOCP |

## Unix/Linux Optimizations

### Prefork Mode Configuration

Prefork mode creates multiple worker processes to handle requests, improving performance on multi-core systems.

#### Basic Prefork Configuration

```go
config := pkg.FrameworkConfig{
    Server: pkg.ServerConfig{
        EnableHTTP1: true,
        HTTP1Port:   8080,
        ListenerConfig: pkg.ListenerConfig{
            EnablePrefork:  true,
            PreforkWorkers: 4, // Number of worker processes
            ReusePort:      true, // Required for prefork on Linux
            ReuseAddr:      true,
        },
    },
}

app, err := pkg.New(config)
if err != nil {
    log.Fatal(err)
}

// Start server - master process will fork workers
if err := app.Start(); err != nil {
    log.Fatal(err)
}
```

#### Optimal Worker Count

```go
import "runtime"

// Rule of thumb: 1-2 workers per CPU core
numWorkers := runtime.NumCPU()

// For CPU-bound workloads
numWorkers = runtime.NumCPU()

// For I/O-bound workloads
numWorkers = runtime.NumCPU() * 2

// For mixed workloads
numWorkers = runtime.NumCPU() + (runtime.NumCPU() / 2)

config.Server.ListenerConfig.PreforkWorkers = numWorkers
```

### File Descriptor Limit Tuning

Unix systems limit the number of open file descriptors per process. For high-traffic servers, increase these limits.

#### Check Current Limits

```bash
# Soft limit (current session)
ulimit -n

# Hard limit (maximum allowed)
ulimit -Hn

# System-wide limits
cat /proc/sys/fs/file-max
```

#### Temporary Increase (Current Session)

```bash
# Increase soft limit to 65536
ulimit -n 65536
```

#### Permanent Configuration

Edit `/etc/security/limits.conf`:

```
# Format: <domain> <type> <item> <value>
*               soft    nofile          65536
*               hard    nofile          65536
root            soft    nofile          65536
root            hard    nofile          65536
```

Edit `/etc/sysctl.conf` for system-wide limits:

```
# Maximum number of file descriptors
fs.file-max = 2097152

# Maximum number of inotify watches
fs.inotify.max_user_watches = 524288
```

Apply changes:

```bash
sudo sysctl -p
```

#### Verify in Application

```go
import (
    "fmt"
    "syscall"
)

func checkFileDescriptorLimits() {
    var rLimit syscall.Rlimit
    err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
    if err != nil {
        fmt.Println("Error getting rlimit:", err)
        return
    }
    fmt.Printf("Current FD limit: %d\n", rLimit.Cur)
    fmt.Printf("Maximum FD limit: %d\n", rLimit.Max)
}
```

### TCP Stack Tuning Parameters

Optimize the Linux TCP stack for high-performance web servers.

#### Recommended sysctl Settings

Edit `/etc/sysctl.conf`:

```bash
# TCP buffer sizes (min, default, max in bytes)
net.ipv4.tcp_rmem = 4096 87380 16777216
net.ipv4.tcp_wmem = 4096 65536 16777216

# Maximum socket receive/send buffer sizes
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216

# Increase the maximum number of connections
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 8192

# Enable TCP Fast Open
net.ipv4.tcp_fastopen = 3

# TCP keepalive settings
net.ipv4.tcp_keepalive_time = 600
net.ipv4.tcp_keepalive_intvl = 60
net.ipv4.tcp_keepalive_probes = 3

# Reuse TIME_WAIT sockets
net.ipv4.tcp_tw_reuse = 1

# Increase local port range
net.ipv4.ip_local_port_range = 10000 65535

# TCP congestion control (BBR for better performance)
net.core.default_qdisc = fq
net.ipv4.tcp_congestion_control = bbr

# Disable TCP slow start after idle
net.ipv4.tcp_slow_start_after_idle = 0
```

Apply settings:

```bash
sudo sysctl -p
```

#### Verify TCP Settings

```bash
# Check specific setting
sysctl net.ipv4.tcp_rmem

# Check all TCP settings
sysctl -a | grep tcp

# Check BBR is enabled
sysctl net.ipv4.tcp_congestion_control
```

### systemd Service Configuration

Create a production-ready systemd service for your application.

#### Service File Example

Create `/etc/systemd/system/rockstar-app.service`:

```ini
[Unit]
Description=Rockstar Web Application
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/opt/rockstar-app

# Environment variables
Environment="ROCKSTAR_ENV=production"
Environment="ROCKSTAR_PORT=8080"

# Increase file descriptor limit
LimitNOFILE=65536

# Resource limits
LimitNPROC=512
LimitCORE=infinity

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/rockstar-app/data

# Restart policy
Restart=always
RestartSec=10
StartLimitInterval=60
StartLimitBurst=3

# Executable
ExecStart=/opt/rockstar-app/server
ExecReload=/bin/kill -HUP $MAINPID
ExecStop=/bin/kill -TERM $MAINPID

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=rockstar-app

[Install]
WantedBy=multi-user.target
```

#### Service Management

```bash
# Reload systemd configuration
sudo systemctl daemon-reload

# Enable service to start on boot
sudo systemctl enable rockstar-app

# Start service
sudo systemctl start rockstar-app

# Check status
sudo systemctl status rockstar-app

# View logs
sudo journalctl -u rockstar-app -f

# Restart service
sudo systemctl restart rockstar-app

# Stop service
sudo systemctl stop rockstar-app
```

### Unix-Specific Monitoring Commands

#### Real-Time Connection Monitoring

```bash
# Active connections
netstat -an | grep :8080 | wc -l

# Connection states
ss -s

# Detailed socket statistics
ss -tan state established '( dport = :8080 or sport = :8080 )'

# Watch connections in real-time
watch -n 1 'ss -tan state established | grep :8080 | wc -l'
```

#### Process Monitoring

```bash
# CPU and memory usage
top -p $(pgrep -f rockstar-app)

# Detailed process information
ps aux | grep rockstar-app

# File descriptors in use
lsof -p $(pgrep -f rockstar-app) | wc -l

# Network connections per process
lsof -i -P -n | grep rockstar-app
```

#### System Resource Monitoring

```bash
# I/O statistics
iostat -x 1

# Network interface statistics
sar -n DEV 1

# TCP statistics
nstat -az | grep -i tcp

# System load
uptime
```

#### Performance Profiling

```bash
# CPU profiling with perf
sudo perf record -F 99 -p $(pgrep -f rockstar-app) -g -- sleep 30
sudo perf report

# System call tracing
strace -c -p $(pgrep -f rockstar-app)

# Network packet capture
sudo tcpdump -i any port 8080 -w capture.pcap
```


## Windows Optimizations

### IOCP (I/O Completion Ports) Configuration

Windows uses IOCP for asynchronous I/O operations. The Go runtime automatically uses IOCP on Windows, but you can optimize its behavior.

#### Basic Configuration

```go
config := pkg.FrameworkConfig{
    Server: pkg.ServerConfig{
        EnableHTTP1: true,
        HTTP1Port:   8080,
        ListenerConfig: pkg.ListenerConfig{
            ReuseAddr:   true, // Windows SO_REUSEADDR behavior differs from Unix
            ReadBuffer:  65536,
            WriteBuffer: 65536,
        },
    },
}

app, err := pkg.New(config)
if err != nil {
    log.Fatal(err)
}
```

#### IOCP Thread Pool Sizing

The Go runtime manages IOCP threads automatically, but you can influence behavior:

```go
import (
    "os"
    "runtime"
    "strconv"
)

func init() {
    // Set GOMAXPROCS to number of CPU cores
    runtime.GOMAXPROCS(runtime.NumCPU())
    
    // For I/O-bound workloads, consider increasing
    // This allows more goroutines to run concurrently
    if os.Getenv("ROCKSTAR_IO_INTENSIVE") == "true" {
        runtime.GOMAXPROCS(runtime.NumCPU() * 2)
    }
}
```

### Thread Pool Tuning

Windows thread pool settings affect application performance.

#### Environment Variables

```powershell
# Set minimum worker threads
$env:GOMAXPROCS = "8"

# For high-concurrency scenarios
$env:GOMAXPROCS = "16"
```

#### Application Configuration

```go
import (
    "runtime"
    "runtime/debug"
)

func configureWindowsThreadPool() {
    // Set GOMAXPROCS
    numCPU := runtime.NumCPU()
    runtime.GOMAXPROCS(numCPU)
    
    // Configure GC for server workloads
    debug.SetGCPercent(100) // Default is 100
    
    // For memory-intensive applications
    debug.SetMemoryLimit(8 * 1024 * 1024 * 1024) // 8GB limit
}
```

### Registry Settings for Network Optimization

Optimize Windows network stack through registry settings.

#### TCP/IP Parameters

Create a PowerShell script to apply settings:

```powershell
# Run as Administrator

# Increase TCP window size
Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters" `
    -Name "TcpWindowSize" -Value 65535 -Type DWord

# Enable TCP window scaling
Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters" `
    -Name "Tcp1323Opts" -Value 3 -Type DWord

# Increase maximum number of connections
Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters" `
    -Name "TcpNumConnections" -Value 16777214 -Type DWord

# Reduce TIME_WAIT timeout
Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters" `
    -Name "TcpTimedWaitDelay" -Value 30 -Type DWord

# Increase dynamic port range
netsh int ipv4 set dynamicport tcp start=10000 num=55535
netsh int ipv6 set dynamicport tcp start=10000 num=55535

# Enable RSS (Receive Side Scaling)
Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters" `
    -Name "EnableRSS" -Value 1 -Type DWord

# Restart required for changes to take effect
Write-Host "Registry settings applied. Restart required."
```

#### Verify Registry Settings

```powershell
# Check TCP window size
Get-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters" `
    -Name "TcpWindowSize"

# Check dynamic port range
netsh int ipv4 show dynamicport tcp
```

#### Network Adapter Settings

```powershell
# Disable TCP offloading (if experiencing issues)
Disable-NetAdapterChecksumOffload -Name "*"
Disable-NetAdapterLso -Name "*"

# Enable Jumbo Frames (for 10GbE networks)
Set-NetAdapterAdvancedProperty -Name "*" -DisplayName "Jumbo Packet" -DisplayValue "9014"

# Configure RSS
Set-NetAdapterRss -Name "*" -Enabled $true -NumberOfReceiveQueues 4
```

### Windows Service Configuration

Create a production-ready Windows service for your application.

#### Service Installation Script

Create `install-service.ps1`:

```powershell
# Run as Administrator

$serviceName = "RockstarWebApp"
$displayName = "Rockstar Web Application"
$description = "High-performance web application using Rockstar Framework"
$binaryPath = "C:\Program Files\RockstarApp\server.exe"
$startupType = "Automatic"

# Create service
New-Service -Name $serviceName `
    -BinaryPathName $binaryPath `
    -DisplayName $displayName `
    -Description $description `
    -StartupType $startupType `
    -DependsOn "Tcpip"

# Configure service recovery options
sc.exe failure $serviceName reset= 86400 actions= restart/60000/restart/60000/restart/60000

# Set service to run as Network Service (more secure than Local System)
$service = Get-WmiObject -Class Win32_Service -Filter "Name='$serviceName'"
$service.Change($null, $null, $null, $null, $null, $null, "NT AUTHORITY\NetworkService", $null)

# Start service
Start-Service -Name $serviceName

Write-Host "Service installed and started successfully"
```

#### Service Management

```powershell
# Start service
Start-Service -Name "RockstarWebApp"

# Stop service
Stop-Service -Name "RockstarWebApp"

# Restart service
Restart-Service -Name "RockstarWebApp"

# Check service status
Get-Service -Name "RockstarWebApp"

# View service configuration
Get-WmiObject -Class Win32_Service -Filter "Name='RockstarWebApp'" | Format-List *

# Uninstall service
Stop-Service -Name "RockstarWebApp"
sc.exe delete "RockstarWebApp"
```

#### Service Configuration File

Create `service-config.json`:

```json
{
    "service": {
        "name": "RockstarWebApp",
        "display_name": "Rockstar Web Application",
        "description": "High-performance web application",
        "startup_type": "automatic",
        "recovery": {
            "reset_period": 86400,
            "actions": [
                {"type": "restart", "delay": 60000},
                {"type": "restart", "delay": 60000},
                {"type": "restart", "delay": 60000}
            ]
        }
    },
    "logging": {
        "event_log": true,
        "file_log": "C:\\ProgramData\\RockstarApp\\logs\\app.log"
    }
}
```

### Windows-Specific Monitoring Tools

#### Performance Monitor (PerfMon)

Monitor application performance using PerfMon:

```powershell
# Open Performance Monitor
perfmon

# Create custom counter set
$counterSet = @(
    "\Process(server)\% Processor Time",
    "\Process(server)\Private Bytes",
    "\Process(server)\Handle Count",
    "\Process(server)\Thread Count",
    "\TCPv4\Connections Established",
    "\Network Interface(*)\Bytes Total/sec"
)

# Collect performance data
Get-Counter -Counter $counterSet -SampleInterval 1 -MaxSamples 60
```

#### Event Tracing for Windows (ETW)

Use ETW for detailed performance analysis:

```powershell
# Start ETW trace
logman create trace RockstarTrace -o C:\traces\rockstar.etl -p Microsoft-Windows-TCPIP -ets

# Run your workload...

# Stop trace
logman stop RockstarTrace -ets

# Analyze trace with Windows Performance Analyzer
# Or use tracerpt
tracerpt C:\traces\rockstar.etl -o report.xml -summary summary.txt
```

#### Resource Monitor

```powershell
# Open Resource Monitor
resmon

# Or use PowerShell to get resource information
Get-Process -Name "server" | Select-Object Name, CPU, WorkingSet, Handles, Threads
```

#### Network Statistics

```powershell
# Active connections
netstat -an | Select-String ":8080" | Measure-Object

# Connection states
netstat -s

# TCP statistics
Get-NetTCPConnection -State Established | Where-Object {$_.LocalPort -eq 8080}

# Real-time monitoring
while ($true) {
    Clear-Host
    $connections = Get-NetTCPConnection -State Established | Where-Object {$_.LocalPort -eq 8080}
    Write-Host "Active connections: $($connections.Count)"
    Start-Sleep -Seconds 1
}
```

#### Windows Performance Toolkit

```powershell
# Install Windows Performance Toolkit (part of Windows SDK)
# Download from: https://developer.microsoft.com/en-us/windows/downloads/windows-sdk/

# Record performance trace
wpr -start GeneralProfile -start Network

# Run your workload...

# Stop recording
wpr -stop C:\traces\performance.etl

# Analyze with Windows Performance Analyzer
wpa C:\traces\performance.etl
```

#### Application Insights

For production monitoring, integrate with Application Insights:

```go
import (
    "github.com/microsoft/ApplicationInsights-Go/appinsights"
)

func setupApplicationInsights() appinsights.TelemetryClient {
    client := appinsights.NewTelemetryClient("YOUR_INSTRUMENTATION_KEY")
    
    // Track custom metrics
    client.TrackMetric("requests_per_second", 100.0)
    
    // Track events
    client.TrackEvent("server_started")
    
    return client
}
```


## Platform-Specific Listener Differences

### Unix vs Windows Listener Implementations

The framework provides platform-specific listener implementations that leverage OS-specific features.

#### Architecture Comparison

| Aspect | Unix/Linux | Windows |
|--------|-----------|---------|
| Process Model | fork() | CreateProcess() |
| Socket Sharing | SO_REUSEPORT | SO_REUSEADDR (limited) |
| I/O Model | epoll/kqueue | IOCP |
| Signal Handling | POSIX signals | Windows events |
| File Descriptors | Unix FD | Windows HANDLE |

#### Unix Listener Implementation

```go
// Unix-specific listener with SO_REUSEPORT
config := pkg.ListenerConfig{
    Network:        "tcp",
    Address:        ":8080",
    EnablePrefork:  true,
    PreforkWorkers: 4,
    ReusePort:      true, // Linux/BSD feature
    ReuseAddr:      true,
    ReadBuffer:     65536,
    WriteBuffer:    65536,
}

listener, err := pkg.CreateListener(config)
if err != nil {
    log.Fatal(err)
}
```

**Key Features:**
- Uses `fork()` for prefork mode
- Supports `SO_REUSEPORT` for load balancing across workers
- Uses `epoll` (Linux) or `kqueue` (BSD) for efficient I/O
- Inherits file descriptors across fork

#### Windows Listener Implementation

```go
// Windows-specific listener
config := pkg.ListenerConfig{
    Network:        "tcp",
    Address:        ":8080",
    EnablePrefork:  true,
    PreforkWorkers: 4,
    ReuseAddr:      true, // Windows behavior differs
    ReadBuffer:     65536,
    WriteBuffer:    65536,
}

listener, err := pkg.CreateListener(config)
if err != nil {
    log.Fatal(err)
}
```

**Key Features:**
- Uses `CreateProcess()` for prefork mode
- `SO_REUSEADDR` allows binding to same port (different semantics than Unix)
- Uses IOCP for asynchronous I/O
- Creates new process group for workers

### Platform-Specific Configuration Options

#### Detecting Platform Capabilities

```go
import "github.com/echterhof/rockstar-web-framework/pkg"

func configurePlatformListener() pkg.ListenerConfig {
    info := pkg.GetPlatformInfo()
    
    config := pkg.ListenerConfig{
        Network:       "tcp",
        Address:       ":8080",
        EnablePrefork: info.SupportsPrefork,
        ReuseAddr:     true,
    }
    
    // Enable SO_REUSEPORT only on platforms that support it
    if info.SupportsReusePort {
        config.ReusePort = true
        config.PreforkWorkers = info.NumCPU
    } else {
        // Windows: Use fewer workers without SO_REUSEPORT
        config.PreforkWorkers = info.NumCPU / 2
    }
    
    return config
}
```

#### Cross-Platform Configuration

```go
import (
    "runtime"
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func createCrossPlatformListener() (pkg.ListenerConfig, error) {
    config := pkg.ListenerConfig{
        Network:    "tcp",
        Address:    ":8080",
        ReuseAddr:  true,
        ReadBuffer: 65536,
        WriteBuffer: 65536,
    }
    
    switch runtime.GOOS {
    case "linux", "darwin", "freebsd":
        // Unix-like systems: Enable prefork with SO_REUSEPORT
        config.EnablePrefork = true
        config.ReusePort = true
        config.PreforkWorkers = runtime.NumCPU()
        
    case "windows":
        // Windows: Enable prefork without SO_REUSEPORT
        config.EnablePrefork = true
        config.PreforkWorkers = runtime.NumCPU() / 2
        
    default:
        // Other platforms: Disable prefork
        config.EnablePrefork = false
    }
    
    return config, nil
}
```

### Signal Handling Differences

#### Unix Signal Handling

Unix systems use POSIX signals for process management:

```go
import (
    "context"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func setupUnixSignalHandling(app *pkg.Framework) {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, 
        syscall.SIGINT,  // Ctrl+C
        syscall.SIGTERM, // Termination signal
        syscall.SIGHUP,  // Reload configuration
        syscall.SIGUSR1, // Custom signal 1
        syscall.SIGUSR2, // Custom signal 2
    )
    
    go func() {
        for sig := range sigChan {
            switch sig {
            case syscall.SIGINT, syscall.SIGTERM:
                // Graceful shutdown
                log.Println("Received shutdown signal:", sig)
                ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
                defer cancel()
                
                if err := app.Shutdown(ctx); err != nil {
                    log.Printf("Shutdown error: %v", err)
                }
                os.Exit(0)
                
            case syscall.SIGHUP:
                // Reload configuration
                log.Println("Received reload signal")
                // Implement configuration reload logic
                
            case syscall.SIGUSR1:
                // Custom action (e.g., reopen log files)
                log.Println("Received SIGUSR1")
                
            case syscall.SIGUSR2:
                // Custom action (e.g., dump statistics)
                log.Println("Received SIGUSR2")
            }
        }
    }()
}
```

#### Windows Event Handling

Windows uses console events and service control messages:

```go
import (
    "context"
    "os"
    "os/signal"
    "time"
)

func setupWindowsEventHandling(app *pkg.Framework) {
    sigChan := make(chan os.Signal, 1)
    
    // Windows supports limited signals
    signal.Notify(sigChan,
        os.Interrupt, // Ctrl+C
        os.Kill,      // Termination
    )
    
    go func() {
        sig := <-sigChan
        log.Println("Received shutdown signal:", sig)
        
        // Graceful shutdown
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        
        if err := app.Shutdown(ctx); err != nil {
            log.Printf("Shutdown error: %v", err)
        }
        os.Exit(0)
    }()
}
```

#### Cross-Platform Signal Handling

```go
import (
    "context"
    "os"
    "os/signal"
    "runtime"
    "syscall"
    "time"
)

func setupCrossPlatformSignalHandling(app *pkg.Framework) {
    sigChan := make(chan os.Signal, 1)
    
    if runtime.GOOS == "windows" {
        // Windows: Limited signal support
        signal.Notify(sigChan, os.Interrupt, os.Kill)
    } else {
        // Unix: Full POSIX signal support
        signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
    }
    
    go func() {
        sig := <-sigChan
        log.Printf("Received signal: %v", sig)
        
        // Graceful shutdown with timeout
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        
        if err := app.Shutdown(ctx); err != nil {
            log.Printf("Shutdown error: %v", err)
            os.Exit(1)
        }
        os.Exit(0)
    }()
}
```

### Graceful Shutdown Patterns Per Platform

#### Unix Graceful Shutdown

```go
func unixGracefulShutdown(app *pkg.Framework) {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
    
    <-sigChan
    log.Println("Initiating graceful shutdown...")
    
    // Create shutdown context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Register shutdown hooks
    app.RegisterShutdownHook(func(ctx context.Context) error {
        log.Println("Closing database connections...")
        // Close database connections
        return nil
    })
    
    app.RegisterShutdownHook(func(ctx context.Context) error {
        log.Println("Flushing caches...")
        // Flush caches
        return nil
    })
    
    // Perform shutdown
    if err := app.Shutdown(ctx); err != nil {
        log.Printf("Shutdown error: %v", err)
        os.Exit(1)
    }
    
    log.Println("Shutdown complete")
}
```

#### Windows Graceful Shutdown

```go
func windowsGracefulShutdown(app *pkg.Framework) {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt)
    
    <-sigChan
    log.Println("Initiating graceful shutdown...")
    
    // Windows: Shorter timeout due to service control manager limits
    ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
    defer cancel()
    
    // Register shutdown hooks
    app.RegisterShutdownHook(func(ctx context.Context) error {
        log.Println("Closing database connections...")
        return nil
    })
    
    // Perform shutdown
    if err := app.Shutdown(ctx); err != nil {
        log.Printf("Shutdown error: %v", err)
        os.Exit(1)
    }
    
    log.Println("Shutdown complete")
}
```

### Platform Detection and Adaptation Patterns

#### Runtime Platform Detection

```go
import (
    "runtime"
    "github.com/echterhof/rockstar-web-framework/pkg"
)

type PlatformAdapter struct {
    OS           string
    Arch         string
    Config       pkg.FrameworkConfig
    ShutdownFunc func(*pkg.Framework)
}

func NewPlatformAdapter() *PlatformAdapter {
    adapter := &PlatformAdapter{
        OS:   runtime.GOOS,
        Arch: runtime.GOARCH,
    }
    
    // Configure based on platform
    adapter.Config = adapter.getPlatformConfig()
    adapter.ShutdownFunc = adapter.getShutdownHandler()
    
    return adapter
}

func (pa *PlatformAdapter) getPlatformConfig() pkg.FrameworkConfig {
    baseConfig := pkg.FrameworkConfig{
        Server: pkg.ServerConfig{
            EnableHTTP1: true,
            HTTP1Port:   8080,
        },
    }
    
    switch pa.OS {
    case "linux":
        baseConfig.Server.ListenerConfig = pkg.ListenerConfig{
            EnablePrefork:  true,
            PreforkWorkers: runtime.NumCPU(),
            ReusePort:      true,
            ReuseAddr:      true,
            ReadBuffer:     131072, // 128KB
            WriteBuffer:    131072,
        }
        
    case "windows":
        baseConfig.Server.ListenerConfig = pkg.ListenerConfig{
            EnablePrefork:  true,
            PreforkWorkers: runtime.NumCPU() / 2,
            ReuseAddr:      true,
            ReadBuffer:     65536, // 64KB
            WriteBuffer:    65536,
        }
        
    case "darwin":
        baseConfig.Server.ListenerConfig = pkg.ListenerConfig{
            EnablePrefork:  true,
            PreforkWorkers: runtime.NumCPU(),
            ReusePort:      true,
            ReuseAddr:      true,
            ReadBuffer:     65536,
            WriteBuffer:    65536,
        }
        
    default:
        // Conservative defaults for unknown platforms
        baseConfig.Server.ListenerConfig = pkg.ListenerConfig{
            EnablePrefork: false,
            ReuseAddr:     true,
            ReadBuffer:    32768,
            WriteBuffer:   32768,
        }
    }
    
    return baseConfig
}

func (pa *PlatformAdapter) getShutdownHandler() func(*pkg.Framework) {
    switch pa.OS {
    case "windows":
        return windowsGracefulShutdown
    default:
        return unixGracefulShutdown
    }
}
```

#### Build Tags for Platform-Specific Code

Use build tags to include platform-specific code:

```go
// +build linux darwin

package main

import "syscall"

func setPlatformOptions() {
    // Unix-specific code
    syscall.Setrlimit(syscall.RLIMIT_NOFILE, &syscall.Rlimit{
        Cur: 65536,
        Max: 65536,
    })
}
```

```go
// +build windows

package main

func setPlatformOptions() {
    // Windows-specific code
    // Configure Windows-specific settings
}
```


## Platform-Specific Network Tuning

### OS-Level TCP Tuning for Linux

#### Comprehensive TCP Optimization

Create `/etc/sysctl.d/99-rockstar-tuning.conf`:

```bash
# TCP Buffer Tuning
# Format: min default max (in bytes)
net.ipv4.tcp_rmem = 4096 87380 16777216
net.ipv4.tcp_wmem = 4096 65536 16777216
net.core.rmem_default = 262144
net.core.wmem_default = 262144
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216

# Connection Queue Tuning
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 8192
net.core.netdev_max_backlog = 16384

# TCP Fast Open (TFO)
# 1 = client, 2 = server, 3 = both
net.ipv4.tcp_fastopen = 3

# TCP Keepalive
net.ipv4.tcp_keepalive_time = 600
net.ipv4.tcp_keepalive_intvl = 60
net.ipv4.tcp_keepalive_probes = 3

# TIME_WAIT Socket Reuse
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_fin_timeout = 30

# Port Range
net.ipv4.ip_local_port_range = 10000 65535

# TCP Congestion Control
net.core.default_qdisc = fq
net.ipv4.tcp_congestion_control = bbr

# Disable TCP Slow Start After Idle
net.ipv4.tcp_slow_start_after_idle = 0

# TCP Window Scaling
net.ipv4.tcp_window_scaling = 1

# TCP Timestamps
net.ipv4.tcp_timestamps = 1

# Selective Acknowledgments
net.ipv4.tcp_sack = 1

# Forward Acknowledgment
net.ipv4.tcp_fack = 1

# MTU Probing
net.ipv4.tcp_mtu_probing = 1

# SYN Cookies (DDoS protection)
net.ipv4.tcp_syncookies = 1
net.ipv4.tcp_syn_retries = 2
net.ipv4.tcp_synack_retries = 2

# Memory Limits
net.ipv4.tcp_mem = 786432 1048576 26777216
net.ipv4.udp_mem = 786432 1048576 26777216

# Increase max orphaned sockets
net.ipv4.tcp_max_orphans = 65536

# Increase max TIME_WAIT sockets
net.ipv4.tcp_max_tw_buckets = 2000000
```

Apply settings:

```bash
sudo sysctl -p /etc/sysctl.d/99-rockstar-tuning.conf
```

#### Workload-Specific Linux Tuning

**High-Throughput API Server:**

```bash
# Optimize for throughput
net.ipv4.tcp_congestion_control = bbr
net.core.default_qdisc = fq
net.ipv4.tcp_notsent_lowat = 16384

# Large buffers
net.ipv4.tcp_rmem = 8192 262144 33554432
net.ipv4.tcp_wmem = 8192 262144 33554432
```

**Low-Latency WebSocket Server:**

```bash
# Optimize for latency
net.ipv4.tcp_congestion_control = cubic
net.ipv4.tcp_low_latency = 1
net.ipv4.tcp_notsent_lowat = 4096

# Smaller buffers, faster response
net.ipv4.tcp_rmem = 4096 65536 4194304
net.ipv4.tcp_wmem = 4096 65536 4194304
```

**High-Connection-Count Server:**

```bash
# Optimize for many connections
net.core.somaxconn = 131072
net.ipv4.tcp_max_syn_backlog = 16384
net.ipv4.tcp_max_orphans = 131072
net.ipv4.tcp_max_tw_buckets = 4000000

# Aggressive connection reuse
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_fin_timeout = 15
```

#### Network Interface Tuning

```bash
# Check current settings
ethtool -g eth0
ethtool -k eth0

# Increase ring buffer size
ethtool -G eth0 rx 4096 tx 4096

# Enable hardware offloading
ethtool -K eth0 tso on
ethtool -K eth0 gso on
ethtool -K eth0 gro on

# Enable receive packet steering (RPS)
echo "f" > /sys/class/net/eth0/queues/rx-0/rps_cpus

# Enable receive flow steering (RFS)
echo 32768 > /proc/sys/net/core/rps_sock_flow_entries
echo 2048 > /sys/class/net/eth0/queues/rx-0/rps_flow_cnt
```

### Windows Network Stack Optimization

#### TCP/IP Registry Tuning

Create `windows-network-tuning.ps1`:

```powershell
# Run as Administrator

# TCP Global Parameters
Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters" `
    -Name "TcpWindowSize" -Value 65535 -Type DWord

Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters" `
    -Name "Tcp1323Opts" -Value 3 -Type DWord

Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters" `
    -Name "DefaultTTL" -Value 64 -Type DWord

Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters" `
    -Name "EnablePMTUDiscovery" -Value 1 -Type DWord

Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters" `
    -Name "TcpMaxDataRetransmissions" -Value 5 -Type DWord

Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters" `
    -Name "TcpTimedWaitDelay" -Value 30 -Type DWord

Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters" `
    -Name "TcpNumConnections" -Value 16777214 -Type DWord

# Enable TCP Fast Open
Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters" `
    -Name "EnableTcpFastOpen" -Value 1 -Type DWord

# Dynamic Port Range
netsh int ipv4 set dynamicport tcp start=10000 num=55535
netsh int ipv6 set dynamicport tcp start=10000 num=55535

# TCP Chimney Offload (deprecated in newer Windows, but safe to set)
netsh int tcp set global chimney=enabled

# Receive-Side Scaling (RSS)
netsh int tcp set global rss=enabled
Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters" `
    -Name "EnableRSS" -Value 1 -Type DWord

# Receive Window Auto-Tuning
netsh int tcp set global autotuninglevel=normal

# ECN (Explicit Congestion Notification)
netsh int tcp set global ecncapability=enabled

# Timestamps
netsh int tcp set global timestamps=enabled

Write-Host "Network tuning applied. Restart required for all changes to take effect."
```

#### Network Adapter Advanced Settings

```powershell
# Get network adapter name
$adapter = Get-NetAdapter | Where-Object {$_.Status -eq "Up"} | Select-Object -First 1
$adapterName = $adapter.Name

# Interrupt Moderation
Set-NetAdapterAdvancedProperty -Name $adapterName -DisplayName "Interrupt Moderation" -DisplayValue "Enabled"

# Receive Buffers
Set-NetAdapterAdvancedProperty -Name $adapterName -DisplayName "Receive Buffers" -DisplayValue "2048"

# Transmit Buffers
Set-NetAdapterAdvancedProperty -Name $adapterName -DisplayName "Transmit Buffers" -DisplayValue "2048"

# RSS Settings
Set-NetAdapterRss -Name $adapterName -Enabled $true -NumberOfReceiveQueues 4

# Large Send Offload (LSO)
Enable-NetAdapterLso -Name $adapterName

# Checksum Offload
Enable-NetAdapterChecksumOffload -Name $adapterName

# Jumbo Frames (for 10GbE networks)
Set-NetAdapterAdvancedProperty -Name $adapterName -DisplayName "Jumbo Packet" -DisplayValue "9014"
```

#### Windows Firewall Optimization

```powershell
# Optimize firewall for high-performance applications
Set-NetFirewallProfile -Profile Domain,Public,Private -Enabled True

# Create high-performance rule for your application
New-NetFirewallRule -DisplayName "Rockstar Web App" `
    -Direction Inbound `
    -Protocol TCP `
    -LocalPort 8080 `
    -Action Allow `
    -Profile Any `
    -Enabled True

# Disable firewall logging for performance (if not needed)
Set-NetFirewallProfile -Profile Domain,Public,Private -LogBlocked False -LogAllowed False
```

### Kernel Parameter Tuning

#### Linux Kernel Parameters

```bash
# Memory Management
vm.swappiness = 10
vm.dirty_ratio = 15
vm.dirty_background_ratio = 5

# File System
fs.file-max = 2097152
fs.nr_open = 2097152

# Network Core
net.core.netdev_budget = 600
net.core.netdev_budget_usecs = 5000

# IPv4 Settings
net.ipv4.ip_forward = 0
net.ipv4.conf.all.rp_filter = 1
net.ipv4.conf.default.rp_filter = 1
net.ipv4.icmp_echo_ignore_broadcasts = 1
net.ipv4.icmp_ignore_bogus_error_responses = 1

# IPv6 (disable if not used)
net.ipv6.conf.all.disable_ipv6 = 1
net.ipv6.conf.default.disable_ipv6 = 1
```

#### Windows Kernel Tuning

```powershell
# Network Driver Interface Specification (NDIS) Settings
Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\NDIS\Parameters" `
    -Name "MaxNumRssCpus" -Value 8 -Type DWord

# TCP/IP Performance Settings
Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters" `
    -Name "MaxUserPort" -Value 65534 -Type DWord

Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters" `
    -Name "TcpMaxConnectRetransmissions" -Value 3 -Type DWord
```

### Network Buffer Sizing Guidance

#### Calculating Optimal Buffer Sizes

**Bandwidth-Delay Product (BDP):**

```
BDP = Bandwidth × RTT
```

Example calculations:

```go
import (
    "fmt"
    "time"
)

func calculateOptimalBufferSize(bandwidthMbps float64, rttMs time.Duration) int {
    // Convert bandwidth to bytes per second
    bandwidthBps := bandwidthMbps * 1000000 / 8
    
    // Convert RTT to seconds
    rttSeconds := float64(rttMs) / 1000.0
    
    // Calculate BDP
    bdp := int(bandwidthBps * rttSeconds)
    
    return bdp
}

func main() {
    // Example: 1 Gbps link with 10ms RTT
    bufferSize := calculateOptimalBufferSize(1000, 10*time.Millisecond)
    fmt.Printf("Optimal buffer size: %d bytes (%.2f KB)\n", 
        bufferSize, float64(bufferSize)/1024)
    
    // Example: 10 Gbps link with 1ms RTT
    bufferSize = calculateOptimalBufferSize(10000, 1*time.Millisecond)
    fmt.Printf("Optimal buffer size: %d bytes (%.2f KB)\n", 
        bufferSize, float64(bufferSize)/1024)
}
```

#### Buffer Size Recommendations

| Network Type | RTT | Bandwidth | Recommended Buffer |
|-------------|-----|-----------|-------------------|
| Local (LAN) | 1ms | 1 Gbps | 128 KB |
| Local (LAN) | 1ms | 10 Gbps | 1.25 MB |
| Regional | 10ms | 1 Gbps | 1.25 MB |
| Regional | 10ms | 100 Mbps | 128 KB |
| Internet | 50ms | 100 Mbps | 625 KB |
| Internet | 100ms | 10 Mbps | 128 KB |

#### Application-Level Buffer Configuration

```go
config := pkg.FrameworkConfig{
    Server: pkg.ServerConfig{
        EnableHTTP1: true,
        HTTP1Port:   8080,
        ListenerConfig: pkg.ListenerConfig{
            // LAN environment: 1 Gbps, 1ms RTT
            ReadBuffer:  131072,  // 128 KB
            WriteBuffer: 131072,  // 128 KB
            
            // For high-bandwidth WAN: 1 Gbps, 10ms RTT
            // ReadBuffer:  1310720,  // 1.25 MB
            // WriteBuffer: 1310720,  // 1.25 MB
        },
    },
}
```

### Platform-Specific Benchmarking Approaches

#### Linux Benchmarking

**Using `wrk` for HTTP benchmarking:**

```bash
# Install wrk
git clone https://github.com/wg/wrk.git
cd wrk
make
sudo cp wrk /usr/local/bin/

# Basic benchmark
wrk -t12 -c400 -d30s http://localhost:8080/

# With custom script
wrk -t12 -c400 -d30s -s script.lua http://localhost:8080/

# Monitor during benchmark
watch -n 1 'ss -s; echo ""; netstat -s | grep -i retrans'
```

**Using `ab` (Apache Bench):**

```bash
# Install
sudo apt-get install apache2-utils

# Run benchmark
ab -n 100000 -c 100 http://localhost:8080/

# With keepalive
ab -n 100000 -c 100 -k http://localhost:8080/
```

**Using `hey`:**

```bash
# Install
go install github.com/rakyll/hey@latest

# Run benchmark
hey -n 100000 -c 100 http://localhost:8080/

# With custom headers
hey -n 100000 -c 100 -H "Authorization: Bearer token" http://localhost:8080/
```

**Network throughput testing:**

```bash
# Install iperf3
sudo apt-get install iperf3

# Server side
iperf3 -s

# Client side
iperf3 -c server-ip -t 30 -P 10
```

#### Windows Benchmarking

**Using PowerShell for HTTP benchmarking:**

```powershell
# Simple benchmark function
function Invoke-HttpBenchmark {
    param(
        [string]$Url,
        [int]$Requests = 1000,
        [int]$Concurrent = 10
    )
    
    $jobs = @()
    $requestsPerJob = [math]::Floor($Requests / $Concurrent)
    
    1..$Concurrent | ForEach-Object {
        $jobs += Start-Job -ScriptBlock {
            param($url, $count)
            $results = @()
            
            for ($i = 0; $i -lt $count; $i++) {
                $start = Get-Date
                try {
                    $response = Invoke-WebRequest -Uri $url -UseBasicParsing
                    $duration = (Get-Date) - $start
                    $results += @{
                        Success = $true
                        Duration = $duration.TotalMilliseconds
                        StatusCode = $response.StatusCode
                    }
                } catch {
                    $duration = (Get-Date) - $start
                    $results += @{
                        Success = $false
                        Duration = $duration.TotalMilliseconds
                        Error = $_.Exception.Message
                    }
                }
            }
            
            return $results
        } -ArgumentList $Url, $requestsPerJob
    }
    
    $allResults = $jobs | Wait-Job | Receive-Job
    $jobs | Remove-Job
    
    # Calculate statistics
    $successful = ($allResults | Where-Object {$_.Success}).Count
    $failed = ($allResults | Where-Object {-not $_.Success}).Count
    $durations = $allResults | ForEach-Object {$_.Duration}
    
    [PSCustomObject]@{
        TotalRequests = $Requests
        Successful = $successful
        Failed = $failed
        AvgDuration = ($durations | Measure-Object -Average).Average
        MinDuration = ($durations | Measure-Object -Minimum).Minimum
        MaxDuration = ($durations | Measure-Object -Maximum).Maximum
    }
}

# Run benchmark
Invoke-HttpBenchmark -Url "http://localhost:8080/" -Requests 10000 -Concurrent 50
```

**Using Windows Performance Toolkit:**

```powershell
# Start network trace
netsh trace start capture=yes tracefile=C:\traces\network.etl

# Run your benchmark...

# Stop trace
netsh trace stop

# Convert to text format
netsh trace convert C:\traces\network.etl
```

**Using PerfMon for monitoring:**

```powershell
# Create performance counter collector
$counterSet = @(
    "\Network Interface(*)\Bytes Total/sec",
    "\Network Interface(*)\Packets/sec",
    "\TCPv4\Connections Established",
    "\TCPv4\Segments Sent/sec",
    "\TCPv4\Segments Received/sec",
    "\Processor(_Total)\% Processor Time",
    "\Memory\Available MBytes"
)

# Collect data during benchmark
Get-Counter -Counter $counterSet -SampleInterval 1 -MaxSamples 60 | 
    Export-Counter -Path C:\traces\performance.blg -FileFormat BLG
```


## Platform-Specific Diagnostics

### Linux Diagnostic Tools

#### strace - System Call Tracing

Monitor system calls made by your application:

```bash
# Trace all system calls
strace -p $(pgrep -f rockstar-app)

# Count system calls
strace -c -p $(pgrep -f rockstar-app)

# Trace specific system calls
strace -e trace=network -p $(pgrep -f rockstar-app)

# Trace file operations
strace -e trace=file -p $(pgrep -f rockstar-app)

# Save trace to file
strace -o trace.log -p $(pgrep -f rockstar-app)

# Trace with timestamps
strace -tt -p $(pgrep -f rockstar-app)

# Trace new connections
strace -e trace=socket,connect,accept -p $(pgrep -f rockstar-app)
```

**Common Issues Detected:**
- Excessive system calls (performance bottleneck)
- Failed file operations (permission issues)
- Network connection failures
- Slow DNS lookups

#### perf - Performance Analysis

CPU profiling and performance analysis:

```bash
# Record CPU profile (30 seconds)
sudo perf record -F 99 -p $(pgrep -f rockstar-app) -g -- sleep 30

# View report
sudo perf report

# Record with call graphs
sudo perf record -F 99 -p $(pgrep -f rockstar-app) -g --call-graph dwarf -- sleep 30

# Top functions consuming CPU
sudo perf top -p $(pgrep -f rockstar-app)

# Cache misses
sudo perf stat -e cache-misses,cache-references -p $(pgrep -f rockstar-app) -- sleep 10

# Context switches
sudo perf stat -e context-switches -p $(pgrep -f rockstar-app) -- sleep 10

# Memory access patterns
sudo perf mem record -p $(pgrep -f rockstar-app) -- sleep 10
sudo perf mem report
```

**Flame Graph Generation:**

```bash
# Install FlameGraph tools
git clone https://github.com/brendangregg/FlameGraph
cd FlameGraph

# Record and generate flame graph
sudo perf record -F 99 -p $(pgrep -f rockstar-app) -g -- sleep 30
sudo perf script | ./stackcollapse-perf.pl | ./flamegraph.pl > flamegraph.svg
```

#### netstat/ss - Network Statistics

Monitor network connections and statistics:

```bash
# Active connections
netstat -an | grep :8080

# Connection states
ss -tan state established

# Listening ports
ss -tlnp

# Connection statistics
ss -s

# TCP statistics
netstat -st

# Watch connections in real-time
watch -n 1 'ss -tan state established | grep :8080 | wc -l'

# Detailed socket information
ss -tiepm

# Show process information
ss -tlnp | grep :8080
```

#### tcpdump - Packet Capture

Capture and analyze network traffic:

```bash
# Capture on specific port
sudo tcpdump -i any port 8080

# Save to file
sudo tcpdump -i any port 8080 -w capture.pcap

# Capture with timestamps
sudo tcpdump -i any port 8080 -tttt

# Capture HTTP traffic
sudo tcpdump -i any port 8080 -A

# Capture specific number of packets
sudo tcpdump -i any port 8080 -c 1000

# Filter by source/destination
sudo tcpdump -i any src 192.168.1.100 and port 8080

# Analyze captured file
tcpdump -r capture.pcap
```

#### lsof - List Open Files

Monitor file descriptors and network connections:

```bash
# All open files for process
lsof -p $(pgrep -f rockstar-app)

# Network connections only
lsof -i -P -n -p $(pgrep -f rockstar-app)

# Count open file descriptors
lsof -p $(pgrep -f rockstar-app) | wc -l

# TCP connections
lsof -i TCP -p $(pgrep -f rockstar-app)

# Listening sockets
lsof -i -sTCP:LISTEN -p $(pgrep -f rockstar-app)

# Watch file descriptor count
watch -n 1 'lsof -p $(pgrep -f rockstar-app) | wc -l'
```

#### iostat - I/O Statistics

Monitor disk I/O performance:

```bash
# Basic I/O stats
iostat -x 1

# Extended statistics
iostat -xz 1

# CPU and I/O stats
iostat -xz -c 1

# Specific device
iostat -x /dev/sda 1
```

#### vmstat - Virtual Memory Statistics

Monitor system resources:

```bash
# Basic stats (1 second intervals)
vmstat 1

# Memory statistics
vmstat -s

# Disk statistics
vmstat -d

# Active/inactive memory
vmstat -a 1
```

#### sar - System Activity Reporter

Comprehensive system monitoring:

```bash
# CPU usage
sar -u 1 10

# Memory usage
sar -r 1 10

# Network statistics
sar -n DEV 1 10

# TCP statistics
sar -n TCP 1 10

# Load average
sar -q 1 10

# Historical data
sar -f /var/log/sysstat/sa$(date +%d)
```

### Windows Diagnostic Approaches

#### Performance Monitor (PerfMon)

Comprehensive performance monitoring:

```powershell
# Open Performance Monitor
perfmon

# Create custom data collector set
$counterSet = @(
    "\Process(server)\% Processor Time",
    "\Process(server)\Private Bytes",
    "\Process(server)\Handle Count",
    "\Process(server)\Thread Count",
    "\Process(server)\IO Read Bytes/sec",
    "\Process(server)\IO Write Bytes/sec",
    "\TCPv4\Connections Established",
    "\TCPv4\Segments Sent/sec",
    "\TCPv4\Segments Received/sec",
    "\Network Interface(*)\Bytes Total/sec",
    "\Memory\Available MBytes",
    "\Memory\Pages/sec",
    "\PhysicalDisk(*)\% Disk Time",
    "\PhysicalDisk(*)\Avg. Disk Queue Length"
)

# Collect performance data
Get-Counter -Counter $counterSet -SampleInterval 1 -MaxSamples 3600 |
    Export-Counter -Path "C:\PerfLogs\performance.blg" -FileFormat BLG

# Real-time monitoring
Get-Counter -Counter $counterSet -SampleInterval 1 -Continuous
```

**Create Data Collector Set:**

```powershell
# Create new data collector set
$collectorSetName = "RockstarAppMonitoring"
$outputPath = "C:\PerfLogs\RockstarApp"

logman create counter $collectorSetName -o $outputPath -f bincirc -max 500 -c `
    "\Process(server)\% Processor Time" `
    "\Process(server)\Private Bytes" `
    "\TCPv4\Connections Established" `
    "\Network Interface(*)\Bytes Total/sec"

# Start collection
logman start $collectorSetName

# Stop collection
logman stop $collectorSetName

# Query status
logman query $collectorSetName
```

#### Event Tracing for Windows (ETW)

Low-overhead event tracing:

```powershell
# Start network trace
netsh trace start capture=yes report=yes tracefile=C:\traces\network.etl maxsize=1024

# Start with specific providers
netsh trace start capture=yes provider=Microsoft-Windows-TCPIP `
    provider=Microsoft-Windows-Winsock-AFD tracefile=C:\traces\detailed.etl

# Stop trace
netsh trace stop

# Convert to text
netsh trace convert C:\traces\network.etl

# View available providers
netsh trace show providers

# Custom ETW session
logman create trace NetworkTrace -o C:\traces\network.etl -p Microsoft-Windows-TCPIP -ets

# Stop custom session
logman stop NetworkTrace -ets
```

**Analyze ETW Traces:**

```powershell
# Using tracerpt
tracerpt C:\traces\network.etl -o report.xml -summary summary.txt

# Using Windows Performance Analyzer (WPA)
wpa C:\traces\network.etl
```

#### Resource Monitor (resmon)

Real-time resource monitoring:

```powershell
# Open Resource Monitor
resmon

# Or use PowerShell to get similar information
Get-Process -Name "server" | Select-Object `
    Name,
    @{Name="CPU(%)";Expression={$_.CPU}},
    @{Name="Memory(MB)";Expression={[math]::Round($_.WorkingSet64/1MB,2)}},
    @{Name="Handles";Expression={$_.HandleCount}},
    @{Name="Threads";Expression={$_.Threads.Count}}

# Monitor in real-time
while ($true) {
    Clear-Host
    Get-Process -Name "server" | Format-Table Name, CPU, `
        @{Name="Memory(MB)";Expression={[math]::Round($_.WorkingSet64/1MB,2)}}, `
        Handles, @{Name="Threads";Expression={$_.Threads.Count}}
    Start-Sleep -Seconds 1
}
```

#### Network Diagnostics

```powershell
# Active connections
Get-NetTCPConnection -State Established | Where-Object {$_.LocalPort -eq 8080}

# Connection statistics
Get-NetTCPConnection | Group-Object State | Select-Object Name, Count

# Network adapter statistics
Get-NetAdapterStatistics

# TCP statistics
Get-NetTCPStatistics

# Real-time connection monitoring
while ($true) {
    Clear-Host
    $connections = Get-NetTCPConnection -State Established | 
        Where-Object {$_.LocalPort -eq 8080}
    Write-Host "Active connections: $($connections.Count)"
    Write-Host "Timestamp: $(Get-Date)"
    $connections | Format-Table LocalAddress, LocalPort, RemoteAddress, RemotePort
    Start-Sleep -Seconds 1
}
```

#### Process Explorer

Advanced process monitoring (Sysinternals):

```powershell
# Download Process Explorer
# https://docs.microsoft.com/en-us/sysinternals/downloads/process-explorer

# Or use PowerShell for similar information
Get-Process -Name "server" | Select-Object `
    Id,
    ProcessName,
    @{Name="CPU(%)";Expression={$_.CPU}},
    @{Name="Memory(MB)";Expression={[math]::Round($_.WorkingSet64/1MB,2)}},
    @{Name="Handles";Expression={$_.HandleCount}},
    @{Name="Threads";Expression={$_.Threads.Count}},
    @{Name="StartTime";Expression={$_.StartTime}}
```

#### Windows Performance Recorder (WPR)

Record detailed performance traces:

```powershell
# Start recording
wpr -start GeneralProfile -start Network

# Run your workload...

# Stop recording
wpr -stop C:\traces\performance.etl

# Analyze with Windows Performance Analyzer
wpa C:\traces\performance.etl
```

### Platform-Specific Profiling Techniques

#### Linux CPU Profiling

**Using pprof with Go applications:**

```go
import (
    "net/http"
    _ "net/http/pprof"
)

func main() {
    // Enable pprof endpoint
    go func() {
        http.ListenAndServe("localhost:6060", nil)
    }()
    
    // Your application code...
}
```

**Collect and analyze profiles:**

```bash
# CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Heap profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine profile
go tool pprof http://localhost:6060/debug/pprof/goroutine

# Generate SVG
go tool pprof -svg http://localhost:6060/debug/pprof/profile > cpu.svg

# Interactive mode
go tool pprof -http=:8081 http://localhost:6060/debug/pprof/profile
```

#### Windows CPU Profiling

**Using Visual Studio Profiler:**

```powershell
# Attach to running process
VSPerfCmd /start:sample /attach:<PID> /output:profile.vsp

# Run workload...

# Stop profiling
VSPerfCmd /detach
VSPerfCmd /shutdown

# Analyze with Visual Studio
devenv profile.vsp
```

**Using Go pprof on Windows:**

```powershell
# Same as Linux, but use PowerShell
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Generate report
go tool pprof -text http://localhost:6060/debug/pprof/profile > profile.txt
```

### Troubleshooting Decision Trees

#### High CPU Usage

```
High CPU Usage Detected
│
├─ Check CPU profile
│  ├─ Hot function identified?
│  │  ├─ Yes → Optimize function
│  │  └─ No → Check goroutine count
│  │
│  └─ Too many goroutines?
│     ├─ Yes → Implement goroutine pooling
│     └─ No → Check system calls
│
├─ Excessive system calls?
│  ├─ Yes (strace/ETW shows high rate)
│  │  ├─ File I/O → Implement buffering
│  │  ├─ Network I/O → Batch operations
│  │  └─ Memory allocation → Use sync.Pool
│  │
│  └─ No → Check GC pressure
│
└─ High GC time?
   ├─ Yes → Reduce allocations, tune GOGC
   └─ No → Profile with perf/WPR for deeper analysis
```

#### High Memory Usage

```
High Memory Usage Detected
│
├─ Check heap profile
│  ├─ Memory leak identified?
│  │  ├─ Yes → Fix leak (unclosed resources, goroutine leaks)
│  │  └─ No → Check allocation patterns
│  │
│  └─ Excessive allocations?
│     ├─ Yes → Use sync.Pool, reduce allocations
│     └─ No → Check goroutine count
│
├─ Too many goroutines?
│  ├─ Yes → Implement goroutine limits
│  └─ No → Check cache sizes
│
└─ Large caches?
   ├─ Yes → Implement cache eviction, reduce cache size
   └─ No → Check for memory fragmentation
```

#### Network Performance Issues

```
Network Performance Issues
│
├─ Check connection count
│  ├─ Too many connections?
│  │  ├─ Yes → Implement connection pooling
│  │  └─ No → Check latency
│  │
│  └─ High latency?
│     ├─ Yes → Check network path (traceroute/tracert)
│     └─ No → Check throughput
│
├─ Low throughput?
│  ├─ Check TCP window size
│  │  ├─ Too small → Increase buffer sizes
│  │  └─ OK → Check for packet loss
│  │
│  └─ Packet loss detected?
│     ├─ Yes → Check network hardware, MTU settings
│     └─ No → Check application-level bottlenecks
│
└─ Application bottleneck?
   ├─ Slow request processing → Profile application
   ├─ Database slow → Optimize queries, add indexes
   └─ External API slow → Implement caching, timeouts
```

### Common Issue Resolution Guides

#### Issue: File Descriptor Limit Reached (Linux)

**Symptoms:**
- "too many open files" errors
- Connection failures
- Application crashes

**Diagnosis:**

```bash
# Check current usage
lsof -p $(pgrep -f rockstar-app) | wc -l

# Check limit
ulimit -n

# Check system-wide limit
cat /proc/sys/fs/file-max
```

**Resolution:**

```bash
# Temporary fix
ulimit -n 65536

# Permanent fix
echo "* soft nofile 65536" | sudo tee -a /etc/security/limits.conf
echo "* hard nofile 65536" | sudo tee -a /etc/security/limits.conf

# System-wide
echo "fs.file-max = 2097152" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

#### Issue: Port Exhaustion (Windows/Linux)

**Symptoms:**
- "cannot assign requested address" errors
- Connection failures after many requests
- TIME_WAIT sockets accumulating

**Diagnosis:**

```bash
# Linux
ss -tan | grep TIME_WAIT | wc -l
cat /proc/sys/net/ipv4/ip_local_port_range

# Windows
netstat -an | findstr TIME_WAIT | find /c /v ""
netsh int ipv4 show dynamicport tcp
```

**Resolution:**

```bash
# Linux
echo "net.ipv4.ip_local_port_range = 10000 65535" | sudo tee -a /etc/sysctl.conf
echo "net.ipv4.tcp_tw_reuse = 1" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p

# Windows (PowerShell as Administrator)
netsh int ipv4 set dynamicport tcp start=10000 num=55535
Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters" `
    -Name "TcpTimedWaitDelay" -Value 30 -Type DWord
```

#### Issue: High Connection Latency

**Symptoms:**
- Slow response times
- Timeouts
- Poor user experience

**Diagnosis:**

```bash
# Linux
ping -c 10 server-ip
traceroute server-ip
mtr server-ip

# Check TCP handshake time
tcpdump -i any port 8080 -w capture.pcap
# Analyze in Wireshark

# Windows
ping -n 10 server-ip
tracert server-ip
Test-NetConnection -ComputerName server-ip -Port 8080 -InformationLevel Detailed
```

**Resolution:**

```bash
# Enable TCP Fast Open
# Linux
echo "net.ipv4.tcp_fastopen = 3" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p

# Windows
Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters" `
    -Name "EnableTcpFastOpen" -Value 1 -Type DWord

# Optimize TCP parameters
# See "Platform-Specific Network Tuning" section above
```

#### Issue: Memory Leaks

**Symptoms:**
- Continuously increasing memory usage
- Out of memory errors
- Application crashes

**Diagnosis:**

```bash
# Linux - Monitor memory over time
while true; do
    ps aux | grep rockstar-app | awk '{print $6}'
    sleep 60
done

# Take heap profiles
go tool pprof http://localhost:6060/debug/pprof/heap

# Windows
Get-Process -Name "server" | Select-Object WorkingSet64
```

**Resolution:**

```go
// Check for common leak sources:
// 1. Unclosed resources
defer file.Close()
defer conn.Close()

// 2. Goroutine leaks
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// 3. Growing slices/maps
// Use bounded caches with eviction policies

// 4. Circular references
// Break references when done
```

## Cross-Platform Best Practices

### Write Once, Optimize Everywhere

```go
import (
    "runtime"
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func NewOptimizedConfig() pkg.FrameworkConfig {
    config := pkg.FrameworkConfig{
        Server: pkg.ServerConfig{
            EnableHTTP1: true,
            HTTP1Port:   8080,
        },
    }
    
    // Platform-specific optimizations
    info := pkg.GetPlatformInfo()
    
    config.Server.ListenerConfig = pkg.ListenerConfig{
        Network:        "tcp",
        Address:        ":8080",
        EnablePrefork:  info.SupportsPrefork,
        PreforkWorkers: info.NumCPU,
        ReusePort:      info.SupportsReusePort,
        ReuseAddr:      true,
    }
    
    // Adjust buffer sizes based on platform
    if runtime.GOOS == "linux" {
        config.Server.ListenerConfig.ReadBuffer = 131072
        config.Server.ListenerConfig.WriteBuffer = 131072
    } else {
        config.Server.ListenerConfig.ReadBuffer = 65536
        config.Server.ListenerConfig.WriteBuffer = 65536
    }
    
    return config
}
```

### Testing Across Platforms

```go
// +build integration

package main

import (
    "runtime"
    "testing"
)

func TestPlatformOptimizations(t *testing.T) {
    config := NewOptimizedConfig()
    
    // Verify platform-specific settings
    switch runtime.GOOS {
    case "linux":
        if !config.Server.ListenerConfig.ReusePort {
            t.Error("ReusePort should be enabled on Linux")
        }
        if config.Server.ListenerConfig.ReadBuffer != 131072 {
            t.Error("Linux should use 128KB buffers")
        }
        
    case "windows":
        if config.Server.ListenerConfig.ReadBuffer != 65536 {
            t.Error("Windows should use 64KB buffers")
        }
    }
}
```

### Documentation and Deployment

Always document platform-specific requirements:

```markdown
## Deployment Requirements

### Linux
- Kernel 4.9+ (for BBR congestion control)
- File descriptor limit: 65536+
- Recommended: Ubuntu 20.04 LTS or newer

### Windows
- Windows Server 2016 or newer
- .NET Framework 4.7.2+ (if using certain features)
- Administrator privileges for registry tuning

### Common
- Go 1.21 or newer
- 4GB RAM minimum, 8GB recommended
- 2 CPU cores minimum, 4+ recommended
```

## See Also

- [Performance Tuning Guide](performance-tuning.md) - General performance optimization
- [Configuration Reference](../CONFIGURATION_REFERENCE.md) - Complete configuration options
- [Monitoring Guide](monitoring.md) - Application monitoring and observability
- [Deployment Guide](deployment.md) - Production deployment best practices

