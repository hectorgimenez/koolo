
echo Installing dependencies...
go mod download

echo %GOPATH%
dir %GOPATH%\pkg\mod

echo Building OpenCV... (this will take a while)
%GOPATH%\pkg\mod\gocv.io\x\gocv@v0.33.0\win_build_opencv.cmd static
