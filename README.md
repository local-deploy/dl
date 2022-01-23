Deploy Local — site deployment assistant locally.

A convenient wrapper over docker-compose, which simplifies the local deployment of the project.

Supported OS: Linux (debian, ubuntu, linux mint), macOS (not tested)  
Supported architectures: x64

Dependencies:

- docker
- docker-compose

### Documentation [local-deploy.github.io](https://local-deploy.github.io/)

## Development status

Beta version

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
