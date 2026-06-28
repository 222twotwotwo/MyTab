@echo off
setlocal EnableDelayedExpansion

set "ROOT=%~dp0"
set "SERVER_EXE=%ROOT%server\server.exe"
set "PORT=8080"
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

for /f "tokens=5" %%P in ('netstat -ano ^| findstr /R /C:":%PORT% .*LISTENING"') do (
  echo !KILLED_PIDS! | findstr /C:" %%P " >nul 2>nul
  if errorlevel 1 (
    set "FOUND=1"
    set "KILLED_PIDS=!KILLED_PIDS!%%P "
    echo [MyTab] Stopping process on port %PORT%: %%P
    taskkill /PID %%P /T /F >nul 2>nul
    if errorlevel 1 (
      echo [Warn] Failed to stop PID %%P.
    ) else (
      echo [MyTab] Stopped PID %%P.
    )
  )
)

powershell -NoProfile -ExecutionPolicy Bypass -Command "$target = [IO.Path]::GetFullPath('%SERVER_EXE%'); Get-CimInstance Win32_Process | Where-Object { $_.ExecutablePath -and ([IO.Path]::GetFullPath($_.ExecutablePath) -ieq $target) } | ForEach-Object { Write-Host ('[MyTab] Stopping residual server.exe PID ' + $_.ProcessId); Stop-Process -Id $_.ProcessId -Force }" 2>nul

netstat -ano | findstr /R /C:":%PORT% .*LISTENING" >nul 2>nul
if not errorlevel 1 (
  echo.
  echo [Warn] Port %PORT% is still in use.
  pause
  exit /b 1
)

if "%FOUND%"=="0" (
  echo [MyTab] No process was listening on port %PORT%.
)

echo [MyTab] Done.
pause
exit /b 0
