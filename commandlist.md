# Command Lists

## oda dev client

NAME:
   oda - Odoo Administration Tool

USAGE:
   oda [global options] command [command options]

VERSION:
   20240731-4e3c756

COMMANDS:
   create     Create the instance
   destroy    Destroy the instance
   rebuild    Rebuild the instance
   start      Start the instance
   stop       Stop the instance
   restart    Restart the instance
   ps         List Odoo Instances
   logs       Follow the logs
   exec       Access the shell
   psql       Access the instance database
   scaffold   Generates an Odoo module skeleton in addons
   query      Query an Odoo model
   backup     Backup database filestore and addons
   restore    Restore database and filestore or addons
   init       initialize oda setup
   hostsfile  Update /etc/hosts file (Requires root access)
   help, h    Shows a list of commands or help for one command
   admin:
     admin  Admin user management
   app:
     app  app management
   base:
     base  Base Image Management
   config:
     config  additional config options
   db:
     db  Access postgresql
   project:
     project  Project level commands [CAUTION]
   repo:
     repo  Odoo community and enterprise repository management

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
(python-3.10.12)

## oda.py

def main():
    """Odoo Server Administration Tool"""
    parser = ArgParser(
        prog="oda",
        description="Odoo Server Administration Tool",
        epilog="thanks for using %(prog)s!",
    )
    subparsers = parser.add_subparsers(
        dest="command", title="commands", help="commands"
    )

    # backup         Backup database filestore and addons
    subparsers.add_parser("backup", help="Backup database filestore and addons")

    # restore        Restore database and filestore or addons
    restore_parser = subparsers.add_parser(
        "restore", help="Restore database and filestore or addons"
    )
    restore_parser.add_argument("file", help="Path to backup file", nargs="+")
    restore_parser.add_argument(
        "--move", help="move database", action="store_true", default=False
    )

    # trim           Trim database backups
    subparsers.add_parser("trim", help="Trim database backups")
    subparsers.add_parser("trimall", help="Trim all database backups")

    # Modules
    # install        Install module(s)
    install_parser = subparsers.add_parser("install", help="Install module(s)")
    install_parser.add_argument("module", help="module(s) to install", nargs="+")
    # upgrade        Upgrade module(s)
    upgrade_parser = subparsers.add_parser("upgrade", help="Upgrade module(s)")
    upgrade_parser.add_argument("module", help="module(s) to upgrade", nargs="+")

    # scaffold        Scaffold module
    scaffold_parser = subparsers.add_parser("scaffold", help="Scaffold module")
    scaffold_parser.add_argument("module", help="module to scaffold", nargs="*")

    # server control
    subparsers.add_parser("start", help="start odoo server")
    subparsers.add_parser("stop", help="top odoo server")
    subparsers.add_parser("restart", help="restart odoo server")

    # hosts          update hosts
    hosts_parser = subparsers.add_parser("hosts", help="update hosts file")
    hosts_parser.add_argument("domain", help="domain", nargs="*")

    # caddy          update hosts
    caddy_parser = subparsers.add_parser("caddy", help="update caddy file")
    caddy_parser.add_argument("domain", help="domain", nargs="*")

    # logs           tail the logs
    subparsers.add_parser("logs", help="tail the logs")

    # config
    parser.add_argument(
        "-c",
        "--config",
        action="store",
        default="/opt/odoo/conf/odoo.conf",
        help="odoo.conf file location",
    )
    parser.add_argument(
        "-d",
        action="store",
        default="/opt/odoo/backups",
        help="backup directory",
    )
    parser.add_argument(
        "-n",
        action="store",
        default=10,
        help="number of backups to keep",
    )
    # ===================
    # process arguments
    args = parser.parse_args(args=None if sys.argv[1:] else ["--help"])

    if args.command == "backup":
        odoo_backup(args.config)
    elif args.command == "restore" and args.file:
        odoo_restore(args.config, args.file, args.move)
    elif args.command == "trim":
        trim(args.config, args.d, int(args.n))
    elif args.command == "trimall":
        trim(args.config, args.d, int(args.n), all=True)
    elif args.command == "install":
        install_upgrade("install", args.module)
    elif args.command == "upgrade":
        install_upgrade("upgrade", args.module)
    elif args.command == "scaffold":
        if len(args.module) > 1:
            print("only one module allowed")
            return
        scaffold(args.module)
    elif args.command == "start":
        odoo_service("start")
    elif args.command == "stop":
        odoo_service("stop")
    elif args.command == "restart":
        odoo_service("stop")
        time.sleep(2)
        odoo_service("start")
    elif args.command == "hosts":
        odoo_hosts(args.domain)
    elif args.command == "caddy":
        odoo_caddy(args.domain)
    elif args.command == "logs":
        odoo_logs()
    return

### oda.py execute
usage: oda [-h] [-c CONFIG] [-d D] [-n N]
           {backup,restore,trim,trimall,install,upgrade,scaffold,start,stop,restart,hosts,caddy,logs}
           ...

Odoo Server Administration Tool

options:
  -h, --help            show this help message and exit
  -c CONFIG, --config CONFIG
                        odoo.conf file location
  -d D                  backup directory
  -n N                  number of backups to keep

commands:
  {backup,restore,trim,trimall,install,upgrade,scaffold,start,stop,restart,hosts,caddy,logs}
                        commands
    backup              Backup database filestore and addons
    restore             Restore database and filestore or addons [backupfiles]
    trim                Trim database backups
    trimall             Trim all database backups

    install             Install module(s) [modules]
    upgrade             Upgrade module(s) [modules]

    scaffold            Scaffold module [modulename]

    start               start odoo server
    stop                top odoo server
    restart             restart odoo server

    hosts               update hosts file
    caddy               update caddy file

    logs                tail the logs

thanks for using oda!

## odaserver
