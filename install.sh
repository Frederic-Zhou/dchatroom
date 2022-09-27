GOOS=darwin go build -o ./dest/myd_darwin
GOOS=windows GOARCH=amd64 go build -o ./dest/myd_windows.exe
GOOS=linux GOARCH=amd64 go build -o ./dest/myd_linux
