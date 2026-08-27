@echo off
echo Stopping all nodes...
echo.

taskkill /FI "WINDOWTITLE eq master*" /F >nul 2>&1
echo master stopped.

taskkill /FI "WINDOWTITLE eq center*" /F >nul 2>&1
echo center stopped.

taskkill /FI "WINDOWTITLE eq web*" /F >nul 2>&1
echo web stopped.

taskkill /FI "WINDOWTITLE eq gate*" /F >nul 2>&1
echo gate stopped.

taskkill /FI "WINDOWTITLE eq game*" /F >nul 2>&1
echo game stopped.

echo.
echo All nodes stopped!
pause
