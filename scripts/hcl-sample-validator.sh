#!/bin/bash


FOUND_ERROR=0

# extract_hcl searches for .markdown files by using glob 'website/docs/{*/,?}*markdown' in:
# * website/docs/*.markdown
# * website/docs/*/*.markdown
# It will look for code blocks starting with '```hcl' and extract their contents until closing '```'
# and store in a file inside a tmp directory. Filename will be "base_filename+total_occurence_number"
# (e.g. edgegateway.html.markdown-100.tf)
function extract_hcl {
    CURDIR=$PWD

    [ -d "tmp" ] && rm -r tmp
    mkdir tmp
    cd tmp

    awk 'function basename(file) {
        sub(".*/", "", file)
        return file
    } /^```hcl/ {
    flag = 1
    ++n
    s = ""
    next
    }
    /^```$/ {
        if (flag==1) {
            printf "Extracting # %d HCL block in file %s\n", n, basename(FILENAME)
            print s > (basename(FILENAME) "-" n ".tf"); close((basename(FILENAME) "-" n ".tf"))
        }
        flag = 0
    }
    flag {
    s = s $0 ORS
    }' ../website/docs/{*/,?}*markdown
    cd $CURDIR
}

# terraform_fmt runs 
function terraform_fmt_check {
    cd tmp
    hcl_file=$1

    # OUTPUT=`terraform fmt -check -diff $hcl_file 2>&1`
    terraform fmt -check $hcl_file &>/dev/null
    retVal=$?
    if [ $retVal -ne 0 ]; then
        FOUND_ERROR=1

        echo "Error: file ${hcl_file} 'terraform fmt'"
        terraform fmt -no-color -diff -check $hcl_file 2>&1
    fi
    cd ..
}

# Check if 'website' directory is present
if [ ! -d "website" ] 
then
    echo "Expected to find 'website' directory. Please run the script from project root directory"
    exit 1
fi

extract_hcl

# Iterate over all extracted blocks and perform `terraform fmt`
for file in `ls tmp` ; do
    terraform_fmt_check ${file}
done

# If at least terraform fmt failed - return non 0 exit code
exit $FOUND_ERROR
