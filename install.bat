@echo off
set /P ARCH="What is your architecture (amd64/i386): "
copy bin/windows-%ARCH%/gyrux.exe C:\Windows\System32
