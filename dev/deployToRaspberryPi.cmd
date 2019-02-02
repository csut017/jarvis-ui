@ECHO off
SETLOCAL ENABLEEXTENSIONS
SET me=%~n0

SET target=192.168.0.7

ECHO %me%: Deploying to Raspberry Pi
pscp -pw letmein1 ../server/server pi@%target%:/home/pi/arduino/quicktest
pscp -pw letmein1 ../server/monitor.service pi@%target%:/home/pi/arduino/
pscp -pw letmein1 ../server/config-pi.json pi@%target%:/home/pi/arduino/config.json
REM pscp -pw letmein1 web/css/*.* pi@%target%:/home/pi/arduino/web/css/
REM pscp -pw letmein1 web/html/*.* pi@%target%:/home/pi/arduino/web/html/
REM pscp -pw letmein1 web/js/*.* pi@%target%:/home/pi/arduino/web/js/
REM pscp -pw letmein1 web/media/*.* pi@%target%:/home/pi/arduino/web/media/

ECHO %me%: Deploy completed
