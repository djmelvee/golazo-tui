@echo off
chcp 65001 >nul
cd /d "%~dp0"

echo.
echo   ===========================================
echo    FIFA WORLD CUP 2026  -  GOLAZO TUI
echo    USA  ^|  Canada  ^|  Mexico  ^|  48 Teams
echo   ===========================================
echo.

if not exist "bin" mkdir bin

echo   Building golazo-seed...
go build -o bin\golazo-seed.exe .\cmd\golazo-seed
if %ERRORLEVEL% neq 0 (
    echo.
    echo   ERROR: Build failed. Is Go installed and in PATH?
    pause
    exit /b 1
)

echo   Building golazo-tui...
go build -o bin\golazo-tui.exe .\cmd\golazo-tui
if %ERRORLEVEL% neq 0 (
    echo.
    echo   ERROR: Build failed.
    pause
    exit /b 1
)

echo   Seeding match data...
bin\golazo-seed.exe
if %ERRORLEVEL% neq 0 (
    echo.
    echo   ERROR: Seed failed.
    pause
    exit /b 1
)

echo.
echo   Launching... (press q to quit)
echo.
bin\golazo-tui.exe
