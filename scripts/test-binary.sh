#!/bin/bash

runtime_dir=$(dirname $0)
cd $runtime_dir
runtime_dir=$PWD
cd -
pause_file=$runtime_dir/pause
dash_line="# ---------------------------------------------------------"
export upgrading=""

version_file=../../VERSION

if [ ! -f $version_file ]
then
  echo "version file ${version_file} not found"
  exit 1
fi
provider_version=$(cat $version_file | sed -e 's/^v//' | tr -d ' \t\n')
# env script will only run if explicitly called.
build_script=cust.full-env.tf
unset in_building
operations=(init plan apply plancheck destroy)
skipping_items=($build_script)
failed_tests=0

if [ -f failed_tests.txt ]
then
    rm -f failed_tests.txt
fi

if [ -f skip-files.txt ]
then
    for f in $(cat skip-files.txt)
    do
        skipping_items+=($f)
    done
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

# Terraform version is used to determine what the target directory should be
terraform_version=$(terraform version | head -n 1| sed -e 's/Terraform v//')
check_empty "$terraform_version" "terraform_version not detected"

terraform_major=$(echo $terraform_version | tr '.' ' '| awk '{print $1}')
check_empty "$terraform_major" "terraform_version major not detected"
terraform_minor=$(echo $terraform_version | tr '.' ' '| awk '{print $2}')
check_empty "$terraform_minor" "terraform_version minor not detected"

function terraform_binary_path {
    version=$1
    tversion_major=$2
    tversion_minor=$3
    if [ -z "$tversion_major" ]
    then
        tversion_major=$terraform_major
    fi
    if [ -z "$tversion_minor" ]
    then
        tversion_minor=$terraform_minor
    fi
    # The default target directory is the pre-0.13 one
    target_dir=.terraform.d/plugins/
    # Getting the version without the initial "v"
    bare_version=$(echo $version | sed -e 's/^v//')
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    check_empty "$os" "operating system not detected"
    arch=${os}_amd64

    # if terraform executable is 0.13+, we use the new path
    if [[ $tversion_major -gt 0 || $tversion_major -eq 0 && $tversion_minor -gt 12 ]]
    then
        target_dir=.terraform.d/plugins/registry.terraform.io/vmware/vcd/$bare_version/$arch
    fi

    echo $target_dir
}

# 'custom_terraform' allows to expose terraform version for logging and execute it
function custom_terraform {
    terraform version
    terraform $*
}

function remove_item_from_skipping {
    item=$1
    new_array=()
    count=0
    for N in ${skipping_items[*]}
    do
        if [ "$N" != "$item" ]
        then
            new_array[$count]=$N
            count=$((count+1))
        fi
    done
    skipping_items=$new_array
}

function get_help {
    echo "Syntax: $(basename $0) [options]"
    echo "Where options are one or more of the following:"
    echo "  h | help               Show this help and exit"
    echo "  t | tags 'tags list'   Sets the tags to use"
    echo "  c | clear              Clears list of run files"
    echo "  p | pause              Pause after each stage"
    echo "  a | validate           Validate scripts without running tests"
    echo "  n | names 'names list' List of file names to test [QUOTES NEEDED]"
    echo "  i | test-env-init      Prepares the environment in a new vCD"
    echo "  b | test-env-apply     Builds the environment in a new vCD"
    echo "  y | test-env-destroy   Destroys the environment built using 'test-env-apply'"
    echo "  u | upgrade 'From To'  Test upgrade - 'From' is the old version and 'To' is the new one"
    echo "  d | dry                Dry-run: show commands without executing them"
    echo "  v | verbose            Gives more info"
    echo ""
    echo "If no options are given, it runs all the vcd*.tf tests in test-artifacts"
    echo ""
    echo "Examples"
    echo "test-binary.sh tags 'catalog gateway' clear pause"
    echo "test-binary.sh t 'catalog gateway' c p"
    echo "test-binary.sh tags vapp dry"
    echo "test-binary.sh tags vm upgrade v2.5.0 v2.6.0"
    echo ""
    echo "test-binary.sh names 'cust*.tf'"
    echo "test-binary.sh names cust.demo.tf pause"
    echo ""
    echo "## During the execution, if you create a file named 'pause',"
    echo "## the program will pause at the next 'terraform' command"
    exit 0
}

function echo_verbose {
    if [ -n "$VERBOSE" ]
    then
        echo "$@"
    fi
}

function add_versions_config {

    fname=$1
    dir=$2
    has_terraform=$(grep '^\s*terraform {' $fname)
    cat << EOF > $dir/versions.tf
terraform {
  required_providers {
    vcd = {
      source  = "vmware/vcd"
      version = "~> $provider_version"
    }
  }
  # required_version = ">= 0.13"
}

EOF
}


