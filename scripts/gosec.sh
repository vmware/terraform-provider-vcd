#!/usr/bin/env bash
scripts_dir=$(dirname $0)
cd $scripts_dir
scripts_dir=$PWD
cd - > /dev/null

sc_exit_code=0

if [ ! -d ./vcd ]
then
    echo "source directory ./vcd not found"
    exit 1
fi

if [ ! -f ./scripts/gosec-config.sh ]
then
    echo "file ./scripts/gosec-config.sh not found"
    exit 1
fi

source ./scripts/gosec-config.sh

function exists_in_path {
    what=$1
    for dir in $(echo $PATH | tr ':' ' ')
    do
        wanted=$dir/$what
        if [ -x $wanted ]
        then
            echo $wanted
            return
        fi
    done
}

function get_gosec {
    gosec=$(exists_in_path gosec)
    if [ -z "$gosec" -a -n "$GITHUB_ACTIONS" ]
    then
        curl=$(exists_in_path curl)
        if [ -z "$curl" ]
        then
            echo "'curl' executable not found - Skipping gosec"
            exit 0
        fi
        $curl -sfL $GOSEC_URL > gosec_install.sh
        exit_code=$?
        if [ "$exit_code" != "0" ]
        then
          echo "Error downloading gosec installer"
          exit $exit_code
        fi
        sh -x gosec_install.sh $GOSEC_VERSION > gosec_install.log 2>&1
        if [ "$exit_code" != "0" ]
        then
          echo "Error installing gosec"
          cat gosec_install.log
          exit $exit_code
        fi
        gosec=$PWD/bin/gosec
    fi
    if [ -n "$gosec" ]
    then
        echo "## Found $gosec"
        echo -n "## "
        $gosec -version
    else
        echo "*** gosec executable not found - Exiting"
        exit 0
    fi
}

function run_gosec {
    if [ -n "$gosec" ]
    then
        go version
        $gosec -tests -tags ALL ./...
        exit_code=$?
        if [ "$exit_code" != "0" ]
        then
            sc_exit_code=$exit_code
        fi
    fi
    echo ""
}

get_gosec
echo ""

run_gosec
echo "Exit code: $sc_exit_code"
exit $sc_exit_code
