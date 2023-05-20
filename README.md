<p align="center">
  <img alt="DL Logo" src="https://avatars.githubusercontent.com/u/92750175?v=4&s=200" height="140" />
  <h3 align="center">Deploy Local — site deployment assistant locally.</h3>
  <p align="center">A convenient wrapper over docker-compose, which simplifies the local deployment of the project.</p>
  <p align="center">
    <a href="https://github.com/local-deploy/dl/releases/latest"><img alt="Release" src="https://img.shields.io/github/release/local-deploy/dl.svg?style=for-the-badge"></a>
    <a href="/LICENSE.md"><img alt="Software License" src="https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge"></a>
    <a href="https://github.com/local-deploy/dl/actions?workflow=release"><img alt="GitHub Actions" src="https://img.shields.io/github/actions/workflow/status/local-deploy/dl/.github/workflows/release.yml?style=for-the-badge"></a>
    <a href="https://goreportcard.com/report/github.com/local-deploy/dl"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/local-deploy/dl?style=for-the-badge"></a>
    <a href="http://godoc.org/github.com/local-deploy/dl"><img alt="Go Doc" src="https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge"></a>
  </p>
</p>

## Documentation [local-deploy.github.io](https://local-deploy.github.io/)

## Supported

Supported OS: Linux (debian, ubuntu, linux mint), macOS (tested)  
Supported architectures: x64, arm64  
Supported frameworks and CMS: Bitrix, Laravel, WordPress

## Dependencies

- docker
- docker-compose v2

The `docker compose` (as plugin) supported

## Development status

Beta version

## Install

```bash
curl -s https://raw.githubusercontent.com/local-deploy/dl/master/install_dl.sh | bash
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
