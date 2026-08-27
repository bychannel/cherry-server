@echo off
setlocal

set BIN_DIR=bin
if not exist %BIN_DIR% mkdir %BIN_DIR%

cd nodes

echo Building master...
go build -o ../%BIN_DIR%/master.exe main.go

echo Building center...
go build -o ../%BIN_DIR%/center.exe main.go

echo Building web...
go build -o ../%BIN_DIR%/web.exe main.go

echo Building gate...
go build -o ../%BIN_DIR%/gate.exe main.go

echo Building game...
go build -o ../%BIN_DIR%/game.exe main.go

cd ..

echo.
echo Build complete!
pause
