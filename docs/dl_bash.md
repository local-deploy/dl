## dl bash

Login to PHP container

### Synopsis

Login to PHP container as www-data or root user and start bash shell.
As the second parameter, you can specify the name or ID of another docker container.
Default is always the PHP container.

```
dl bash [flags]
```

### Examples

```
dl bash
dl bash -r
dl bash site.com_db
dl bash fcb13f1a3ea7
```

### Options

```
  -h, --help   help for bash
  -r, --root   Login as root
```

### Options inherited from parent commands

```
      --debug   Show more output
```

### SEE ALSO

* [dl](dl.md)     - Deploy Local

