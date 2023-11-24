@echo off

echo Installing dependencies...
go mod download

echo Building OpenCV... (this will take a while)
%GOPATH%\pkg\mod\gocv.io\x\gocv@v0.35.0\win_build_opencv.cmd static
