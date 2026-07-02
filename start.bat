@echo off
setlocal EnableDelayedExpansion

set "ROOT=%~dp0"
set "WEB_DIR=%ROOT%web"
set "SERVER_DIR=%ROOT%server"
set "PORT_HELPER=%ROOT%scripts\mytab-port.ps1"
set "PORT_RANGE_START=8080"
set "PORT_RANGE_END=8099"
set "DEFAULT_PORT=8080"
set "CACHE_DIR=%ROOT%.cache"
set "PORT_FILE=%CACHE_DIR%\mytab-port.txt"
set "DATA_DIR=%SERVER_DIR%\data"
set "GOCACHE=%CACHE_DIR%\go-build"
set "GOMODCACHE=%CACHE_DIR%\gomod"

echo.
echo [MyTab] Starting sandbox website...
echo.

set "RUNNING_PORT="
for /f "delims=" %%P in ('powershell -NoProfile -ExecutionPolicy Bypass -File "%PORT_HELPER%" -Mode running -StartPort %PORT_RANGE_START% -EndPort %PORT_RANGE_END% -PortFile "%PORT_FILE%"') do set "RUNNING_PORT=%%P"
if defined RUNNING_PORT (
  echo [MyTab] Server is already running.
  start "" "http://localhost:!RUNNING_PORT!"
  exit /b 0
)

set "PORT="
for /f "delims=" %%P in ('powershell -NoProfile -ExecutionPolicy Bypass -File "%PORT_HELPER%" -Mode free -StartPort %PORT_RANGE_START% -EndPort %PORT_RANGE_END%') do set "PORT=%%P"
if not defined PORT (
  echo [Error] No available local port found.
  pause
  exit /b 1
)

set "ADDR=127.0.0.1:%PORT%"
set "SITE_URL=http://127.0.0.1:%PORT%"
if not "%PORT%"=="%DEFAULT_PORT%" (
  echo [MyTab] Port %DEFAULT_PORT% is unavailable; using %PORT%.
  echo.
)

if not exist "%CACHE_DIR%" mkdir "%CACHE_DIR%"

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
echo %PORT%>"%PORT_FILE%"

echo.
echo [MyTab] Website: %SITE_URL%
echo [MyTab] Press Ctrl+C to stop the server.
echo.

start "" "%SITE_URL%"

pushd "%SERVER_DIR%"
go run -buildvcs=false ./cmd/server
set "EXIT_CODE=%ERRORLEVEL%"
popd
del /q "%PORT_FILE%" >nul 2>nul

echo.
echo [MyTab] Server stopped.
pause
exit /b %EXIT_CODE%
