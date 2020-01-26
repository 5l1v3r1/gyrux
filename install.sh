#!/bin/sh

read -p $'What is your OS? (linux/darwin/freebsd/openbsd/netbsd): ' OS
if [[ $OS == "darwin" || $OS == "freebsd" || $OS == "openbsd" || $OS == "netbsd" ]]
then
    if [[ $OS == "darwin" ]]
    then
        {
        cp bin/darwin-amd64/gyrux /usr/local/bin
        chmod +x /usr/local/bin/gyrux
        } &> /dev/null
        exit
    else
        {
        cp bin/$OS-amd64/gyrux /bin
        chmod +x /bin/gyrux
        } &> /dev/null
        exit
    fi
else
    read -p $'What is your architecture? (amd64/arm64/i386): ' $ARCH
    {
    cp bin/linux-$ARCH /bin
    chmod +x /bin/gyrux
    } &> /dev/null
    exit
fi
