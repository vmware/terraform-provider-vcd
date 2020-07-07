#!/usr/bin/env bash

sources=.changes

if [ ! -d $sources ]
then
    echo "Directory $sources not found"
    exit 1
fi

cd $sources


for section in  features improvements bugfixes deprecations notes
do
    echo "## $(echo $section | tr 'a-z' 'A-Z')"
    for f in $(ls *${section}.md | sort -n)
    do
        #echo $f
        cat $f
    done
    echo ""
done


#484-improvements.md   487-improvements.md   499-improvements.md   504-features.md       511-bugfixes.md       512-bugfixes.md       514-features.md       518-features.md       522-bugfixes.md
# 484-improvements.md~  492-notes.md          501-improvements.md   505-improvements.md   511-features.md       513-features.md       518-deprecations.md   520-features.md       all.md
