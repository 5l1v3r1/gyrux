#!/bin/sh -e

# Should be invoked from repo root.
# Requires Go >= 1.13 (for the -trimpath flag).

: ${VERSION:=v1.0}
: ${BIN_DIR:=./_bin}
: ${MANIFEST:=/dev/null}

export GOOS GOARCH
export CGO_ENABLED=0

# build $os $arch...
build() {
    local GOARCH=$1
    shift
    for GOOS in $@; do
        if [[ $GOARCH == "386" ]]
        then
            DST_DIR=$BIN_DIR/$GOOS-i386
        else
            DST_DIR=$BIN_DIR/$GOOS-$GOARCH
        fi
        mkdir -p $DST_DIR
        buildone
    done
}

# buildone
# Uses: $GOOS $GOARCH $BIN_DIR $MANIFEST
buildone() {
    if [[ $GOARCH == "386" ]]
    then    
        local DST_DIR=$BIN_DIR/$GOOS-i386
    else
        local DST_DIR=$BIN_DIR/$GOOS-$GOARCH
    fi
    local STEM=gyrux

    mkdir -p $DST_DIR

    if test $GOOS = windows; then
        BIN=$STEM.exe
        ARCHIVE=$STEM.zip
    else
        BIN=$STEM
        ARCHIVE=$STEM.tar.gz
    fi

    if [[ $GOARCH == "386" ]]
    then
        echo "Building for $GOOS-i386..."
    else
        echo "Building for $GOOS-$GOARCH..."
    fi
    go build -o $DST_DIR/$BIN -trimpath -ldflags \
        "-X github.com/entynetproject/gyrux/pkg/buildinfo.Version=$VERSION \
         -X github.com/entynetproject/gyrux/pkg/buildinfo.Reproducible=true" || {
        echo "Build: failed!"
        return
    }

    (
    cd $DST_DIR
    if test $GOOS = windows; then
        zip -q $ARCHIVE $BIN
    else
        tar cfz $ARCHIVE $BIN
    fi
    )

    echo "Build: done."
    echo $GOOS-$GOARCH/$BIN >> $MANIFEST
    echo $GOOS-$GOARCH/$ARCHIVE >> $MANIFEST

    if which sha256sum > /dev/null; then
        (
        cd $DST_DIR
        sha256sum $BIN > $BIN.sha256sum
        sha256sum $ARCHIVE > $ARCHIVE.sha256sum
        )
        echo $GOOS-$GOARCH/$BIN.sha256sum >> $MANIFEST
        echo $GOOS-$GOARCH/$ARCHIVE.sha256sum >> $MANIFEST
    fi
}

build amd64 linux darwin freebsd openbsd netbsd windows
build 386   linux windows
build arm64 linux
