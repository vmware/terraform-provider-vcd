#!/usr/bin/env bash

# This test checks that all the build tags defined in config_test.go and provider_test.go
# can run individually

if [ ! -d vcd ]
then
    echo "./vcd directory missing"
    exit 1
fi
cd vcd

if [ ! -f config_test.go ]
then
    echo "file ./vcd/config_test.go not found"
    exit 1
fi

start=$(date +%s)
tags1=$(head -n 1 config_test.go | sed -e 's/^.*build //')
tags2=$(head -n 1 provider_test.go | sed -e 's/^.*build //')
tags=$(echo $tags1 $tags2 | tr " " "\n" | sort | uniq| tr "\n" " "; echo) 

echo "=== RUN TagsTest"
for tag in $tags
do
    
    go test -tags $tag -timeout 0 -v -vcd-help > /dev/null

    if [ "$?" == "0" ]
    then
        echo "  --- PASS: TagsTest/$tag"
    else
        echo "  --- FAIL: TagsTest/$tag"
        failed="$failed $tag"
    fi
done

end=$(date +%s)
elapsed=$((end-start))
if [ -n "$failed" ]
then
    echo "--- FAIL: TagsTest - Tests for tags [$failed] have failed (${elapsed}s)"
    exit 1
fi
echo "--- PASS: TagsTest (${elapsed}s)"
exit 0
