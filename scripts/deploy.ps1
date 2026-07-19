param(
  [string]$Server = "0x53.cn",
  [string]$User = "root",
  [string]$RemoteDir = "/root",
  [string]$FrontendDir = "/var/www/html",
  [switch]$SkipBuild,
  [switch]$SkipMP
)

$ProjectRoot = Resolve-Path "$PSScriptRoot\.."
$BackendDir = "$ProjectRoot\backend"
$FrontendDirLocal = "$ProjectRoot\frontend"
$BinaryName = "wtserver"

$ErrorActionPreference = "Stop"

function Log($msg) {
  Write-Host "[$(Get-Date -Format HH:mm:ss)] $msg" -ForegroundColor Cyan
}

function Step($msg) {
  Write-Host "`n=== $msg ===" -ForegroundColor Yellow
}

if (-not $SkipBuild) {
  Step "1/5 Build backend (linux/amd64)"
  Push-Location $BackendDir
  $env:GOOS = "linux"
  $env:GOARCH = "amd64"
  go build -o $BinaryName "./cmd/server/" 2>&1
  if (-not $?) { throw "Backend build failed" }
  Pop-Location
  Log "Backend built: $BackendDir\$BinaryName"

  Step "2/5 Build frontend H5"
  Push-Location $FrontendDirLocal
  npm run build:h5 2>&1
  if (-not $?) { throw "H5 build failed" }
  Pop-Location
  Log "H5 built"

  if (-not $SkipMP) {
    Step "3/5 Build frontend MP-Weixin"
    Push-Location $FrontendDirLocal
    npm run build:mp-weixin 2>&1
    if (-not $?) { throw "MP-Weixin build failed" }
    Pop-Location
    Log "MP-Weixin built"
  }
}

Step "4/5 Deploy backend"
Log "Upload binary..."
scp "$BackendDir\$BinaryName" "${User}@${Server}:${RemoteDir}/${BinaryName}_new" 2>&1
if (-not $?) { throw "SCP failed" }

Log "Replace binary & restart..."
$remoteCmd = "pm2 stop wtserver 2>/dev/null; fuser -k 8080/tcp 2>/dev/null; sleep 2; mv ${RemoteDir}/${BinaryName} ${RemoteDir}/${BinaryName}.bak; mv ${RemoteDir}/${BinaryName}_new ${RemoteDir}/${BinaryName}; chmod +x ${RemoteDir}/${BinaryName}; pm2 start ${RemoteDir}/${BinaryName} --name wtserver; sleep 3; pm2 status"
ssh ${User}@${Server} "$remoteCmd" 2>&1
Log "Backend deployed"

Step "5/5 Deploy frontend H5"
Log "Upload H5 files..."
ssh ${User}@${Server} "rm -rf ${FrontendDir}/assets ${FrontendDir}/static ${FrontendDir}/index.html ${FrontendDir}/favicon.ico" 2>&1 | Out-Null
scp -r "${FrontendDirLocal}\dist\build\h5\*" "${User}@${Server}:${FrontendDir}/" 2>&1
Log "H5 deployed"

Write-Host "`n===== DEPLOY COMPLETE =====" -ForegroundColor Green
Write-Host "H5:    https://wt.${Server}/" -ForegroundColor Green
Write-Host "API:   https://wt.${Server}/api/v1/" -ForegroundColor Green
if (-not $SkipMP) {
  Write-Host "MP:    dist\build\mp-weixin\ (import in WeChat DevTools)" -ForegroundColor Green
}