function check_exit_code {
    out=$1
    if [ "$exit_code" != "0" ]
    then
        cat $out
        exit $exit_code
    fi
}

function validate_script {
    script=$1
    echo "## $script ##"
    if [ -d vtmp ]
    then
        rm -rf vtmp
    fi
    mkdir vtmp
    cp $script vtmp
    add_versions_config $script vtmp
    cd vtmp
    terraform init > init.out 2>&1
    exit_code=$?
    check_exit_code init.out
    terraform validate > validate.out 2>&1
    exit_code=$?
    check_exit_code validate.out
    cd - > /dev/null
    rm -rf vtmp
}


while [ "$1" != "" ]
do
  opt=$1
  case $opt in
    h|help)
        get_help
        ;;
    p|pause)
        if [ -n "$validating" ]
        then
            echo "pause not available for validation"
        else
            will_pause=1
            echo "will pause"
        fi
        ;;
    c|clear)
        if [ -f already_run.txt ]
        then
            rm -f already_run.txt
            echo "already_run.txt removed"
        fi
        ;;
    d|dry)
        DRY_RUN=1
        ;;
    t|tags)
        shift
        opt=$1
        if [ -z "$opt" ]
        then
            echo "option 'tags' requires an argument"
            exit 1
        fi
        tags="$opt"
        echo "tags: $tags"
        ;;
    u|upgrade)
        shift
        from_version=$1
        shift
        to_version=$1
        upgrading=1
        if [ -z "$to_version" ]
        then
            echo "option 'upgrade' requires two arguments (such as 'v2.5.0' 'v2.6.0')"
            exit 1
        fi
        ;;
    n|names)
        shift
        opt=$1
        test_names="$opt"
        if [ -z "$opt" ]
        then
            echo "option 'names' requires an argument"
            exit 1
        fi
        ;;
    i|test-env-init)
        remove_item_from_skipping $build_script
        operations=(init plan)
        in_building=yes
        test_names="$build_script"
        ;;
    b|test-env-apply)
        remove_item_from_skipping $build_script
        operations=(plan apply)
        in_building=yes
        test_names="$build_script"
        ;;
    y|test-env-destroy)
        remove_item_from_skipping $build_script
        rm -f already_run.txt
        operations=(plan destroy)
        in_building=yes
        test_names="$build_script"
        ;;
    v|verbose)
        export VERBOSE=1
        ;;
    validate)
        validating=1
        if [ -n "$will_pause" ]
        then
            echo "pause disabled for validation"
            unset will_pause
        fi
        ;;
    *)
        get_help
        ;;
  esac
  shift
done

exit_code=0
start_time=$(date +%s)
start_timestamp=$(date)

[ -z  "$test_names" ] && test_names='vcd.*.tf'

function summary {
    end_time=$(date +%s)
    end_timestamp=$(date)
    secs=$(($end_time-$start_time))
    minutes=$((secs/60))
    remainder_sec=$((secs-minutes*60))
    if [[ $minutes -lt 60 ]]
    then
        elapsed=$(printf "%dm:%02ds" ${minutes} ${remainder_sec})
    else
        hours=$((minutes/60))
        remainder_minutes=$((minutes-hours*60))
        elapsed=$(printf "%dh:%dm:%02ds" ${hours} ${remainder_minutes} ${remainder_sec})
    fi
    echo "$dash_line"
    echo "# Operations dir: $runtime_dir/$opsdir"
    echo "# Started:        $start_timestamp"
    echo "# Ended:          $end_timestamp"
    echo "# Elapsed:        $elapsed ($secs sec)"
    echo "# exit code:      $exit_code"
    echo "$dash_line"
    if [ "$failed_tests" != 0 ]
    then
        echo "$dash_line"
        echo "# FAILED TESTS    $failed_tests"
        echo "$dash_line"
        cat $runtime_dir/failed_tests.txt
        echo "$dash_line"
    fi
    exit $exit_code
}

function interactive {
    echo "Paused at user request - Press ENTER when ready"
    echo "(Enter 'q' to exit with 0, 'x' to exit with 1, 'c' to continue without further pause)"
    read answer
    case $answer in
        q)
            echo "Exit at user request"
            exit 0
            ;;
        x)
            echo "Exit with non-zero code at user request"
            exit 1
            ;;
        c)
            unset will_pause
            echo "Execution will not pause any more"
            ;;
    esac
}

function run {
    cmd="$@"
    echo "# $cmd"
    if [ -n "$DRY_RUN" ]
    then
        return
    fi
    $cmd
    exit_code=$?
    if [ "$exit_code" != "0" ]
    then
        echo "EXECUTION ERROR"
        summary
    fi
    if [ -f $pause_file ]
    then
        rm -f $pause_file
        export will_pause=1
    fi
    if [ -n "$will_pause" ]
    then
        interactive
    fi
}

