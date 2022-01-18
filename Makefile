.PHONY: all

build:
	go build -tags static --ldflags '-extldflags="-static"' -o build/koolo.exe ./cmd/koolo/main.go
	xcopy /E /I assets build\assets
	xcopy /E /I config build\config