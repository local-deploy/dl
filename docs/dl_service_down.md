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
  -h, --help             help for down
  -s, --service string   Stop and remove single service
```

### SEE ALSO

* [dl service](dl_service.md)	 - Local services configuration