function run_with_recover {
    test_name=$1
    phase=$2
    shift
    shift
    cmd="$@"
    echo "# $cmd"
    if [ -n "$DRY_RUN" ]
    then
        return
    fi
    $cmd
    exit_code=$?
    if [ "$exit_code" != "0" ]
    then
        echo "$(date) - $test_name ($phase)" >> $runtime_dir/failed_tests.txt
        failed_tests=$((failed_tests+1))
        case $phase in
            init)
                # An error on initialization should not be recoverable
                echo $dash_line
                echo "NON-RECOVERABLE EXECUTION ERROR (phase: $phase)"
                echo $dash_line
                summary
                ;;
            plan | plancheck)
                # an error in plan does not need any recovery,
                # in addition to recording the file in the failed list
                # The destroy will be called anyway
                echo $dash_line
                echo "RECOVERING FROM plancheck phase. A 'destroy' will run next"
                echo $dash_line
                ;;
            apply | destroy)

                # errors in apply should be recoverable after a destroy

                # an error in destroy means we are leaving behind hanging entities.
                # nonetheless we try to recover with an additional destroy

                echo $dash_line
                echo "# ATTEMPTING RECOVERY AFTER FAILURE (phase $phase - exit code $exit_code)"
                echo $dash_line
                run terraform destroy -auto-approve
                # if 'run' doesn't produce an error, we continue the tests,
                # leaving behind the name of the test failed in failed_tests.txt
                # otherwise, the test is definitely terminated
                ;;

            *)
                echo $dash_line
                echo "unhandled phase in recovery'$phase'"
                echo $dash_line
                exit $exit_code
                ;;

        esac
    fi
    if [ -f $pause_file ]
    then
        rm -f $pause_file
        export will_pause=1
    fi
    if [ -n "$will_pause" ]
    then
        interactive
    fi
}


if [ ! -f already_run.txt ]
then
    touch already_run.txt
fi

