#!/bin/sh

echo "Compiling AMD64 .deb"
fpm --debug --no-depends -m BladeMaker -s dir -t deb -n pgsql-dumper -v $RELEASE_TAG -p bin/pgsql-dumper_amd64.deb bin/pgsql-dumper-x86_64-linux=/usr/local/bin/pgsql-dumper

echo "Compiling i386 .deb"
fpm --debug --no-depends -m BladeMaker -s dir -t deb -n pgsql-dumper -v $RELEASE_TAG -a i386 -p bin/pgsql-dumper_i386.deb bin/pgsql-dumper-i386-linux=/usr/local/bin/pgsql-dumper
