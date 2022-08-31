#!/bin/sh

echo "Compiling AMD64 .deb"
fpm --debug --no-depends -m BladeMaker -s dir -t deb -n psql-dumper -v $RELEASE_TAG -p bin/psql-dumper_${RELEASE_TAG}_amd64.deb bin/psql-dumper-x86_64-linux=/usr/local/psql-dumper

echo "Compiling i386 .deb"
fpm --debug --no-depends -m BladeMaker -s dir -t deb -n psql-dumper -v $RELEASE_TAG -a i386 -p bin/psql-dumper_${RELEASE_TAG}_i386.deb bin/psql-dumper-i386-linux=/usr/local/psql-dumper