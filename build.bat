rmdir /s /q build
go build -tags static --ldflags -extldflags="-static" -o build/koolo.exe ./cmd/koolo/main.go
mkdir build\config
copy config\config.yaml.dist build\config\config.yaml
xcopy /E /I /y config\pickit build\config\pickit
xcopy /y rustdecrypt.dll build
xcopy /y koolo-map.exe build
xcopy /y d2.install.reg build
xcopy /y README.md build
