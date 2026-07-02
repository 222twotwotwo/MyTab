@echo off
setlocal EnableDelayedExpansion

set "ROOT=%~dp0"
set "PORT_HELPER=%ROOT%scripts\mytab-port.ps1"
set "SERVER_EXE=%ROOT%server\server.exe"
set "PORT_FILE=%ROOT%.cache\mytab-port.txt"
set "PORT_RANGE_START=8080"
set "PORT_RANGE_END=8099"
set "FOUND=0"
set "KILLED_PIDS= "

echo.
echo [MyTab] Stopping sandbox website...
echo.

net session >nul 2>nul
if errorlevel 1 (
  echo [MyTab] Requesting administrator permission...
  powershell -NoProfile -ExecutionPolicy Bypass -Command "Start-Process -FilePath '%~f0' -Verb RunAs"
  if errorlevel 1 (
    echo [Error] Failed to request administrator permission.
    pause
    exit /b 1
  )
  exit /b 0
)

set "PORTS="
for /f "delims=" %%A in ('powershell -NoProfile -ExecutionPolicy Bypass -File "%PORT_HELPER%" -Mode running -StartPort %PORT_RANGE_START% -EndPort %PORT_RANGE_END% -PortFile "%PORT_FILE%" -All') do set "PORTS=!PORTS! %%A"

for %%A in (!PORTS!) do (
  for /f "tokens=5" %%P in ('netstat -ano ^| findstr /R /C:":%%A .*LISTENING"') do (
    echo !KILLED_PIDS! | findstr /C:" %%P " >nul 2>nul
    if errorlevel 1 (
      set "FOUND=1"
      set "KILLED_PIDS=!KILLED_PIDS!%%P "
      echo [MyTab] Stopping process on port %%A: %%P
      taskkill /PID %%P /T /F >nul 2>nul
      if errorlevel 1 (
        echo [Warn] Failed to stop PID %%P.
      ) else (
        echo [MyTab] Stopped PID %%P.
      )
    )
  )
)

powershell -NoProfile -ExecutionPolicy Bypass -Command "$target = [IO.Path]::GetFullPath('%SERVER_EXE%'); Get-CimInstance Win32_Process | Where-Object { $_.ExecutablePath -and ([IO.Path]::GetFullPath($_.ExecutablePath) -ieq $target) } | ForEach-Object { Write-Host ('[MyTab] Stopping residual server.exe PID ' + $_.ProcessId); Stop-Process -Id $_.ProcessId -Force }" 2>nul

set "STILL_RUNNING="
for /f "delims=" %%A in ('powershell -NoProfile -ExecutionPolicy Bypass -File "%PORT_HELPER%" -Mode running -StartPort %PORT_RANGE_START% -EndPort %PORT_RANGE_END% -PortFile "%PORT_FILE%" -All') do set "STILL_RUNNING=!STILL_RUNNING! %%A"
if defined STILL_RUNNING (
  echo.
  echo [Warn] MyTab is still running on port(s):%STILL_RUNNING%
  pause
  exit /b 1
)

if "%FOUND%"=="0" (
  echo [MyTab] No MyTab server was listening on ports %PORT_RANGE_START%-%PORT_RANGE_END%.
)

del /q "%PORT_FILE%" >nul 2>nul

echo [MyTab] Done.
pause
exit /b 0
