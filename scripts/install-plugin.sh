#!/bin/bash

if [ ! -f VERSION ]
then
    echo "file 'VERSION' not found"
    exit 1
fi

version=$(cat VERSION)

[ -z "$GOPATH" ] && GOPATH=$HOME/go

if [ ! -d $GOPATH ]
then
    echo "GOPATH directory ($GOPATH) not found"
    exit 1
fi

plugin_name=terraform-provider-vcd

plugin_path=$GOPATH/bin/$plugin_name

if [ ! -f $plugin_path ]
then
    echo "Plugin binaries ($plugin_path) not found"
    echo "Run 'make build' to create them"
    exit 1
fi

target_dir=$HOME/.terraform.d/plugins/

if [ ! -d $target_dir ]
then
    mkdir -p $target_dir
fi

if [ ! -d $target_dir ]
then
    echo "I could not create $target_dir"
    exit 1
fi

target_file=$target_dir/${plugin_name}_$version

mv -v $plugin_path $target_file
exit_code=$?
if [ "$exit_code" == "0" ]
then
    echo "Installed"
    ls -lotr $target_dir
fi

exit $exit_code
