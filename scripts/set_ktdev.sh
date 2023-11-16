# Add alias to compiled CLI
alias ktdev="$(pwd)/cli/cli/scripts/launch-cli.sh" 

# Setup bash completion
source <(ktdev completion bash)
complete -F __start_kurtosis ktdev