#!/bin/bash
# Based on install.sh from idada/v8.go - thanks.
# To compile a debug version use one of the following commands
#  version=4.3.59 target=x64.debug ./build.sh
#  version=4.3.59 target=x64.release ./build.sh
#  version=4.3.59 target=ia32.debug ./build.sh
#  version=4.3.59 arget=ia32.release ./build.sh

set -e -x

# go into directory containing this script
cd `dirname "${BASH_SOURCE[0]}"`

# build v8
make -C "v8-$version" i18nsupport=off $target

outdir="`pwd`/v8-$version/out/$target"

libv8_base="`find $outdir -name 'libv8_base.a' | head -1`"
libv8_libbase="`find $outdir -name 'libv8_libbase.a' | head -1`"
libv8_libplatform=`find $outdir -name 'libv8_libplatform.a' | head -1`""
libv8_snapshot=`find $outdir -name 'libv8_snapshot.a' | head -1`""

# for Linux
librt=''
start_group=''
end_group=''
if [ `go env | grep GOHOSTOS` == 'GOHOSTOS="linux"' ]; then
	librt='-lrt'
	start_group='-Wl,--start-group'
	end_group='-Wl,--end-group'
fi

# for Mac
libstdcpp=''
if  [ `go env | grep GOHOSTOS` == 'GOHOSTOS="darwin"' ]; then
	libstdcpp='-stdlib=libstdc++'
fi

# create package config file
echo "Name: v8
Description: v8 javascript engine
Version: $version $target
Cflags: $libstdcpp -I`pwd` -I`pwd`/v8-$version/include -I`pwd`/v8-$version/
Libs: $libstdcpp $start_group $libv8_libbase $libv8_base $libv8_libplatform \
$libv8_snapshot $end_group $librt" > v8.pc
