#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"



# ==================================================================================================
#                                             Constants
# ==================================================================================================
# TODO Fill constants here with UPPER_SNAKE_CASE, noting that the only variables constants may use are:
# TODO  1) other constants (with the "${OTHER_CONSTANT}" syntax)
# TODO  2) script_dirpath/root_dirpath from above
DEFAULT_SOME_OPTIONAL_ARG_VALUE="A default value"   # TODO Replace with your own constants



# ==================================================================================================
#                                       Arg Parsing & Validation
# ==================================================================================================
# TODO Modify the arguments below to match the argument to your script
show_helptext_and_exit() {
    echo "Usage: $(basename "${0}") some_filepath_arg some_other_arg [some_optional_arg]"
    echo ""
    echo "  some_filepath_arg   The description of some_arg_1 goes here"
    echo "  some_other_arg      The description of some_arg_2 goes here, and if the description is really long then"
    echo "                          we break it into two lines like this"
    echo "  some_optional_arg   Optional positional argument that doesn't need to be provided (default: ${DEFAULT_SOME_OPTIONAL_ARG_VALUE})"
    echo ""
    exit 1  # Exit with an error so that if this is accidentally called by CI, the script will fail
}

# TODO Modify the arg-grabbing below, noting that:
# TODO - Non-constant variables are lower_snake_case
# TODO - The "${X:-}" syntax is necessary to pass Bash strict mode
some_filepath_arg="${1:-}"
some_other_arg="${2:-}"
some_optional_arg="${3:-"${DEFAULT_SOME_OPTIONAL_ARG_VALUE}"}"  # Note how optional arguments get a constant default value

# TODO Modify this arg validation to match your arguments, keeping in mind:
# TODO - Almost every arg should be verified to be non-empty
# TODO - Filepath/dirpath ags often need to have their existence checked
if [ -z "${some_filepath_arg}" ]; then
    echo "Error: no filepath arg provided" >&2
    show_helptext_and_exit
fi
if ! [ -f "${some_filepath_arg}" ]; then
    echo "Error: filepath arg '${some_filepath_arg}' isn't a valid file" >&2
    show_helptext_and_exit
fi
if [ -z "${some_other_arg}" ]; then
    echo "Error: some other arg is empty" >&2
    show_helptext_and_exit
fi



# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
# TODO implement your main logic here using the args & constants, keeping in mind that:
# TODO - Non-constant variables should be lower_snake_case
# TODO - Variables should be referenced like this: ${some_variable}
# TODO - Capturing subprocess output should be done using $(), NOT backticks ``, like so:
# TODO
# TODO      my_variable="$(echo "Something")"
# TODO
# TODO - Every call should have its return value checked like so:
# TODO
# TODO       if ! some_command; then 
# TODO           echo "Error: Some description of the error" >&2
# TODO           exit 1
# TODO       fi
