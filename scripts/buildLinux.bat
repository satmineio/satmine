cd %~dp0
cd ..
set GOOS=linux
set GOARCH=amd64
go build -o ./build/satmine ./cmd





