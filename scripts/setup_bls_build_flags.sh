# no shebang; to be sourced from other scripts

unset -v progdir
case "${0}" in
*/*) progdir="${0%/*}";;
*) progdir=.;;
esac

case "${ASTRA_PATH+set}" in
"")
   unset -v gopath
   gopath=$(go env GOPATH)
   # ASTRA_PATH is the common root directory of all astra repos
   ASTRA_PATH="${gopath%%:*}/src/github.com/astra-net"
   if [ ! -d $ASTRA_PATH ]; then
      # "env pwd" uses external pwd(1) implementation and not the Bash built-in,
      # which does not fully dereference symlinks.
      ASTRA_PATH=$(cd $progdir/../.. && env pwd)
   fi
   ;;
esac
: ${OPENSSL_DIR="/usr/local/opt/openssl"}
: ${MCL_DIR="${ASTRA_PATH}/mcl"}
: ${BLS_DIR="${ASTRA_PATH}/bls"}
export CGO_CFLAGS="-I${BLS_DIR}/include -I${MCL_DIR}/include"
export CGO_LDFLAGS="-L${BLS_DIR}/lib"
export LD_LIBRARY_PATH=${BLS_DIR}/lib:${MCL_DIR}/lib

OS=$(uname -s)
case $OS in
   Darwin)
      export CGO_CFLAGS="-I${BLS_DIR}/include -I${MCL_DIR}/include -I${OPENSSL_DIR}/include"
      export CGO_LDFLAGS="-L${BLS_DIR}/lib -L${OPENSSL_DIR}/lib"
      export LD_LIBRARY_PATH=${BLS_DIR}/lib:${MCL_DIR}/lib:${OPENSSL_DIR}/lib
      export DYLD_FALLBACK_LIBRARY_PATH=$LD_LIBRARY_PATH
      ;;
esac

if [ "$1" = "-v" ]; then
   echo "{ \"CGO_CFLAGS\" : \"$CGO_CFLAGS\",
            \"CGO_LDFLAGS\" : \"$CGO_LDFLAGS\",
            \"LD_LIBRARY_PATH\" : \"$LD_LIBRARY_PATH\",
            \"DYLD_FALLBACK_LIBRARY_PATH\" : \"$DYLD_FALLBACK_LIBRARY_PATH\"}" | jq "."
fi
