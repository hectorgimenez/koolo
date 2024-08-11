@echo off
setlocal enabledelayedexpansion

:: Change to the script's directory
cd /d "%~dp0"

call :print_header "Starting Koolo Build Process"

:: Check for Go installation
call :check_go_installation

:: Main script execution
call :main
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

:: Check if build folder exists
if not exist build (
    call :print_step "Creating build folder"
    mkdir build
    if !errorlevel! equ 0 (
        call :print_success "Build folder created at %cd%\build"
        
        :: Copy Settings.json to the new build folder
        call :print_step "Copying Settings.json to build folder"
        if exist config\Settings.json (
            copy /y config\Settings.json build\config\Settings.json > nul
            if !errorlevel! equ 0 (
                call :print_success "Settings.json successfully copied to %cd%\build\config"
            ) else (
                call :print_error "Failed to copy Settings.json from %cd%\config\ to %cd%\build\config"
                call :check_file_permissions "config\Settings.json"
            )
        ) else (
            call :print_error "Settings.json not found in %cd%\config folder"
        )
    ) else (
        call :print_error "Failed to create build folder"
        call :check_folder_permissions "%cd%"
        exit /b 1
    )
) else (
    call :print_step "Checking build folder"
    call :print_info "Build folder already exists at %cd%\build"
    
    :: Check and delete koolo.exe if it exists
    if exist build\koolo.exe (
        call :print_step "Removing existing koolo.exe"
        del /f /q build\koolo.exe
        if not exist build\koolo.exe (
            call :print_success "koolo.exe successfully deleted"
        ) else (
            call :print_error "Failed to delete koolo.exe"
            call :check_file_permissions "build\koolo.exe"
            exit /b 1
        )
    )
    
    :: Check and delete tools folder if it exists
    if exist build\tools (
        call :print_step "Removing existing tools folder"
        rmdir /s /q build\tools
        if not exist build\tools (
            call :print_success "Tools folder successfully deleted"
        ) else (
            call :print_error "Failed to delete tools folder"
            call :check_folder_permissions "build\tools"
            exit /b 1
        )
    )
)

:: Handle Settings.json
call :print_header "Handling Configuration Files"
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
            call :check_file_permissions "config\Settings.json"
            call :check_file_permissions "build\config\Settings.json"
        )
    ) else (
        call :print_info "Keeping existing Settings.json"
    )
) else (
    call :print_info "No existing Settings.json found in %cd%\build\config"
    call :print_step "Copying Settings.json"
    if not exist build\config mkdir build\config
    copy /y config\Settings.json build\config\Settings.json > nul
    if !errorlevel! equ 0 (
        call :print_success "Settings.json successfully copied to build\config"
    ) else (
        call :print_error "Failed to copy Settings.json to build\config"
        call :check_file_permissions "config\Settings.json"
        call :check_folder_permissions "build\config"
    )
)

:: Handle tools folder
call :print_header "Handling Tools"
if not exist build\tools (
    call :print_step "Copying tools folder"
    xcopy /q /E /I /y tools build\tools > nul
    if !errorlevel! equ 0 (
        call :print_success "Tools folder successfully copied"
    ) else (
        call :print_error "Failed to copy tools folder"
        call :check_folder_permissions "tools"
        call :check_folder_permissions "build"
    )
) else (
    call :print_info "Tools folder already exists in %cd%\build\tools"
)

:: Build Koolo binary
call :print_header "Building Koolo Binary"
call :print_step "Compiling Koolo"
if "%1"=="" (set VERSION=dev) else (set VERSION=%1)
go build -trimpath -tags static --ldflags -extldflags="-static" -ldflags="-s -w -H windowsgui -X 'github.com/hectorgimenez/koolo/internal/config.Version=%VERSION%'" -o build/koolo.exe ./cmd/koolo
if !errorlevel! equ 0 (
    call :print_success "Koolo binary successfully built at %cd%\build\koolo.exe"
) else (
    call :print_error "Failed to build Koolo binary"
    goto :error
)

:: Copy remaining assets
call :print_header "Copying Additional Assets"
if not exist build\config mkdir build\config

:: Handle koolo.yaml
if not exist build\config\koolo.yaml (
    call :print_step "Copying koolo.yaml.dist"
    copy config\koolo.yaml.dist build\config\koolo.yaml > nul
    if !errorlevel! equ 0 (
        call :print_success "koolo.yaml.dist successfully copied to build\config\koolo.yaml"
    ) else (
        call :print_error "Failed to copy koolo.yaml.dist"
        call :check_file_permissions "config\koolo.yaml.dist"
        call :check_folder_permissions "build\config"
    )
) else (
    call :print_info "koolo.yaml already exists in build\config, skipping copy"
)

call :print_step "Copying template folder"
xcopy /q /E /I /y config\template build\config\template > nul
if !errorlevel! equ 0 (
    call :print_success "Template folder successfully copied"
) else (
    call :print_error "Failed to copy template folder"
    call :check_folder_permissions "config\template"
    call :check_folder_permissions "build\config"
)

call :print_step "Copying README.md"
copy README.md build > nul
if !errorlevel! equ 0 (
    call :print_success "README.md successfully copied"
) else (
    call :print_error "Failed to copy README.md"
    call :check_file_permissions "README.md"
    call :check_folder_permissions "build"
)

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
powershell -Command "$acl = Get-Acl '%~1'; $identity = [System.Security.Principal.WindowsIdentity]::GetCurrent(); $principal = New-Object System.Security.Principal.WindowsPrincipal($identity); $adminRole = [System.Security.Principal.WindowsBuiltInRole]::Administrator; if ($principal.IsInRole($adminRole)) { Write-Host '    INFO: Script is running with administrator privileges' -ForegroundColor Yellow } else { Write-Host '    WARNING: Script is not running with administrator privileges' -ForegroundColor Yellow }; Write-Host ('    INFO: Current user: ' + $identity.Name) -ForegroundColor Yellow; Write-Host ('    INFO: File owner: ' + $acl.Owner) -ForegroundColor Yellow; $acl.Access | ForEach-Object { Write-Host ('    INFO: ' + $_.IdentityReference + ' has ' + $_.FileSystemRights + ' rights') -ForegroundColor Yellow }"
goto :eof

:: Function to check folder permissions
:check_folder_permissions
powershell -Command "$acl = Get-Acl '%~1'; $identity = [System.Security.Principal.WindowsIdentity]::GetCurrent(); $principal = New-Object System.Security.Principal.WindowsPrincipal($identity); $adminRole = [System.Security.Principal.WindowsBuiltInRole]::Administrator; if ($principal.IsInRole($adminRole)) { Write-Host '    INFO: Script is running with administrator privileges' -ForegroundColor Yellow } else { Write-Host '    WARNING: Script is not running with administrator privileges' -ForegroundColor Yellow }; Write-Host ('    INFO: Current user: ' + $identity.Name) -ForegroundColor Yellow; Write-Host ('    INFO: Folder owner: ' + $acl.Owner) -ForegroundColor Yellow; $acl.Access | ForEach-Object { Write-Host ('    INFO: ' + $_.IdentityReference + ' has ' + $_.FileSystemRights + ' rights') -ForegroundColor Yellow }"
goto :eof