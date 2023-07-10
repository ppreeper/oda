# oda - Odoo Administration Tools

These tools allow quick Odoo development on a local machine. Was build with a linux desktop in mind, but can be used with WSL systems.

## Requirements

- podman (for postgres container)
- odoo community repository
- odoo enterprise repository
- odooquery (for querying of Odoo instance)

### odooquery

`odooquery` requires `go` as its base language. This must be installed first to install `odooquery`.

To install `odooquery` run the following command:

```bash
go install github.com/ppreeper/odooquery@latest
```

To install the latest version of `go` run the following command:

```bash
wget -qcO- https://raw.githubusercontent.com/ppreeper/gup/main/gup | sudo bash -s - install go
```

## Installation

After requirements are installed. Copy the files to your path, the oda and oda_db.py need to be in the same directory.

## CLI Usage

### Command: `db`

#### DB Admin Commands

|           |                                    |
| --------- | ---------------------------------- |
| fullreset | Fully wipe all databases [CAUTION] |
| start     | Start the instance                 |
| stop      | Stop the instance                  |
| restart   | Restart the instance               |
| logs      | Follow the logs                    |
| stats     | Get POD stats                      |
| top       | POD top command                    |
| psql      | Access the raw database            |

### Command: `oda`

### Requirements

- gem install bashly

#### Project Admin Commands

|                     |                                                   |
| ------------------- | ------------------------------------------------- |
| initproject         | Create a new project                              |
| destroy             | Fully Destroy the project and its files [CAUTION] |
| reset               | Drop database and filestore [CAUTION]             |
| backup              | Backup database and filestore [CAUTION]           |
| restore <dump_file> | Restore database and filestore [CAUTION]          |

#### Database Application Commands

|                   |                                          |
| ----------------- | ---------------------------------------- |
| init              | Initialize the database                  |
| install <modules> | Install module(s) (comma seperated list) |
| upgrade <modules> | Upgrade module(s) (comma seperated list) |

#### Database Admin

|               |                         |
| ------------- | ----------------------- |
| start         | Start the instance      |
| stop          | Stop the instance       |
| restart       | Restart the instance    |
| logs          | Follow the logs         |
| bin <command> | Run an odoo-bin command |
| psql          | Access the raw database |
