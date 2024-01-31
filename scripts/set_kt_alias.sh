# Add alias to compiled CLI
alias ktdev="$(pwd)/cli/cli/scripts/launch-cli.sh"
alias ktdebug="$(pwd)/cli/cli/scripts/debug-cli.sh"

CURRENT_SHELL=$(sh -c 'ps -p $$ -o ppid=' | xargs ps -o comm= -p) 
SHELL_NAME=$(basename -- $CURRENT_SHELL)

echo "Detected shell: $SHELL_NAME"

case "$SHELL_NAME" in

bash)  echo "Setting $SHELL_NAME completion"
    source <(ktdev completion bash)
    complete -F __start_kurtosis ktdev
    ;;
zsh)  echo "Setting $SHELL_NAME completion"
    source <(ktdev completion zsh)
    compdef __start_kurtosis ktdev
    ;;
*) echo "Shell $SHELL_NAME is not supported"
   ;;
esac
