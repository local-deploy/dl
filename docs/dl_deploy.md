## dl deploy

Downloading db and files from the production server

### Synopsis

Downloading database and kernel files from the production server.
Without specifying the flag, files and the database are downloaded by default.
If you specify a flag, for example -d, only the database will be downloaded.

Directories that are downloaded by default
Bitrix CMS: "bitrix"
WordPress: "wp-admin" and "wp-includes"
Laravel: only the database is downloaded

```
dl deploy [flags]
```

### Examples

```
dl deploy
dl deploy -d
dl deploy -f -o bitrix,upload
```

### Options

```
  -d, --database           Dump only database from server
  -f, --files              Download only files from server
  -h, --help               help for deploy
  -o, --override strings   Override downloaded files (comma separated values)
```

### Options inherited from parent commands

```
      --debug   Show more output
```

### SEE ALSO

* [dl](dl.md)     - Deploy Local

