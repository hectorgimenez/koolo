@echo off

echo Start building Koolo
echo Cleaning up previous artifacts...
if exist build rmdir /s /q build > NUL || goto :error

echo Building Koolo binary...
go build -tags static --ldflags -extldflags="-static" -o build/koolo.exe ./cmd/koolo/main.go > NUL || goto :error

echo Copying assets...
mkdir build\config > NUL || goto :error
copy config\config.yaml.dist build\config\config.yaml  > NUL || goto :error
xcopy /q /E /I /y config\pickit build\config\pickit  > NUL || goto :error
xcopy /q /y rustdecrypt.dll build  > NUL || goto :error
xcopy /q /y koolo-map.exe build > NUL || goto :error
xcopy /q /y d2.install.reg build > NUL || goto :error
xcopy /q /y README.md build > NUL || goto :error

echo Done! Artifacts are in build directory.

:error
if %errorlevel% neq 0 (
    echo Error occurred #%errorlevel%.
    exit /b %errorlevel%
)
