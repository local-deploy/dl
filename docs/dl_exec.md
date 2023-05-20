## dl exec

Executing a command in a PHP container

### Synopsis

Running bash command in PHP container as user www-data

```
dl exec [command] [flags]
```

### Examples

```
dl exec composer install
dl exec "ls -la"
```

### Options

```
  -h, --help   help for exec
```

### Options inherited from parent commands

```
      --debug   Show more output
```

### SEE ALSO

* [dl](dl.md)     - Deploy Local

