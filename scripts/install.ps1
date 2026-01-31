# ISP Visual Monitor Agent - Windows Installation Script
# Run with: powershell -ExecutionPolicy Bypass -File install.ps1

$ErrorActionPreference = "Stop"

# Configuration
$InstallDir = "C:\Program Files\ISPAgent"
$ConfigDir = "C:\ProgramData\ISPAgent"
$LogDir = "C:\ProgramData\ISPAgent\logs"
$ServiceName = "ISPAgent"
$RepoUrl = "https://github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent"

Write-Host "ISP Visual Monitor Agent - Installation" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

# Check if running as administrator
if (-NOT ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Host "Error: Please run as Administrator" -ForegroundColor Red
    exit 1
}

# Detect architecture
$Arch = $env:PROCESSOR_ARCHITECTURE
if ($Arch -eq "AMD64") {
    $BinaryName = "ispagent-windows-amd64.exe"
} else {
    Write-Host "Error: Unsupported architecture: $Arch" -ForegroundColor Red
    exit 1
}

Write-Host "Detected architecture: $Arch"
Write-Host "Binary: $BinaryName"
Write-Host ""

# Get latest release version
Write-Host "Fetching latest release..."
try {
    $LatestRelease = Invoke-RestMethod -Uri "https://api.github.com/repos/MohamadKhaledAbbas/ISPVisualMonitor-Agent/releases/latest"
    $LatestVersion = $LatestRelease.tag_name
} catch {
    Write-Host "Warning: Could not fetch latest version. Using 'latest'" -ForegroundColor Yellow
    $LatestVersion = "latest"
}

Write-Host "Version: $LatestVersion"
Write-Host ""

# Download binary
Write-Host "Downloading agent binary..."
$DownloadUrl = "$RepoUrl/releases/download/$LatestVersion/ispagent-windows-amd64.zip"
$TempDir = New-Item -ItemType Directory -Path "$env:TEMP\ispagent-install-$(Get-Random)" -Force
$ZipPath = Join-Path $TempDir "ispagent.zip"

try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipPath
} catch {
    Write-Host "Error: Failed to download binary" -ForegroundColor Red
    Remove-Item -Recurse -Force $TempDir
    exit 1
}

# Extract and install
Write-Host "Installing agent..."
Expand-Archive -Path $ZipPath -DestinationPath $TempDir -Force
New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
Copy-Item -Path (Join-Path $TempDir "ispagent-windows-amd64.exe") -Destination (Join-Path $InstallDir "ispagent.exe") -Force

# Create directories
Write-Host "Creating directories..."
New-Item -ItemType Directory -Path $ConfigDir -Force | Out-Null
New-Item -ItemType Directory -Path $LogDir -Force | Out-Null

# Download example config
Write-Host "Installing configuration..."
$ConfigFile = Join-Path $ConfigDir "agent.yaml"
if (-not (Test-Path $ConfigFile)) {
    $ExampleConfigUrl = "$RepoUrl/raw/main/configs/agent.yaml.example"
    try {
        Invoke-WebRequest -Uri $ExampleConfigUrl -OutFile $ConfigFile
        Write-Host "Created default configuration: $ConfigFile" -ForegroundColor Green
    } catch {
        Write-Host "Warning: Could not download example config" -ForegroundColor Yellow
    }
} else {
    Write-Host "Configuration already exists, skipping..." -ForegroundColor Yellow
}

# Install as Windows Service
Write-Host "Installing Windows service..."
$ServicePath = Join-Path $InstallDir "ispagent.exe"
$ServiceArgs = "--config `"$ConfigFile`""

# Remove existing service if present
if (Get-Service -Name $ServiceName -ErrorAction SilentlyContinue) {
    Write-Host "Removing existing service..."
    Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue
    sc.exe delete $ServiceName
    Start-Sleep -Seconds 2
}

# Create service
sc.exe create $ServiceName binPath= "`"$ServicePath`" $ServiceArgs" start= auto DisplayName= "ISP Visual Monitor Agent"
sc.exe description $ServiceName "Collects monitoring data from network routers"
sc.exe failure $ServiceName reset= 86400 actions= restart/60000/restart/60000/restart/60000

Write-Host "Service installed and set to automatic start" -ForegroundColor Green

# Cleanup
Remove-Item -Recurse -Force $TempDir

Write-Host ""
Write-Host "Installation complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:"
Write-Host "1. Edit configuration: $ConfigFile"
Write-Host "2. Add your license key and router details"
Write-Host "3. Start the service: Start-Service $ServiceName"
Write-Host "4. Check status: Get-Service $ServiceName"
Write-Host "5. View logs: Get-Content $LogDir\agent.log -Tail 50 -Wait"
Write-Host ""
Write-Host "Documentation: $RepoUrl/tree/main/docs"
Write-Host ""
