@ECHO off
SETLOCAL ENABLEEXTENSIONS
SET me=%~n0

SET target=192.168.0.7

ECHO %me%: Deploying to Raspberry Pi
pscp -pw letmein1 ../server/server pi@%target%:/home/pi/arduino/quicktest
pscp -pw letmein1 ../server/monitor.service pi@%target%:/home/pi/arduino/
pscp -pw letmein1 ../server/config-pi.json pi@%target%:/home/pi/arduino/config.json
pscp -pw letmein1 ../server/web/*.* pi@%target%:/home/pi/arduino/web/

ECHO %me%: Deploy completed
