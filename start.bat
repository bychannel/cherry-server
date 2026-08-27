@echo off
echo Starting all nodes...
echo.

cd bin

echo Starting master...
start "master" ./master.exe master --path=../config/cluster.json --node=gc-master

echo Starting center...
start "center" ./center.exe center --path=../config/cluster.json --node=gc-center

echo Starting web...
start "web" ./web.exe web --path=../config/cluster.json --node=gc-web-1

echo Starting gate...
start "gate" ./gate.exe gate --path=../config/cluster.json --node=gc-gate-1

echo Starting game...
start "game" ./game.exe game --path=../config/cluster.json --node=10001

cd ..
echo.
echo All nodes started!
pause
