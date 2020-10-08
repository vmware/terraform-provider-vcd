#!/bin/bash

runtime_dir=$(dirname $0)
cd $runtime_dir
runtime_dir=$PWD
cd -

function check_exit_code {
    message=$1
    exit_code=$?
    if [ "$exit_code" != "0" ]
    then
        if [ -n "$message" ]
        then
            echo $message
        else
            echo "EXECUTION ERROR"
        fi
        exit $exit_code
    fi
 
}

# Make sure we have both versions available
for version_file in VERSION PREVIOUS_VERSION
do
    if [ ! -f $version_file ]
    then
        echo "$version_file not found"
        exit 1
    fi
done

# Gets the current branch
branch=$(git status | grep '^On branch' | awk '{print $3}')

if [ -z "$branch" ]
then
    echo "Current branch not detected"
    exit 1
fi


# ---------------------------------------------------------------------------
# Use "export skip_tags_fetching=1" if you want to test
# with your unchanged environment during development
#
# You will need to fetch upstream tags manually.
# Use with caution
# ---------------------------------------------------------------------------
if [ -n "$skip_tags_fetching" ]
then
    echo "# Tags fetching skipped: variable 'skip_tags_fetching' was set"
else
    # Make sure we have an upstream connection to the current branch
    upstream_exist=$(git remote | grep upstream)

    if [ -z "$upstream_exist" ]
    then
        git remote add upstream https://github.com/vmware/terraform-provider-vcd
        check_exit_code "Error adding upstream"
    fi

    # Retrieves tags from upstream. This will allow us to check out the
    # previously release tag
    git fetch --tags upstream master
    check_exit_code "Error fetching tags"

fi

# Gets the current and previous versions from files

current_version=$(cat VERSION | tr -d ' \t\n')
previous_version=$(cat PREVIOUS_VERSION | tr -d ' \t\n')

previous_version_exists=$(git tag | grep  "\<$previous_version\>")
if [ -z "$previous_version_exists" ]
then
    echo "Previous version $previous_version not among tags fetched from upstream"
    exit 1
fi

# ---------------------------------------------------------------------------
# use "export skip_binary_creation=1" to avoid creating binary test scripts.
#
# You will need to create the binary test files and the plugins on your own.
# Use with caution
# ---------------------------------------------------------------------------
if [ -n "$skip_binary_creation" ]
then
    echo "# Binary test creation skipped: variable 'skip_binary_creation' was set"
else

    # Checks that there are no uncommitted files

    modified=$(git status | grep -c modified )
    added=$(git status | grep -c 'new file' )
    if [ "$modified" != "0" -o "$added" != "0" ]
    then
        echo "# There are new or modified files"
        echo "# Upgrade test requires a clean branch"
        exit 1
    fi

    # Checks out the previous version. The binary test scripts and the old plugin 
    # need to be created with the older one
    git checkout $previous_version
    check_exit_code "error checking out previous version ($previous_version)"

    # Remove leftover .tf scripts to avoid running newer ones with the old plugins
    if [ -d vcd/test-artifacts ]
    then
        rm -f vcd/test-artifacts/*.tf
    fi

    # Creates the plugin and the test scripts with the older release
    make test-binary-prepare
    check_exit_code "error preparing binary tests"

    # Back to the current version
    git checkout $branch
    check_exit_code "Error checking out branch '$branch'"

    # Creates the plugin with the current branch
    make install
    check_exit_code "Error installing provider from branch '$branch'"
fi

# We need the latest 'test-binary.sh' to run the upgrade tests
cp $runtime_dir/test-binary.sh vcd/test-artifacts/
cp $runtime_dir/skip-upgrade-tests.txt vcd/test-artifacts/
cd vcd/test-artifacts

# Finally, we run the upgrade test itself
# During this test, the 'init', 'plan', and 'apply' phases will run with the old plugin
# while the 'plancheck' and 'destroy' phases will run with the new one
echo ./test-binary.sh upgrade $previous_version $current_version
if [ -n "$skip_upgrade_execution" ]
then
    echo "# upgrade test execution skipped - Variable 'skip_upgrade_execution' was set"
else
    ./test-binary.sh upgrade $previous_version $current_version
fi

