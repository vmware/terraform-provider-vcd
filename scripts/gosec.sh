#!/bin/bash
scripts_dir=$(dirname $0)
cd $scripts_dir
scripts_dir=$PWD
cd - > /dev/null

sc_exit_code=0

if [ ! -d ./govcd ]
then
    echo "source directory ./govcd not found"
    exit 1
fi

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
        # Variables found in gosec-config.sh
        # GOSEC_URL
        # GOSEC_VERSION
        if [ -f $scripts_dir/gosec-config.sh ]
        then
            source $scripts_dir/gosec-config.sh
        else
            echo "File $scripts_dir/gosec-config.sh not found - Skipping gosec"
            exit 0
        fi
        curl=$(exists_in_path curl)
        if [ -z "$curl" ]
        then
            echo "'curl' executable not found - Skipping gosec"
            exit 0
        fi
        $curl -sfL "$GOSEC_URL" | sh -s "$GOSEC_VERSION"
        exit_code=$?
        if [ "$exit_code" != "0" ]
        then
          echo "Error installing gosec"
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
        $gosec ./... -tags ALL .
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
