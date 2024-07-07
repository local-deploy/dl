## dl service down

Stop and remove services

### Synopsis

Stops and removes portainer, mailcatcher and traefik containers.
Valid parameters for the "--service" flag: portainer, mail, traefik

```
dl service down [flags]
```

### Examples

```
dl down
dl down -s portainer
```

### Options

```
  -h, --help   help for down
```

### Options inherited from parent commands

```
      --debug              Show more output
  -s, --services strings   Manage only specified services (comma separated values)
```

### SEE ALSO

* [dl service](dl_service.md)     - Local services configuration

