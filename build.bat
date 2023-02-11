go build -tags static --ldflags -extldflags="-static" -o build/koolo.exe ./cmd/koolo/main.go
xcopy /E /I /y config build\config
xcopy /y rustdecrypt.dll build
xcopy /y koolo-map.exe build
xcopy /y d2.install.reg build