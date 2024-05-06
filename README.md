<p align="center">
  <img alt="DL Logo" src="https://avatars.githubusercontent.com/u/92750175?v=4&s=200" height="140" />
  <h3 align="center">Deploy Local — site deployment assistant locally.</h3>
  <p align="center">A convenient wrapper over docker-compose, which simplifies the local deployment of the project.</p>
  <p align="center">
    <a href="https://github.com/local-deploy/dl/releases/latest"><img alt="Release" src="https://img.shields.io/github/release/local-deploy/dl.svg?style=for-the-badge"></a>
    <a href="/LICENSE"><img alt="Software License" src="https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge"></a>
    <a href="https://github.com/local-deploy/dl/actions?workflow=release"><img alt="GitHub Actions" src="https://img.shields.io/github/actions/workflow/status/local-deploy/dl/.github/workflows/release.yml?style=for-the-badge"></a>
    <a href="https://goreportcard.com/report/github.com/local-deploy/dl"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/local-deploy/dl?style=for-the-badge"></a>
    <a href="http://godoc.org/github.com/local-deploy/dl"><img alt="Go Doc" src="https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge"></a>
  </p>
</p>

Deploy Local — is a command line interface designed to help developers quickly deploy projects to their local machine.  
DL is a wrapper on top of Docker and docker-compose, in basic use no additional software or libraries are required to be installed.

## Documentation [local-deploy.github.io](https://local-deploy.github.io/)

![cast](docs/dl.gif)

## Supported

Supported OS: Linux, macOS, Windows (via WSL2)  
Supported architectures: x64, arm64  
Supported frameworks and CMS: Bitrix, Laravel, WordPress, and many others with manual settings

## Features

- Support for PHP versions (apache and php-fpm) 7.3, 7.4, 8.0, 8.1, 8.2, 8.3
- Support for MySQL, MariaDB and PostgreSQL
- Downloading the database and files from the production server
- Redis
- Memcached
- Nginx
- Cross-platform
- Interception of mail sent via php
- Portainer - docker container management system
- Does not require root access (when installing the executable file in the user's directory)
- Accessing sites from the browser via .localhost or .nip.io
- Ability to add custom docker-compose.yaml files to DL configuration

## Dependencies

- docker (more than v22)
- docker-compose v2

The `docker compose` (as plugin) supported

## Install

Choose the installation method that suits you:

- [Installing the deb-package](https://local-deploy.github.io/getting-started/install#installing-the-deb-package)
- [Installation in the user's home directory](https://local-deploy.github.io/getting-started/install#installation-in-the-users-home-directory)
- [Manual installation](https://local-deploy.github.io/getting-started/install#manual-installation)

See all available installation methods https://local-deploy.github.io/getting-started/install

## Upgrading from version 0.* to version 1.*

Just reinstall the application using one of the above methods. Updating using the built-in commands is not possible.

## Usage

1. Start service containers (traefik, mailhog, portainer) with the command (at the first start):

    ```bash
    dl service up
    ```

2. Create `.env` file in the root directory of your project with the command:

    ```bash
    dl env
    ```
3. Set the required variables in the `.env` file
4. Run the command if you need to download the database and/or files from the production-server:

    ```bash
    dl deploy
    ```
5. Run local project with command:

    ```bash
    dl up
    ```

### [See](docs/dl.md) quick reference for available commands.

---

> Remember that dl is meant for development purposes, not production.
