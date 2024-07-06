@echo off
setlocal enabledelayedexpansion

:: Main script execution
call :main
echo.
powershell -Command "Write-Host 'Press any key to exit...' -ForegroundColor Yellow"
pause > nul
exit /b 0

:main
call :print_header "Starting Koolo Build Process"

:: Check if build folder exists
if not exist build (
    call :print_step "Creating build folder"
    mkdir build
    call :print_success "Build folder created at %cd%\build"
) else (
    call :print_step "Checking build folder"
    call :print_info "Build folder already exists at %cd%\build"
    
    :: Check and delete koolo.exe if it exists
    if exist build\koolo.exe (
        call :print_step "Removing existing koolo.exe"
        del /q build\koolo.exe
        if not exist build\koolo.exe (
            call :print_success "koolo.exe successfully deleted"
        ) else (
            call :print_error "Failed to delete koolo.exe"
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
        )
    ) else (
        call :print_info "Keeping existing Settings.json"
    )
) else (
    call :print_info "No existing Settings.json found in %cd%\build\config"
    call :print_step "Copying Settings.json"
    if not exist build\config (
    	call :print_info "No config folder found in %cd%\build"
	call :print_step "Creating config folder in %cd%\build"
	mkdir build\config
    )
    copy /y config\Settings.json build\config\Settings.json > nul
    if !errorlevel! equ 0 (
            call :print_success "Settings.json successfully copied"
        ) else (
            call :print_error "Failed to copy Settings.json"
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
)

call :print_step "Copying README.md"
copy README.md build > nul
if !errorlevel! equ 0 (
    call :print_success "README.md successfully copied"
) else (
    call :print_error "Failed to copy README.md"
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