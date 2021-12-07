Deploy Local — site deployment assistant locally.

A convenient wrapper over docker-compose, which simplifies the local deployment of the project.

Supported OS: Linux (debian, ubuntu, linux mint), macOS (not tested)  
Supported architectures: x64

Dependencies:

- docker
- docker-compose

## Development status

Alpha version

## Install

```bash
wget --no-check-certificate https://raw.githubusercontent.com/local-deploy/dl/master/install_dl.sh && chmod +x ./install_dl.sh && ./install_dl.sh
```

## Usage

1. Start service containers (traefik, mailhog, portainer) with the `dl service up` command (at the first start)
2. Create `.env` file in the root directory of your project with the `dl env` command
3. Set the required variables in the `.env` file
4. Run the `dl deploy` command if you need to download the database and/or files from the production-server
5. Run local project with `dl up` command

[See](docs/dl.md) quick reference for available commands.

#### Local service links

http://portainer.localhost  
http://traefik.localhost  
http://mail.localhost

## Move to doc

Available env:

| env                     | default value                                       |
|-------------------------|---------------------------------------------------  |
| LOCALTIME               | Europe/Moscow                                       |
| PHP_MODULES             | opcache                                             |
| PHP_MEMORY_LIMIT        | 256M                                                |
| PHP_POST_MAX_SIZE       | 100M                                                |
| PHP_UPLOAD_MAX_FILESIZE | 100M                                                |
| PHP_MAX_FILE_UPLOADS    | 50                                                  |
| PHP_MAX_EXECUTION_TIME  | 60                                                  |
|                         |                                                     |
| SSH_KEY                 | id_rsa (from ~/.ssh/id_rsa)                         |
| NETWORK_NAME            | generated from HOST_NAME without special characters |
|                         |                                                     |
| XDEBUG                  | off                                                 |
| XDEBUG_IDE_KEY          | PHPSTORM                                            |
| XDEBUG_PORT             | 9003                                                |
|                         |                                                     |
| REDIS                   | false                                               |
| REDIS_PASSWORD          | pass                                                |
| MEMCACHED               | false                                               |
|                         |                                                     |
| LOCAL_IP                | external local IP (e.g. 192.168.0.5)                |
| NGINX_CONF              | ~/.config/dl/config-files/default.conf.template     |

Xdebug mode https://xdebug.org/docs/all_settings#mode
