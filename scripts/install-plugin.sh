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

function check_empty {
    variable="$1"
    message="$2"
    if [ -z "$variable" ]
    then
        echo $message
        exit 1
    fi
}

[ -z "$terraform" ] && terraform=terraform
# Getting the version without the initial "v"
bare_version=$(echo $version | sed -e 's/^v//')

# The default target directory is the pre-0.13 one
target_dir=$HOME/.terraform.d/plugins/

terraform_exec=""
for dir in $(echo $PATH | tr ':' ' ')
do
    if [ -x $dir/$terraform ]
    then
        terraform_exec=$dir/$terraform
        break
    fi
done

if [ -z "$terraform_exec" ]
then
    echo "$terraform executable not found"
    exit 1
fi

# Terraform version is used to determine what the target directory should be
terraform_version=$($terraform version | head -n 1| sed -e 's/Terraform v//')
check_empty "$terraform_version" "terraform_version not detected"


terraform_major=$(echo $terraform_version | tr '.' ' '| awk '{print $1}')
check_empty "$terraform_major" "terraform_version major not detected"
terraform_minor=$(echo $terraform_version | tr '.' ' '| awk '{print $2}')
check_empty "$terraform_minor" "terraform_version minor not detected"
os=$(uname -s | tr '[A-Z]' '[a-z]')
check_empty "$os" "operating system not detected"
goos=$(go env GOOS)
goarch=$(go env GOARCH)
arch=${goos}_${goarch}


# if terraform executable is 0.13+, we use the new path
if [[ $terraform_major -gt 0 || $terraform_major -eq 0 && $terraform_minor > 12 ]]
then
    target_dir=$HOME/.terraform.d/plugins/registry.terraform.io/vmware/vcd/$bare_version/$arch
fi

plugin_name=terraform-provider-vcd

plugin_path=$GOPATH/bin/$plugin_name

if [ ! -f $plugin_path ]
then
    echo "Plugin binaries ($plugin_path) not found"
    echo "Run 'make build' to create them"
    exit 1
fi


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
