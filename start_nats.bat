@echo off
echo Starting NATS server with Docker...
docker run -d --name nats-server -p 4222:4222 -p 8222:8222 nats:latest
echo.
echo NATS is running!
echo - Client port: 4222
echo - Monitor port: 8222
echo.
echo To stop: docker stop nats-server
echo To view logs: docker logs nats-server
pause