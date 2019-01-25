@ECHO off
SETLOCAL ENABLEEXTENSIONS
SET me=%~n0

ECHO %me%: Building Windows executable
go build

IF %ERRORLEVEL% NEQ 0 (
    EXIT /B 0
)

ECHO %me%: Building Raspberry Pi executable
set GOARCH=arm
set GOOS=linux
go build

IF %ERRORLEVEL% NEQ 0 (
    EXIT /B 0
)

ECHO %me%: Build completed
