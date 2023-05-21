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

## Documentation [local-deploy.github.io](https://local-deploy.github.io/)

## Supported

Supported OS: Linux (debian, ubuntu, linux mint), macOS (tested)  
Supported architectures: x64, arm64  
Supported frameworks and CMS: Bitrix, Laravel, WordPress

## Dependencies

- docker
- docker-compose v2

The `docker compose` (as plugin) supported

## Install

### Install using the apt repository (recommended)

> Only for debian-like operating systems: debian, ubuntu, linux mint, etc

#### Uninstall old versions

```bash
cd ~ && rm -rf .local/bin/dl .config/dl
```

#### Set up the repository

Before you install DL for the first time on a new host machine, you need to set up the DL repository. Afterward, you can install and update DL from the repository.

1. Update the apt package index and install packages to allow apt to use a repository over HTTPS:
    ```bash
    sudo apt update
    sudo apt install ca-certificates curl gnupg
    ```
2. Add DL’s official GPG key:
    ```bash
    sudo install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://apt.fury.io/local-deploy/gpg.key | sudo gpg --dearmor -o /etc/apt/keyrings/dl.gpg
    sudo chmod a+r /etc/apt/keyrings/dl.gpg
    ```
3. Use the following command to set up the repository:
    ```bash
    echo "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/dl.gpg] https://apt.fury.io/local-deploy/ /" | sudo tee /etc/apt/sources.list.d/dl.list > /dev/null
    ```
4. Update the apt package index and install the latest version DL:
    ```bash
    sudo apt update
    sudo apt install dl
   
    dl version
    ```

### Install locally via bash script

> For Linux and MacOS

The script will check the dependencies, download and install the latest release.

The executable file `dl` will be written to the user's home directory under the path `~/.local/bin/dl`. If the directory does not exist, it will be created.

This installation method is simpler and does not require root access, but limits the functionality of the application.

```bash
curl -s https://raw.githubusercontent.com/local-deploy/dl/master/install_dl.sh | bash
```

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

[See](docs/dl.md) quick reference for available commands.

#### Local service links

http://portainer.localhost  
http://traefik.localhost  
http://mail.localhost
