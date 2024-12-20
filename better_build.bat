@echo off
setlocal enabledelayedexpansion

:: Change to the script's directory
cd /d "%~dp0"

call :print_header "Starting Koolo Build Process"

:: Check for Go installation
call :check_go_installation

:: Main script execution
call :main
if !errorlevel! neq 0 (
    call :print_error "Build process failed with error code !errorlevel!"
    exit /b !errorlevel!
)
echo.
powershell -Command "Write-Host 'Press any key to exit...' -ForegroundColor Yellow"
pause > nul
exit /b 0

:check_go_installation
call :print_info "Checking if Go is installed"
where go >nul 2>&1
if %errorlevel% neq 0 (
    call :print_error "Go is not installed or not in the system PATH."
    call :print_info "You can download Go from https://golang.org/dl/"
    call :get_user_input "Do you want to attempt automatic installation using Chocolatey? (Y/N) " install_go
    if /i "!install_go!"=="Y" (
        call :install_go_with_chocolatey
    ) else (
        call :print_info "Please install Go manually and run this script again."
        exit /b 1
    )
) else (
    for /f "tokens=3" %%v in ('go version') do set go_version=%%v
    call :print_success "Go version !go_version! is installed."
)
goto :eof

:install_go_with_chocolatey
call :print_step "Attempting to install Go using Chocolatey..."
where choco >nul 2>&1
if %errorlevel% neq 0 (
    call :print_error "Chocolatey is not installed. Please install Go manually."
    call :print_info "You can install Chocolatey from https://chocolatey.org/install"
    exit /b 1
)
powershell -Command "Start-Process powershell -Verb runAs -ArgumentList 'choco install golang -y' -Wait"
where go >nul 2>&1
if %errorlevel% neq 0 (
    call :print_error "Failed to install Go. Please install it manually."
    exit /b 1
) else (
    call :print_success "Go has been successfully installed."
)
goto :eof

:main
:: Initial validation checks
call :validate_environment
if !errorlevel! neq 0 exit /b !errorlevel!

:: Build Koolo binary
call :print_header "Building Koolo Binary"
call :print_step "Compiling Koolo"
if "%1"=="" (set VERSION=dev) else (set VERSION=%1)
go build -trimpath -tags static --ldflags -extldflags="-static" -ldflags="-s -w -H windowsgui -X 'github.com/hectorgimenez/koolo/internal/config.Version=%VERSION%'" -o build/koolo.exe ./cmd/koolo
if !errorlevel! neq 0 (
    call :print_error "Failed to build Koolo binary"
    exit /b 1
)
call :print_success "Successfully built koolo.exe"

:: Handle tools folder first
call :print_header "Handling Tools"
if exist build\tools (
    call :print_step "Removing existing tools folder"
    rmdir /s /q build\tools
    if exist build\tools (
        call :print_error "Failed to delete tools folder"
        call :check_folder_permissions "build\tools"
        exit /b 1
    )
)
call :print_step "Copying tools folder"
xcopy /q /E /I /y tools build\tools > nul
if !errorlevel! neq 0 (
    call :print_error "Failed to copy tools folder"
    call :check_folder_permissions "tools"
    call :check_folder_permissions "build"
    exit /b 1
)
call :print_success "Tools folder successfully copied"

:: Handle Settings.json
call :print_header "Handling Configuration Files"
if not exist build\config mkdir build\config
if exist build\config\Settings.json (
    call :print_step "Checking Settings.json"
    call :print_info "Settings.json found in %cd%\build\config"
    call :get_user_input "    Do you want to replace it? (Y/N) " replace_settings
    if /i "!replace_settings!"=="Y" (
        call :print_step "Replacing Settings.json"
        copy /y config\Settings.json build\config\Settings.json > nul
        if !errorlevel! equ 0 (
            call :print_success "Settings.json successfully replaced"
        ) else (
            call :print_error "Failed to copy Settings.json"
            exit /b 1
        )
    ) else (
        call :print_info "Keeping existing Settings.json"
    )
) else (
    call :print_info "No existing Settings.json found in %cd%\build\config"
    call :print_step "Copying Settings.json"
    copy /y config\Settings.json build\config\Settings.json > nul
    if !errorlevel! neq 0 (
        call :print_error "Failed to copy Settings.json"
        exit /b 1
    )
    call :print_success "Settings.json successfully copied"
)

