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

TODO

#### Local service links

http://portainer.localhost  
http://traefik.localhost  
http://mail.localhost

## Available Commands

TODO

## DEV info

systray  
https://github.com/getlantern/systray

valid ssl certificates  
https://hollo.me/devops/routing-to-multiple-docker-compose-development-setups-with-traefik.html#chapter-1-create-valid-ssl-certificates

php tags  
https://github.com/docker-library/docs/blob/master/php/README.md#supported-tags-and-respective-dockerfile-links

traefik labels  
https://doc.traefik.io/traefik/reference/dynamic-configuration/docker/

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
