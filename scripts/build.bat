@echo off
setlocal enabledelayedexpansion

echo ========================================
echo Building Rafta for Windows, Linux, Android
echo ========================================

set VERSION=1.0.0
set PROJECT_ROOT=%CD%
set DIST_DIR=%PROJECT_ROOT%\dist

REM Create dist directory
if not exist "%DIST_DIR%" mkdir "%DIST_DIR%"

REM ------------------------------------------------------------
REM 1. Build for Windows (amd64)
REM ------------------------------------------------------------
echo.
echo [1/2] Building for Windows (amd64)...
go build -ldflags="-s -w" -o "%DIST_DIR%\rafta-windows-amd64.exe" .\cmd\rafta
if errorlevel 1 (
    echo ERROR: Windows build failed
    exit /b 1
)
echo OK: %DIST_DIR%\rafta-windows-amd64.exe

REM ------------------------------------------------------------
REM 2. Build for Android (APK)
REM ------------------------------------------------------------
echo.
echo [2/2] Building for Android...
cd cmd\rafta

REM Check if icon exists
if not exist ..\..\icon.png (
    echo WARNING: icon.png not found, building without icon
    set ICON_FLAG=
) else (
    set ICON_FLAG=--icon ..\..\icon.png
)

fyne package --os android --app-id com.burhanarif.rafta %ICON_FLAG%
if errorlevel 1 (
    echo ERROR: Android build failed
    cd ..\..
    exit /b 1
)

REM Move the generated APK to dist
if exist rafta.apk (
    copy rafta.apk ..\..\dist\rafta-android.apk > nul
    del rafta.apk
    echo OK: ..\..\dist\rafta-android.apk
) else (
    echo ERROR: APK not found after build
    cd ..\..
    exit /b 1
)
cd ..\..

REM ------------------------------------------------------------
REM Done
REM ------------------------------------------------------------
echo.
echo ========================================
echo All builds completed successfully!
echo Binaries are in the 'dist' folder.
echo ========================================
