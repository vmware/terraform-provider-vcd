#!/bin/bash

runtime_dir=$(dirname $0)
cd $runtime_dir
runtime_dir=$PWD
cd -
pause_file=$runtime_dir/pause.txt
dash_line="# ---------------------------------------------------------"

function get_help {
    echo "Syntax: $(basename $0) [options]"
    echo "Where options are one or more of the following:"
    echo "  h | help               Show this help and exit"
    echo "  t | tags 'tags list'   Sets the tags to use"
    echo "  c | clear              Clears list of run files"
    echo "  p | pause              Pause after each stage"
    echo "  d | dry                Dry-run: show commands without executing them"
    echo "  v | verbose            Gives more info"
    echo ""
    echo "If no options are given, it runs all the tests in test-artifacts"
    echo ""
    echo "Examples"
    echo "test-binary.sh tags 'catalog gateway' clear pause"
    echo "test-binary.sh t 'catalog gateway' c p"
    echo "test-binary.sh tags vapp dry"
    echo ""
    echo "## During the execution, if you create a file named 'pause.txt',"
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

test_names='*.tf'

function summary {
    end_time=$(date +%s)
    end_timestamp=$(date)
    elapsed=$(($end_time-$start_time))
    echo "# Started:   $start_timestamp"
    echo "# Ended:     $end_timestamp"
    echo "# Elapsed:   $elapsed"
    echo "# exit code: $exit_code"
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

for CF in $test_names
do
    is_provider=$(grep ^provider $CF)
    is_resource=$(grep ^resource $CF)
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
    echo "# $CF"
    echo $dash_line
    if [ -d ./tmp ]
    then
        rm -rf tmp
    fi
    mkdir tmp
    cp $CF tmp/config.tf
    if [ -z "$DRY_RUN" ]
    then
        echo $CF >> already_run.txt
    fi
    cd tmp

    run terraform init
    run terraform plan
    run terraform apply -auto-approve
    run terraform destroy -auto-approve
    cd ..
done

summary

