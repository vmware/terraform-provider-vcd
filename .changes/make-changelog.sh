#!/usr/bin/env bash

sources=.changes

if [ ! -d $sources ]
then
    echo "Directory $sources not found"
    exit 1
fi

cd $sources


for section in  features improvements bug-fixes deprecations notes
do
    echo "## $(echo $section | tr 'a-z' 'A-Z' | tr '-' ' ')"
    for f in $(ls *${section}.md | sort -n)
    do
        #echo $f
        cat $f
    done
    echo ""
done

