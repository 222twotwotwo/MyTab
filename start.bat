@echo off
setlocal

set "ROOT=%~dp0"
set "WEB_DIR=%ROOT%web"
set "SERVER_DIR=%ROOT%server"
set "ADDR=:8080"
set "DATA_DIR=%SERVER_DIR%\data"
set "GOCACHE=%ROOT%.cache\go-build"
set "GOMODCACHE=%ROOT%.cache\gomod"

echo.
echo [MyTab] Starting sandbox website...
echo.

powershell -NoProfile -ExecutionPolicy Bypass -Command "try { $r = Invoke-WebRequest -UseBasicParsing -TimeoutSec 2 http://localhost:8080/healthz; if ($r.StatusCode -eq 200 -and $r.Content -match 'ok') { exit 0 } } catch { } exit 1" >nul 2>nul
if not errorlevel 1 (
  echo [MyTab] Server is already running.
  start "" "http://localhost:8080"
  exit /b 0
)

netstat -ano | findstr /R /C:":8080 .*LISTENING" >nul 2>nul
if not errorlevel 1 (
  echo [Error] Port 8080 is already in use by another process.
  echo [Error] Stop that process or change ADDR in start.bat.
  pause
  exit /b 1
)

where go >nul 2>nul
if errorlevel 1 (
  echo [Error] Go is not installed or not in PATH.
  pause
  exit /b 1
)

where npm >nul 2>nul
if errorlevel 1 (
  echo [Error] npm is not installed or not in PATH.
  pause
  exit /b 1
)

if not exist "%WEB_DIR%\node_modules" (
  echo [MyTab] Installing frontend dependencies...
  pushd "%WEB_DIR%"
  call npm install
  if errorlevel 1 (
    popd
    echo [Error] npm install failed.
    pause
    exit /b 1
  )
  popd
)

echo [MyTab] Building frontend...
pushd "%WEB_DIR%"
call npm run build
if errorlevel 1 (
  popd
  echo [Error] Frontend build failed.
  pause
  exit /b 1
)
popd

if not exist "%DATA_DIR%" mkdir "%DATA_DIR%"
if not exist "%GOCACHE%" mkdir "%GOCACHE%"
if not exist "%GOMODCACHE%" mkdir "%GOMODCACHE%"

echo.
echo [MyTab] Website: http://localhost:8080
echo [MyTab] Press Ctrl+C to stop the server.
echo.

start "" "http://localhost:8080"

pushd "%SERVER_DIR%"
go run -buildvcs=false ./cmd/server
set "EXIT_CODE=%ERRORLEVEL%"
popd

echo.
echo [MyTab] Server stopped.
pause
exit /b %EXIT_CODE%