:: Handle koolo.yaml
if not exist build\config\koolo.yaml (
    call :print_step "Copying koolo.yaml.dist"
    copy config\koolo.yaml.dist build\config\koolo.yaml > nul
    if !errorlevel! neq 0 (
        call :print_error "Failed to copy koolo.yaml.dist"
        exit /b 1
    )
    call :print_success "koolo.yaml.dist successfully copied"
) else (
    call :print_info "koolo.yaml already exists in build\config, skipping copy"
)

:: Copy template folder
call :print_step "Copying template folder"
if exist build\config\template rmdir /s /q build\config\template
xcopy /q /E /I /y config\template build\config\template > nul
if !errorlevel! neq 0 (
    call :print_error "Failed to copy template folder"
    exit /b 1
)
call :print_success "Template folder successfully copied"

:: Copy README
call :print_step "Copying README.md"
copy README.md build > nul
if !errorlevel! neq 0 (
    call :print_error "Failed to copy README.md"
    exit /b 1
)
call :print_success "README.md successfully copied"

call :print_header "Build Process Completed"
call :print_success "Artifacts are in the build directory"
goto :eof

:error
call :print_header "Build Process Failed"
call :print_error "Error occurred during the build process"
goto :eof

:: Function to get user input
:get_user_input
setlocal enabledelayedexpansion
call :print_prompt "%~1"
set /p "user_input="
endlocal & set "%~2=%user_input%"
goto :eof

:: Function to print a colored prompt
:print_prompt
powershell -Command "Write-Host '%~1' -ForegroundColor Yellow -NoNewline"
goto :eof

:: Function to print a header
:print_header
echo.
powershell -Command "Write-Host '=== %~1 ===' -ForegroundColor Magenta"
echo.
goto :eof

:: Function to print a step
:print_step
powershell -Command "Write-Host '  - %~1' -ForegroundColor Cyan"
goto :eof

:: Function to print a success message
:print_success
powershell -Command "Write-Host '    SUCCESS: %~1' -ForegroundColor Green"
goto :eof

:: Function to print an error message
:print_error
powershell -Command "Write-Host '    ERROR: %~1' -ForegroundColor Red"
goto :eof

:: Function to print an info message
:print_info
powershell -Command "Write-Host '    INFO: %~1' -ForegroundColor Yellow"
goto :eof

:: Function to check file permissions
:check_file_permissions
dir "%~1" >nul 2>&1
if !errorlevel! neq 0 (
    call :print_error "Cannot access file: %~1"
) else (
    call :print_info "File %~1 is accessible"
)
goto :eof

:: Function to check folder permissions
:check_folder_permissions
dir "%~1\*" >nul 2>&1
if !errorlevel! neq 0 (
    call :print_error "Cannot access directory: %~1"
) else (
    call :print_info "Directory %~1 is accessible"
)
goto :eof

:: Function to validate environment
:validate_environment
call :print_header "Validating Environment"

:: Check for required source files and folders
if not exist config (
    call :print_error "Config directory is missing"
    exit /b 1
)

if not exist config\koolo.yaml.dist (
    call :print_error "koolo.yaml.dist is missing from config directory"
    exit /b 1
)

if not exist config\Settings.json (
    call :print_error "Settings.json is missing from config directory"
    exit /b 1
)

if not exist tools (
    call :print_error "Tools directory is missing"
    exit /b 1
)

:: Check for required tools
if not exist tools\handle64.exe (
    call :print_error "handle64.exe is missing from tools directory"
    exit /b 1
)

if not exist tools\koolo-map.exe (
    call :print_error "koolo-map.exe is missing from tools directory"
    exit /b 1
)

:: Check for required build dependencies
call :print_step "Checking build dependencies"
go version >nul 2>&1
if !errorlevel! neq 0 (
    call :print_error "Go is not installed or not in PATH"
    exit /b 1
)

:: Verify write permissions in current directory
call :print_step "Checking write permissions"
echo. > test_write.tmp 2>nul
if !errorlevel! neq 0 (
    call :print_error "No write permissions in current directory"
    exit /b 1
)
del test_write.tmp >nul 2>&1

call :print_success "Environment validation completed"
goto :eof

:: Function to print a warning message
:print_warning
powershell -Command "Write-Host '    WARNING: %~1' -ForegroundColor Yellow"
goto :eof