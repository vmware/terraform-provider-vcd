#!/bin/bash

SYNTAX_ERRORS=0
SEMANTIC_ERRORS=0

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
            print s > (basename(FILENAME) "-" n ".tf"); close((basename(FILENAME) "-" n ".tf"))
        }
        flag = 0
    }
    flag {
    s = s $0 ORS
    }' ../website/docs/{*/,?}*markdown
    cd $CURDIR
}

# Runs the fmt subcommand of Terraform to check syntax errors on temporary HCL files
# that contain the documentation snippets.
function terraform_fmt_check {
    hcl_file="$1"

    # OUTPUT=`terraform fmt -check -diff $hcl_file 2>&1`
    terraform fmt -check $hcl_file &>/dev/null
    retVal=$?
    if [ $retVal -ne 0 ]; then
        ((SYNTAX_ERRORS++))

        echo "(!) ERROR: File ${hcl_file} contains syntax errors! (terraform fmt)"
        terraform fmt -no-color -diff -check $hcl_file 2>&1
    fi
}

# Performs a Terraform init on a temporary HCL file that contain the documentation snippets.
# If the init command fails, gives an error message and the script will fail.
function terraform_validation_check {
    hcl_file="$1"
    folder="$(dirname $hcl_file)"
    initial_dir="$(pwd)"

    # Create a temporary folder called validate where we put the file to test and the provider definition file.
    if [ ! -d "$folder/validate" ]
    then
        mkdir "$folder/validate"
    fi
    cp "$hcl_file" "$folder/validate/current.tf"
    cd "$folder/validate" || exit 1

    echo "
terraform {
  required_providers {
    vcd = {
      source  = \"vmware/vcd\"
      version = \"$(git describe --abbrev=0 --tags | cut -d'v' -f 2)\"
    }
    nsxt = {
      source = \"vmware/nsxt\"
    }
  }
  required_version = \">= 0.13\"
}" > provider_setup.tf

    terraform init -no-color > /dev/null
    retVal=$?
    if [ $retVal -ne 0 ]; then
        ((SEMANTIC_ERRORS++))
        echo "(!) ERROR: File ${hcl_file} contains semantic errors! (terraform init)"
    fi

    rm -f current.tf # We don't remove the provider so we don't download it everytime
    cd "$initial_dir" || exit 1

}

# Check if 'website' directory is present
if [ ! -d "website" ] 
then
    echo "Expected to find 'website' directory. Please run the script from project root directory"
    exit 1
fi

rm -rf tmp
extract_hcl

# Iterate over all extracted blocks and perform `terraform fmt`
for file in tmp/*.tf ; do
    terraform_fmt_check "${file}"
done

# Iterate over all extracted blocks and perform `terraform init`
for file in tmp/*.tf ; do
    terraform_validation_check "${file}"
done

echo "------------------------------
    Summary:

    * Syntax errors: $SYNTAX_ERRORS
    * Semantic errors $SEMANTIC_ERRORS
------------------------------"

# If at least terraform fmt failed - return non 0 exit code
if [[ $SYNTAX_ERRORS = 0 ]] && [[ $SEMANTIC_ERRORS = 0 ]]
then
    echo 'Finished SUCCESSFULLY!'
    exit 0;
fi
echo 'Finished with FAILURES!'
exit 1