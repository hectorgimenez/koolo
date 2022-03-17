go build -tags static --ldflags '-extldflags="-static"' -o build/koolo.exe ./cmd/koolo/main.go
xcopy /E /I config build\config