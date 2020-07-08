#!/usr/bin/env bash

# This script collects the single change files and generates CHANGELOG entries
# for the whole release

# .changes is the directory where the change files are
sources=.changes

if [ ! -d $sources ]
then
    echo "Directory $sources not found"
    exit 1
fi

# We can indicate a version on the command line
version=$1

# If no version was provided, we use the current release version
[  -z "$version" ] && version=$(cat VERSION)

# If the provided version does not exist, there is nothing to do
if [ ! -d  $sources/$version ]
then
    echo "# Changes directory $sources/$version not found"
    exit 1
fi

# The "sections" file contains the CHANGELOG headers
if [ ! -f $sources/sections ]
then
    echo "File $sources/sections not found"
    exit 1
fi
sections=$(cat $sources/sections)

cd $sources/$version

for section in $sections
do
    # Check whether we have any file for this section
    num=$(ls | grep "\-${section}.md" | wc -l | tr -d ' \t')
    # if there are no files for this section, we skip
    if [ "$num" == "0" ]
    then
        continue
    fi

    # Generate the header
    echo "## $(echo $section | tr 'a-z' 'A-Z' | tr '-' ' ')"

    # Print the changes files, sorted by PR number
    for f in $(ls *${section}.md | sort -n)
    do
        cat $f
    done
    echo ""
done

