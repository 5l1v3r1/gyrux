windows-amd64:
	@copy bin/windows-amd64/gyrux.exe C:\Windows\System32
windows-i386:
	@copy bin/windows-i386/gyrux.exe C:\Windows\System32
darwin-amd64:
	@cp bin/darwin-amd64/gyrux /usr/local/bin
	@chmod +x /usr/local/bin/gyrux
freebsd-amd64:
	@cp bin/freebsd-amd64/gyrux /bin
	@chmod +x /bin/gyrux
openbsd-amd64:
	@cp bin/openbsd-amd64/gyrux /bin
	@chmod +x /bin/gyrux
netbsd-amd64:
	@cp bin/netbsd-amd64/gyrux /bin
	@chmod +x /bin/gyrux
linux-amd64:
	@cp bin/linux-amd64/gyrux /bin
	@chmod +x /bin/gyrux
linux-arm64:
	@cp bin/linux-arm64/gyrux /bin
	@chmod +x /bin/gyrux
linux-i386:
	@cp bin/linux-i386/gyrux /bin
	@chmod +x /bin/gyrux
build:
	@chmod +x tools/build.sh
	@tools/build.sh
clean:
	@rm -r _bin
help:
	@echo "make linux-amd64"
	@echo "make linux-arm64"
	@echo "make linux-i386"
	@echo "make darwin-amd64"
	@echo "make windows-amd64"
	@echo "make windows-i386"
	@echo "make freebsd-amd64"
	@echo "make openbsd-amd64"
	@echo "make netbsd-amd64"
	@echo "make help"
	@echo "make build"
	@echo "make clean"
