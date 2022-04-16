go build -tags static --ldflags '-extldflags="-static"' -o build/koolo.exe ./cmd/koolo/main.go
xcopy /E /I /y config build\config
xcopy /E /I /y MapAssist build\MapAssist