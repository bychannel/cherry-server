@echo off

echo build go protocol file...
cd /d "%~dp0internal\protocol"
for %%f in (*.proto) do (
    protoc --go_out=../pb/ --go_opt=paths=source_relative %%f
)
echo build go proto complete!
