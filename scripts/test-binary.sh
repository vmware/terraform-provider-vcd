#!/bin/bash

runtime_dir=$(dirname $0)
cd $runtime_dir
runtime_dir=$PWD
cd -
pause_file=$runtime_dir/pause
dash_line="# ---------------------------------------------------------"

# env script will only run if explicitly called.
build_script=cust.full-env.tf
unset in_building
operations=(init plan apply destroy)
skipping_items=($build_script)

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
    echo "  n | names 'names list' List of file names to test [QUOTES NEEDED]"
    echo "  i | test-env-init      Prepares the environment in a new vCD"
    echo "  b | test-env-apply     Builds the environment in a new vCD"
    echo "  y | test-env-destroy   Destroys the environment built using 'test-env-apply'"
    echo "  d | dry                Dry-run: show commands without executing them"
    echo "  v | verbose            Gives more info"
    echo ""
    echo "If no options are given, it runs all the vcd*.tf tests in test-artifacts"
    echo ""
    echo "Examples"
    echo "test-binary.sh tags 'catalog gateway' clear pause"
    echo "test-binary.sh t 'catalog gateway' c p"
    echo "test-binary.sh tags vapp dry"
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

while [ "$1" != "" ]
do
  opt=$1
  case $opt in
    h|help)
        get_help
        ;;
    p|pause)
        will_pause=1
        echo "will pause"
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
    exit $exit_code
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
    skip_request=$(grep '^\s*#\s*skip-binary-test' $CF)
    if [ -n "$skip_request" ]
    then
        echo "# $CF skipped ($file_count of $how_many)"
        echo "$skip_request"
        continue
    fi
    unset will_skip
    for skip_file in ${skipping_items[*]}
    do
        if [  "$CF" == "$skip_file" ]
        then
            will_skip=1
        fi
    done
    if [ -n "$will_skip" ]
    then
        echo "# $CF skipped ($file_count of $how_many)"
        continue
    fi
    is_provider=$(grep '^\s*provider' $CF)
    is_resource=$(grep '^\s*resource' $CF)
    has_missing_fields=$(grep '"\*\*\* MISSING FIELD' $CF)
    if [ -z "$is_resource" ]
    then
        echo_verbose "$CF not a resource"
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
    init_options=$(grep '^# init-options' $CF | sed -e 's/# init-options //')
    plan_options=$(grep '^# plan-options' $CF | sed -e 's/# plan-options //')
    apply_options=$(grep '^# apply-options' $CF | sed -e 's/# apply-options //')
    destroy_options=$(grep '^# destroy-options' $CF | sed -e 's/# destroy-options //')
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
    fi
    if [ -z "$DRY_RUN" ]
    then
        echo $CF >> already_run.txt
    fi
    cd $opsdir

    for phase in ${operations[*]}
    do
        case $phase in
            init)
                run terraform init $init_options
                ;;
            plan)
                run terraform plan $plan_options
                ;;
            apply)
                run terraform apply -auto-approve $apply_options
                ;;
            destroy)
                if [ ! -f terraform.tfstate ]
                then
                    echo "terraform.tfstate not found - aborting"
                    exit 1
                fi
                run terraform destroy -auto-approve $destroy_options
                ;;
        esac
    done
    cd ..
done

summary

