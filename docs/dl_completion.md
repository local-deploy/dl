## dl completion

Generate completion script

### Synopsis

To load completions:

Bash:

$ source <(dl completion bash)

# To load completions for each session, execute once:

$ echo "source <(dl completion bash)" >> ~/.bashrc

Zsh:

# If shell completion is not already enabled in your environment,

# you will need to enable it. You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:

$ dl completion zsh > "${fpath[1]}/_dl"

You will need to start a new shell for this setup to take effect.

```
dl completion [bash|zsh]
```

### Options

```
  -h, --help   help for completion
```

### Options inherited from parent commands

```
      --debug   Show more output
```

### SEE ALSO

* [dl](dl.md)     - Deploy Local

