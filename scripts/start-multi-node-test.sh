#!/usr/bin/env bash

# Quality set for bash scripts:
# [set -e] fails on error
set -e

# [set -u ] stops the script on unset variables
set -u

# [set -o pipefail] triggers a failure if one of the commands in the pipe fail
set -o pipefail

node=$1
shift
partitions=$#
relative_configs=$@

if [[  $partitions < 2 ]]
then
    echo "Syntax: ./scripts/start-multi-node-test.sh node-number config1 config2 [config3 [config4 ...]]"
    exit 1
fi

if [ ! -d scripts ]
then
    echo "This script must be launched at the top of the repository, as ./scripts/start-multi-node-test.sh"
    exit 1
fi

multi_dir="split$partitions"

if [ -d "$multi_dir" ]
then
    echo "directory '$multi_dir' already exists"
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

for needed in tmux git realpath jq
do
    found_need=$(exists_in_path $needed)
    if [ -z "$found_need" ]
    then
        echo "needed tool $needed not found"
        exit 1
    fi
done

count=0
for tb in $relative_configs
do
    if [ ! -f $tb ]
    then
        echo "$tb not found"
        exit 1
    fi
    configs[$count]=$(realpath $tb)
    count=$((count+1))
done

current_repo_url=$(git config --get remote.origin.url)
if [ -z "$current_repo_url" ]
then
    echo "error retrieving current repository URL"
    exit 1
fi

current_commit=$(git rev-parse HEAD)
if [ -z "$current_commit" ]
then
    echo "error retrieving current commit hash"
    exit 1
fi

mkdir "${multi_dir}" || exit 1
cd ${multi_dir} || exit 1

echo "configs ${configs[*]}"
echo "partitions ${partitions}"

for n in $(seq 1 $partitions)
do
    git clone ${current_repo_url}
    if [ "$?" != "0" ]
    then
        echo "error cloning current repository"
        exit 1
    fi
    if [ ! -d terraform-provider-vcd ]
    then
        echo "terraform-provider-vcd not cloned in node $n"
        exit 1
    fi
    mv terraform-provider-vcd "node$n"
    cd "node$n/vcd"
    pwd
    index=$(($n-1))
    cp ${configs[$index]} vcd_test_config.json

    # TODO: this change is only needed while developing the current PR in branch
    git checkout test-partitioning
    # TODO: start tmux session with test
    # go test -tags functional -v -vcd-partitions=$partitions -vcd-partition-node=$n -vcd-short 2>&1 > testacc-node$n.txt
    cd -

done

