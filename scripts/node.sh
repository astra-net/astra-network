#!/usr/bin/env bash

set -eu

# This script is used to download latest astra node release
# And run the binary. Many codes are copied from prysm.sh (eth2 client).
# Downloaded binaries is saved in staging/ directory
# Use ASTRA_RELEASE to specify a specific release version
# Example: ASTRA_RELEASE=v3.0.0 ./node.sh astra

ASTRA_SIGNING_KEY=539910B02B5BAC637F6615C9799ACE3A9B83DF60
DOWNLOAD_URL=https://github.com/astra-net/astra-network/releases/download
ASTRA_PUB_KEY=https://astranetwork.com/releases/pgp_keys.asc
version="v3 20201206.0"

unset -f progname color usage print_usage get_version do_verify do_download
progname="${0##*/}"

color() {
    # Usage: color "31;5" "string"
    # Some valid values for color:
    # - 5 blink, 1 strong, 4 underlined
    # - fg: 31 red,  32 green, 33 yellow, 34 blue, 35 purple, 36 cyan, 37 white
    # - bg: 40 black, 41 red, 44 blue, 45 purple
    printf '\033[%sm%s\033[0m\n' "$@"
}

# return the astra release version
get_version() {
    if [[ -n ${ASTRA_RELEASE:-} ]]; then
        readonly reason="specified in \$ASTRA_RELEASE"
        readonly astra_rel="${ASTRA_RELEASE}"
    else
        # Find the latest Astra release available for download.
        readonly reason="automatically selected latest available release"
        astra_rel=$(curl -f -s https://astranetwork.com/latest) || (color "31" "Get latest version failed. You may manually download the binaries from Github release page. https://github.com/astra-net/astra-network/releases/" && exit 1)
        readonly astra_rel
    fi
}

######## main #########
print_usage() {
   cat <<- ENDEND

usage: ${progname} [OPTIONS] PROCESS [ARGS]

PROCESS can be: validator/astra, btcrelay, ethrelay

OPTIONS:
   -h             print this help and exit
   -d             download only (default: off)
   -v             print out the version of the node.sh
   -V             print out the version of the Astra binary

ARGS will be passed to the PROCESS.

Ex:
   ${progname} -d validator
   ${progname} validator --help
   ${progname} validator --run explorer --run.shard=0

ENDEND
}

usage() {
   color "31" "$@"
   print_usage >&2
   exit 64  # EX_USAGE
}

failed_verification() {
    MSG=$(
        cat <<-END
Failed to verify Astra binary. Please erase downloads in the
staging directory and run this script again. Alternatively, you can use a
A prior version by specifying environment variable ASTRA_RELEASE 
with the specific version, as desired. Example: ASTRA_RELEASE=v2.4.0
If you must wish to continue running an unverified binary, specific the
environment variable ASTRA_UNVERIFIED=1
END
    )
    color "31" "$MSG"
    exit 1
}

do_verify() {
   local file=$1
   local binary="${file}-${arch}"

   skip=${ASTRA_UNVERIFIED-0}
   if [[ $skip == 1 ]]; then
       return 0
   fi
   checkSum="shasum -a 256"
   hash shasum 2>/dev/null || {
	  checkSum="sha256sum"
   	hash sha256sum 2>/dev/null || {
	     echo >&2 "SHA checksum utility not available. Either install one (shasum or sha256sum) or run with ASTRA_UNVERIFIED=1."
		  exit 1
    	}
   }
   hash gpg 2>/dev/null || {
      echo >&2 "gpg is not available. Either install it or run with ASTRA_UNVERIFIED=1."
      exit 1
   }

   color "32" "Verifying binary integrity."

   gpg --list-keys $ASTRA_SIGNING_KEY >/dev/null 2>&1 || curl --silent -L $ASTRA_PUB_KEY | gpg --import
   (
      cd "$wrapper_dir"
	   $checkSum -c "${binary}.sha256" || failed_verification
   )
   (
      cd "$wrapper_dir"
      gpg -u $ASTRA_SIGNING_KEY --verify "${binary}.sig" "$binary" || failed_verification
   )

   color "32;1" "Verified ${binary} has been signed by Astra."
}

do_download() {
   local file=$1
   local binary="${file}-${arch}"

   if [[ ! -x "${wrapper_dir}/${binary}" ]]; then
      color "32" "Downloading ${binary} (${reason})"

      curl -L "${DOWNLOAD_URL}/${astra_rel}/${binary}" -o "${wrapper_dir}/${binary}"
      curl --silent -L "${DOWNLOAD_URL}/${astra_rel}/${binary}.sha256" -o "${wrapper_dir}/${binary}.sha256"
      curl --silent -L "${DOWNLOAD_URL}/${astra_rel}/${binary}.sig" -o "${wrapper_dir}/${binary}.sig"
      chmod +x "${wrapper_dir}/${binary}"
   else
      color "37" "${binary} is up to date."
   fi
   cp -f "${wrapper_dir}/${binary}" "$file"
}

unset OPTIND OPTARG opt download_only
OPTIND=1
download_only=false

while getopts ":hdvV" opt
do
   case "${opt}" in
   '?') usage "unrecognized option -${OPTARG}";;
   ':') usage "missing argument for -${OPTARG}";;
   h) print_usage; exit 0;;
   d) download_only=true;;
   v) color "32" "$progname version: $version"
      exit 0 ;;
   V) INSTALLED_VERSION=$(./astra version 2>&1)
      RUNNING_VERSION=$(curl -s --request POST 'http://127.0.0.1:9500/' --header 'Content-Type: application/json' --data-raw '{ "jsonrpc": "2.0", "method": "astrav2_getNodeMetadata", "params": [], "id": 1}' | grep -Eo '"version":"[^"]*"' | cut -c11- | tr -d \")
      echo "Binary  Version: $INSTALLED_VERSION"
      echo "Running Version: $RUNNING_VERSION"
      exit 0 ;;
   *) color "31" "unhandled option -${OPTARG}";;  # EX_SOFTWARE
   esac
done

shift $((OPTIND - 1))
if [ "$#" -lt 1 ]; then
   usage "Missing PROCESS parameter"
fi

get_version
color "37" "Latest Astra release is: $astra_rel."

readonly wrapper_dir="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/staging/${astra_rel}"
VALIDATOR="astra"
BTCRELAY="astra-btcrelay"
ETHRELAY="astra-ethrelay"
mkdir -p "${wrapper_dir}"

arch=$(uname -m)
arch=${arch/x86_64/amd64}
arch=${arch/aarch64/arm64}

case "$1" in
   validator|astra)
      readonly process=${VALIDATOR}
      ;;
   btcrelay)
      readonly process=${BTCRELAY}
      ;;
   ethrelay)
      readonly process=${ETHRELAY}
      ;;
   *)
      usage "Process $1 is not found"
      ;;
esac

do_download "${process}"
do_verify "${process}"

if ${download_only}; then
   color "37" "Only download operation is requested, done."
   exit 0
fi

color "36" "Starting astra $1 ${*:2}"

exec -a "$0 ${process}" "./${process}" "${@:2}"

# vim: set expandtab:ts=3
