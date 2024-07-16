@echo off

echo Start building Koolo
echo Cleaning up previous artifacts...
if exist build rmdir /s /q build > NUL || goto :error

echo Building Koolo binary...
if "%1"=="" (set VERSION=dev) else (set VERSION=%1)
go build -trimpath -tags static --ldflags -extldflags="-static" -ldflags="-s -w -H windowsgui -X 'github.com/hectorgimenez/koolo/internal/config.Version=%VERSION%'" -o build/koolo.exe ./cmd/koolo > NUL || goto :error

echo Copying assets...
mkdir build\config > NUL || goto :error
copy config\koolo.yaml.dist build\config\koolo.yaml  > NUL || goto :error
copy config\Settings.json build\config\Settings.json  > NUL || goto :error
xcopy /q /E /I /y config\template build\config\template  > NUL || goto :error
xcopy /q /E /I /y tools build\tools > NUL || goto :error
xcopy /q /y README.md build > NUL || goto :error

echo Done! Artifacts are in build directory.

:error
if %errorlevel% neq 0 (
    echo Error occurred #%errorlevel%.
    exit /b %errorlevel%
)