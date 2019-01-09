#!/bin/bash

echo "==> Searching for OVA download URL..."
ova_download_url=$(grep -w ovaDownloadUrl vcd/vcd_test_config.json | tr -d '",' | awk '{print $2}')
ova_file_name=$(grep -w ovaTestFileName vcd/vcd_test_config.json | tr -d '",' | awk '{print $2}')

if [[ -z $ova_download_url ]]; then
    echo 'no url found'; \
    exit 1
fi

echo "downloading ova"
if [ ! -f test-resources/$ova_file_name ] ; then \
    wget $ova_download_url -O "$1/test-resources/$ova_file_name"; \
else
    echo "file already downloaded"
fi 

exit 0