how_many=$(ls $test_names | wc -l)
file_count=0
for CF in $test_names
do
    file_count=$((file_count+1))
    if [ -n "$upgrading" -a -f skip-upgrade-tests.txt ]
    then
        skip_upgrade_request=$(grep "$from_version" skip-upgrade-tests.txt | grep $CF)
        if [ -n "$skip_upgrade_request" ]
        then
            echo "# $CF skipped ($file_count of $how_many)"
            echo "$skip_upgrade_request"
            continue
        fi
    fi
    skip_request=$(grep '^\s*#\s*skip-binary-test' $CF)
    if [ -n "$skip_request" -a -z "$validating" ]
    then
        echo "# $CF skipped ($file_count of $how_many)"
        echo "$skip_request"
        continue
    fi
    unset will_skip
    for skip_file in ${skipping_items[*]}
    do
        if [  "$CF" == "$skip_file" -a -z "$validating" ]
        then
            will_skip=1
        fi
    done
    if [ -n "$will_skip" ]
    then
        echo "# $CF skipped ($file_count of $how_many)"
        continue
    fi
    is_provider=$(grep '^\s*provider\>' $CF)
    is_resource=$(grep '^\s*resource\>' $CF)
    is_data_source=$(grep '^\s*data\>' $CF)
    has_missing_fields=$(grep '"\*\*\* MISSING FIELD' $CF)
    if [ -z "$is_resource" -a -z "$is_data_source" ]
    then
        echo_verbose "$CF not a resource or data source"
        continue
    fi
    if [ -z "$is_provider" ]
    then
        echo_verbose "$CF does not contain a provider"
        continue
    fi
    if [ -n "$has_missing_fields" ]
    then
        echo "# $dash_line"
        echo "# Missing fields in $CF"
        echo "# $dash_line"
        continue
    fi
    init_options="-compact-warnings $(grep '^# init-options' $CF | sed -e 's/# init-options //')"
    plan_options="-compact-warnings $(grep '^# plan-options' $CF | sed -e 's/# plan-options //')"
    apply_options="-compact-warnings $(grep '^# apply-options' $CF | sed -e 's/# apply-options //')"
    plancheck_options="-compact-warnings $(grep '^# plancheck-options' $CF | sed -e 's/# plancheck-options //')"
    destroy_options="-compact-warnings $(grep '^# destroy-options' $CF | sed -e 's/# destroy-options //')"
    using_tags=$(grep '^# tags' $CF | sed -e 's/# tags //')
    already_run=$(grep $CF already_run.txt)
    if [ -n "$already_run" ]
    then
        echo "$CF already run"
        continue
    fi
    unset will_run
    # No tags were requested: we will run every file
    if [ "$tags" == ""  -o "$tags" == "ALL" ]
    then
        will_run=1
    else
        for utag in $using_tags
        do
            for wtag in $tags
            do
                if [ "$utag" == "$wtag" ]
                then
                    echo_verbose "Using tag $utag"
                    will_run=1
                fi
            done
        done
    fi
    if [ -z "$will_run" ]
    then
        echo_verbose "$CF skipped for non-matching tag"
        continue
    fi
    echo $dash_line
    echo "# $CF ($file_count of $how_many)"
    echo $dash_line
    opsdir=tmp
    if [ -n "$validating" ]
    then
        validate_script $CF
        continue
    fi
    if [ -n  "$in_building" ]
    then
        url=$(grep  "^\s\+url " $build_script | awk '{print $3}' | sed -e 's/https:..//' -e 's/.api//' | tr -d '"' | tr '.' '-')
        # echo "URL<$url>"
        opsdir=test-env-build-$url
    fi
    if [ "${operations[0]}" == "init" ]
    then
        if [ -d $opsdir ]
        then
            rm -rf $opsdir
        fi
    else
        # if it is not "init", we need to find it already
        if [ ! -d $opsdir ]
        then
            echo "$PWD/$opsdir not found"
            echo "Run ./test-binary.sh test-env-init"
            exit 1
        fi
    fi
    if [ ! -d $opsdir ]
    then
        mkdir $opsdir
    fi
    if [ "${operations[0]}" == "init" ]
    then
        cp $CF $opsdir/config.tf
        add_versions_config $CF $opsdir
    fi
    if [ -z "$DRY_RUN" ]
    then
        echo $CF >> already_run.txt
    fi
    cd $opsdir


    # Prepare variables for "upgrade" testing
    if [ -n "$upgrading" ]
    then
        major_from=$(echo $from_version | tr -d 'v' | tr '.' ' '| awk '{print $1}')
        minor_from=$(echo $from_version | tr -d 'v' | tr '.' ' '| awk '{print $2}')
        major_to=$(echo $to_version | tr -d 'v' | tr '.' ' '| awk '{print $1}')
        minor_to=$(echo $to_version | tr -d 'v' | tr '.' ' '| awk '{print $2}')
        short_from="${major_from}.${minor_from}"
        short_to="${major_to}.${minor_to}"
    fi

    for phase in ${operations[*]}
    do
        case $phase in
            init)
                if [ -n "$upgrading" ]
                    then
                    # Remove any version definition in config.tf because it is deprecated
                    sed  -i -e '/^\s*version *=/d' config.tf
                    # Set exact Terraform version constraint `version = "3.0"` instead of `version = "~> 3.0"` as such
                    # constraint would still pull newer version and this is bad for upgrade tests
                    sed -i -e '/^\s*version *= ".*"/version = "'${short_from}'"/'  versions.tf
                    run terraform init
                    run terraform version
                fi


                # 'custom_terraform' can process both regular and upgrade operations
                run_with_recover $CF init custom_terraform init $init_options
                ;;
            plan)
                # 'custom_terraform' can process both regular and upgrade operations
                run_with_recover $CF plan custom_terraform plan $plan_options
                ;;
            apply)
                # 'custom_terraform' can process both regular and upgrade operations
                run_with_recover $CF apply custom_terraform apply -auto-approve $apply_options
                ;;
            plancheck)
                # Skip plan check if a `.tf` example contains line "# skip-plan-check"
                # During upgrades, 'plancheck' runs with the latest plugin
                skip_plancheck=$(grep '^\s*#\s*skip-plan-check' "../$CF")
                if [ -n "$skip_plancheck" ]
                then
                    echo "# $CF plan check skipped"
                else
                    # Explicitly change provider version and run `terraform init -upgrade`
                    if [ -n "$upgrading" ]
                    then
                        # Replace the version in HCL configuration file
                        sed -i -e 's/version *= "'${short_from}'"/version = "'${short_to}'"/' versions.tf
                        run terraform init -upgrade
                        run terraform version
                    fi

                    # -detailed-exitcode will return exit code 2 when the plan was not empty
                    # and this allows to validate if reads work properly and there is no immediate
                    # plan change right after apply succeeded
                    run_with_recover $CF plancheck terraform plan -detailed-exitcode $plancheck_options
                fi
                ;;
            destroy)
                # During upgrades, 'destroy' runs with the latest plugin
                if [ ! -f terraform.tfstate ]
                then
                    echo "terraform.tfstate not found - exiting"
                    exit 1
                fi
                [ -n "$upgrading" ] && run terraform version
                run_with_recover $CF destroy terraform destroy -auto-approve $destroy_options
                ;;
        esac
    done
    cd ..
done

summary
