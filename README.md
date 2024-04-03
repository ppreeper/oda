# oda - Odoo Administration Tools

These tools allow quick Odoo development on a local machine. Was build with a linux desktop in mind, but can be used with WSL systems.

## Requirements

- [incus](https://linuxcontainers.org/incus/)
- [caddy](https://caddyserver.com/)
- [sshconfig](https://github.com/ppreper/sshconfig)
- odoo community repository
- odoo enterprise repository
- [go](https://go.dev/)

### go

To install the latest version of `go` run the following command:

```bash
wget -qcO- https://raw.githubusercontent.com/ppreeper/gup/main/gup | sudo bash -s - install go
```

## Installation

```bash
go install -u github.com/ppreeper/oda/cmd/oda@latest
```

## CLI Usage

### Command: `oda`

| command   | Description                                      |
| --------- | ------------------------------------------------ |
| create    | Create the instance                              |
| destroy   | Destroy the instance                             |
| rebuild   | Rebuild the instance                             |
| start     | Start the instance                               |
| stop      | Stop the instance                                |
| restart   | Restart the instance                             |
| ps        | List Odoo Instances                              |
| logs      | Follow the logs                                  |
| exec      | Access the shell                                 |
| psql      | Access the instance database                     |
| scaffold  | Generates an Odoo module skeleton in addons      |
| query     | Query an Odoo model                              |
| backup    | Backup database filestore and addons             |
| restore   | Restore database and filestore or addons         |
| init      | initialize oda setup                             |
| hostsfile | Update /etc/hosts file (Requires root access)    |
| help, h   | Shows a list of commands or help for one command |

### Subcommands

#### `admin` Admin user management

| command  | description         |
| -------- | ------------------- |
| username | Odoo Admin username |
| password | Odoo Admin password |

#### `app` app management

| command           | description       |
| ----------------- | ----------------- |
| install <modules> | Install module(s) |
| upgrade <modules> | Upgrade module(s) |

#### `base` Base Image Management

| command | description                |
| ------- | -------------------------- |
| create  | Create base image          |
| destroy | Destroy base image         |
| rebuild | Rebuild base image         |
| update  | Update base image packages |

#### `config` additional config options

| command | description                                 |
| ------- | ------------------------------------------- |
| vscode  | Setup vscode settings and launch json files |
| pyright | Setup pyright settings                      |

#### `db` Access postgresql

| command   | description        |
| --------- | ------------------ |
| psql      | database psql      |
| start     | database start     |
| stop      | database stop      |
| restart   | database restart   |
| fullreset | database fullreset |

#### `project` Project level commands [CAUTION]

| command | description                  |
| ------- | ---------------------------- |
| init    | initialize project directory |
| branch  | initialize branch of project |
| rebuild | rebuild from another project |
| reset   | reset project dir and db     |

#### `repo` Odoo community and enterprise repository management

| command | description            |
| ------- | ---------------------- |
| base    | Odoo Source Repository |
| branch  | Odoo Source Branch     |

| command     | description                   |
| ----------- | ----------------------------- |
| base clone  | clone Odoo source repository  |
| base update | update Odoo source repository |

| command       | description                   |
| ------------- | ----------------------------- |
| branch clone  | clone Odoo branch repository  |
| branch update | update Odoo branch repository |
