#!/usr/bin/env python3
"""Odoo Administration DB tool for Servers"""
import argparse
from shutil import rmtree, which
import time
import json
import os
import sys

import subprocess


from contextlib import closing
from passlib.context import CryptContext

import psycopg2

sys.path.append("odoo")

import odoo


# Backup


def dump_db_manifest(db_name, cr):
    """Generate Odoo Manifest data"""
    pg_version = "%d.%d" % divmod(cr.connection.server_version / 100, 100)
    cr.execute(
        "SELECT name, latest_version FROM ir_module_module WHERE state = 'installed'"
    )
    modules = dict(cr.fetchall())
    manifest = {
        "odoo_dump": "1",
        "db_name": db_name,
        "version": odoo.release.version,
        "version_info": odoo.release.version_info,
        "major_version": odoo.release.major_version,
        "pg_version": pg_version,
        "modules": modules,
    }
    return manifest


def _dump_addons_tar(configfile, bkp_prefix, bkp_dest="/opt/odoo/backups"):
    """Backup Odoo DB addons folders"""
    cwd = os.getcwd()
    db_name = get_odoo_conf(configfile, "db_name")
    addons = get_odoo_conf(configfile, "addons_path").split(",")[2:]
    for addon in addons:
        folder = addon.replace(cwd + "/", "")
        dirlist = os.listdir(folder)
        if len(dirlist) != 0:
            tar_cmd = "tar"
            bkp_file = f"{bkp_prefix}__{db_name}__{folder}.tar.zst"
            file_path = os.path.join(bkp_dest, bkp_file)
            tar_args = ["ahcf", file_path, "-C", folder, "."]
            r = subprocess.run(
                [tar_cmd, *tar_args],
                stdout=subprocess.DEVNULL,
                stderr=subprocess.STDOUT,
                check=True,
            )
            if r.returncode != 0:
                raise OSError(f"could not backup addons {addon}")
            return file_path


def db_connect(configfile):
    db_host = get_odoo_conf(configfile, "db_host")
    db_port = get_odoo_conf(configfile, "db_port")
    db_name = get_odoo_conf(configfile, "db_name")
    db_user = get_odoo_conf(configfile, "db_user")
    db_password = get_odoo_conf(configfile, "db_password")
    return psycopg2.connect(
        dbname=db_name, user=db_user, password=db_password, host=db_host, port=db_port
    )


def db_connect_postgres(configfile):
    db_host = get_odoo_conf(configfile, "db_host")
    db_port = 5432
    db_name = "postgres"
    db_user = "postgres"
    db_password = "postgres"
    return psycopg2.connect(
        dbname=db_name, user=db_user, password=db_password, host=db_host, port=db_port
    )


def _dump_db_tar(configfile, bkp_prefix, bkp_dest="/opt/odoo/backups"):
    """Backup Odoo DB to dump file"""
    # print(configfile, bkp_prefix, bkp_dest)
    db_host = get_odoo_conf(configfile, "db_host")
    db_port = get_odoo_conf(configfile, "db_port")
    db_name = get_odoo_conf(configfile, "db_name")
    db_user = get_odoo_conf(configfile, "db_user")
    db_password = get_odoo_conf(configfile, "db_password")
    bkp_file = f"{bkp_prefix}__{db_name}.tar.zst"
    dump_dir = os.path.abspath(os.path.join(bkp_dest, bkp_prefix))
    file_path = os.path.join(bkp_dest, bkp_file)

    # create dump_dir
    try:
        os.mkdir(dump_dir)
    except FileExistsError as e:
        raise OSError("directory already exists") from e

    # postgresql database
    PG_ENV = {"PGPASSWORD": db_password}
    pg_cmd = [
        which("pg_dump"),
        "-h",
        db_host,
        "-p",
        db_port,
        "-U",
        db_user,
        "--no-owner",
        "--file",
        os.path.join(dump_dir, "dump.sql"),
        db_name,
    ]
    # print("pg_cmd", pg_cmd)
    r = subprocess.run(
        pg_cmd,
        env=PG_ENV,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.STDOUT,
        check=True,
    )
    if r.returncode != 0:
        raise OSError(f"could not backup postgresql database {db_name}")

    # manifest.json
    with open(os.path.join(dump_dir, "manifest.json"), "w", encoding="UTF-8") as fh:
        db = db_connect(configfile)
        with db.cursor() as cr:
            json.dump(dump_db_manifest(db_name, cr), fh, indent=4)

    # filestore
    filestore = odoo.tools.config.filestore(db_name)
    filestore_back = os.path.join(dump_dir, "filestore")
    if os.path.exists(filestore):
        try:
            os.symlink(filestore, filestore_back, target_is_directory=True)
        except FileExistsError as e:
            raise OSError("symlink failed") from e

    # create tar archive
    tar_cmd = ["tar", "achf", os.path.abspath(file_path), "-C", dump_dir, "."]
    r = subprocess.run(
        tar_cmd,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.STDOUT,
        check=True,
    )
    if r.returncode != 0:
        raise OSError(f"could not backup database {db_name}")

    # cleanup dump_dir
    if os.path.exists(dump_dir):
        rmtree(dump_dir)
    return file_path


# Restore


def _restore_addons_tar(dump_file, bkp_dir="/opt/odoo/backups"):
    """Restore Odoo DB addons folders"""
    addons = ""
    bkp_file = os.path.basename(dump_file)
    dest = addons if addons != "" else dump_file.split("_")[-1:][0].split(".")[0]
    dest = os.path.join("/opt/odoo", dest)
    rmrf(dest)
    tar_cmd = "tar"
    tar_args = ["axf", os.path.join(bkp_dir, bkp_file), "-C", dest, "."]
    try:
        subprocess.run(
            [tar_cmd, *tar_args],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.STDOUT,
            check=True,
        )
    except Exception as e:
        raise OSError(f"could not restore addons {dest}") from e


def _restore_db_tar(
    configfile,
    dump_file,
    bkp_dir="/opt/odoo/backups",
    copy=True,
):
    """Restore Odoo DB from dump file"""
    db_host = get_odoo_conf(configfile, "db_host")
    db_port = get_odoo_conf(configfile, "db_port")
    db_name = get_odoo_conf(configfile, "db_name")
    db_user = get_odoo_conf(configfile, "db_user")
    db_password = get_odoo_conf(configfile, "db_password")

    bkp_file = os.path.basename(dump_file)

    db_name = get_odoo_conf(configfile, "db_name")
    data_dir = get_odoo_conf(configfile, "data_dir")

    #########
    # Database Drop and Restore
    # drop postgresql database
    _drop_database(configfile, db_name)

    # create new postgresql database
    _create_empty_database(configfile, db_name)

    # restore postgresql database
    tarpg_cmd = ["tar", "Oaxf", os.path.join(bkp_dir, bkp_file), "./dump.sql"]
    pg_cmd = [
        which("psql"),
        "-h",
        db_host,
        "-p",
        db_port,
        "-U",
        db_user,
        "--dbname",
        db_name,
        "-q",
    ]
    tarpg = subprocess.Popen(
        tarpg_cmd,
        stdout=subprocess.PIPE,
        stderr=subprocess.DEVNULL,
    )
    pg = subprocess.Popen(
        pg_cmd,
        env={"PGPASSWORD": db_password},
        stdin=tarpg.stdout,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
    )
    tarpg.stdout.close()
    pg.communicate()
    if pg.returncode != 0:
        raise OSError(f"could not backup postgresql database {db_name}")

    #########
    # Filestore Restore
    # drop current filestore
    data_dir = get_odoo_conf(configfile, "data_dir")
    rmrf(data_dir)
    # restore filestore from backup
    # restore filestore: make dir
    filestore = os.path.join("/opt/odoo/data", "filestore", db_name)
    try:
        os.makedirs(filestore, exist_ok=True)
    except FileExistsError as e:
        raise OSError("filestore already exists") from e
    # restore filestore: extract from archive
    tar_cmd = [
        "tar",
        "axf",
        os.path.join(bkp_dir, bkp_file),
        "-C",
        filestore,
        "--strip-components=2",
        "./filestore",
    ]
    # print(tar_cmd)
    tar = subprocess.run(
        tar_cmd, stdout=subprocess.DEVNULL, stderr=subprocess.STDOUT, check=True
    )
    if tar.returncode != 0:
        raise OSError(f"could not restore filestore for {db_name}")

    #########
    # odoo database registry
    if copy:
        db = db_connect(configfile)
        with closing(db.cursor()) as cr:
            try:
                cr.execute(
                    "delete from ir_config_parameter where key='database.enterprise_code'"
                )
            except:
                pass
            try:
                cr.execute(
                    """update ir_config_parameter set value=(select gen_random_uuid())
                    where key = 'database.uuid'"""
                )
            except:
                pass
            try:
                cr.execute(
                    """insert into ir_config_parameter
                    (key,value,create_uid,create_date,write_uid,write_date)
                    values
                    ('database.expiration_date',(current_date+'3 months'::interval)::timestamp,1,
                    current_timestamp,1,current_timestamp)
                    on conflict (key)
                    do update set value = (current_date+'3 months'::interval)::timestamp;"""
                )
            except:
                pass
    return


# Helpers
class DatabaseExists(Warning):
    """Empty Database Class"""


def _create_empty_database(configfile, db_name):
    """Create empty postgres db"""
    db_template = get_odoo_conf(configfile, "db_template")
    db_user = get_odoo_conf(configfile, "db_user")
    dbc = db_connect_postgres(configfile)
    with closing(dbc.cursor()) as cr:
        chosen_template = db_template
        cr.execute("SELECT datname FROM pg_database WHERE datname = %s", (db_name,))
        if cr.fetchall():
            raise DatabaseExists("database %r already exists!" % (db_name,))
        else:
            cr.connection.rollback()
            cr.connection.autocommit = True
            # 'C' collate is only safe with template0, but provides more useful indexes
            cr.execute(
                "CREATE DATABASE %s ENCODING 'unicode' %s TEMPLATE %s OWNER %s"
                % (
                    db_name,
                    "LC_COLLATE 'C'",
                    chosen_template,
                    db_user,
                ),
            )
    dbu = db_connect(configfile)
    with closing(dbu.cursor()) as cr:
        try:
            cr.execute("CREATE EXTENSION IF NOT EXISTS pg_trgm")
            if odoo.tools.config["unaccent"]:
                cr.execute("CREATE EXTENSION IF NOT EXISTS unaccent")
                cr.execute("ALTER FUNCTION unaccent(text) IMMUTABLE")
        except psycopg2.Error as e:
            raise OSError(f"Unable to create PostgreSQL extensions : {e}") from e


def _drop_conn(cr, db_name):
    """Try to terminate all other connections that might prevent dropping the database"""
    try:
        pid_col = "pid" if cr.connection.server_version >= 90200 else "procpid"

        cr.execute(
            """SELECT pg_terminate_backend(%(pid_col)s)
                      FROM pg_stat_activity
                      WHERE datname = %%s AND
                            %(pid_col)s != pg_backend_pid()"""
            % {"pid_col": pid_col},
            (db_name,),
        )
    except Exception as e:
        raise OSError("database could not be dropped") from e


def _drop_database(configfile, db_name):
    """Drop PostgreSQL DB"""
    db = db_connect_postgres(configfile)
    with closing(db.cursor()) as cr:
        # database-altering operations cannot be executed inside a transaction
        cr.connection.autocommit = True
        _drop_conn(cr, db_name)
        try:
            cr.execute(
                psycopg2.sql.SQL("DROP DATABASE IF EXISTS {}").format(
                    psycopg2.sql.Identifier(db_name)
                )
            )
        except Exception as e:
            raise OSError(f"Couldn't drop database {db_name}: {e}") from e


def install_upgrade(mode, module):
    module_list = []
    for m in module:
        mods = m.split(",")
        for mod in mods:
            module_list.append(mod)
    modules = list(dict.fromkeys(module_list))
    """Install/Upgrade module"""
    odoo_cmd = [
        "odoo/odoo-bin",
        "-c",
        "/opt/odoo/conf/odoo.conf",
        "--no-http",
        "--stop-after-init",
    ]
    if mode == "install":
        odoo_cmd.append("-i")
        odoo_cmd.append(",".join(modules))
    elif mode == "upgrade":
        odoo_cmd.append("-u")
        odoo_cmd.append(",".join(modules))
    r = subprocess.run(
        odoo_cmd,
        check=True,
    )
    return


def scaffold(module):
    """Scaffold module"""
    odoo_cmd = [
        "odoo/odoo-bin",
        "scaffold",
        module[0],
        "/opt/odoo/addons/",
    ]
    r = subprocess.run(
        odoo_cmd,
        check=True,
    )
    print(f"module '{module[0]}' scaffolded")
    return


def trim(bkp_path=".", limit=10):
    """Trim database backups"""
    onlyfiles = [
        f for f in os.listdir(bkp_path) if os.path.isfile(os.path.join(bkp_path, f))
    ]
    backups = {}
    addons = {}
    for f in onlyfiles:
        fparts = f.split("__")
        # backups
        if fparts[-1].endswith("zst") and len(fparts) == 2:
            fs = f.replace(".tar.zst", "").split("__")
            if len(fs) == 2:
                if fs[1] not in backups:
                    backups[fs[1]] = [fs[0]]
                elif fs[1] in backups and len(backups[fs[1]]) == 0:
                    backups[fs[1]] = [fs[0]]
                else:
                    backups[fs[1]].append(fs[0])
        # addons
        if fparts[-1].endswith("zst") and len(fparts) == 3:
            fs = f.replace(".tar.zst", "").split("__")
            if len(fs) == 3:
                if fs[1] not in addons:
                    addons[fs[1]] = [fs[0]]
                elif fs[1] in addons and len(addons[fs[1]]) == 0:
                    addons[fs[1]] = [fs[0]]
                else:
                    addons[fs[1]].append(fs[0])
    rmlist = []
    bkeys = list(backups.keys())
    bkeys.sort()
    for k in bkeys:
        backups[k].sort()
        destroy = backups[k][:-limit]
        for d in destroy:
            rmlist.append("__".join([d, k]) + ".tar.zst")
    akeys = list(addons.keys())
    akeys.sort()
    for k in akeys:
        addons[k].sort()
        destroy = addons[k][:-limit]
        for d in destroy:
            rmlist.append("__".join([d, k, "addons"]) + ".tar.zst")
    rmlist.sort()
    for r in rmlist:
        print("rm -f ", os.path.join(bkp_path, r))
        if os.path.exists(os.path.join(bkp_path, r)):
            os.remove(os.path.join(bkp_path, r))


def rmrf(data_dir):
    """Remove Directory Contents"""
    for filename in os.listdir(data_dir):
        file_path = os.path.join(data_dir, filename)
        try:
            if os.path.isfile(file_path):
                os.remove(file_path)
            elif os.path.isdir(file_path):
                rmtree(file_path)
        except Exception as e:
            raise OSError(f"Error deleting {file_path}") from e


def get_odoo_conf(configfile, key):
    """get key value from odoo.conf"""
    with open(
        os.path.join(configfile),
        "r",
        encoding="UTF-8",
    ) as f:
        lines = f.readlines()
        for line in lines:
            if line.startswith(key):
                return line.split("=")[1].strip()
    return


def odoo_service(action):
    """Start/Stop Odoo Service"""
    print(f"Odoo Service {action}")
    subprocess.run(
        [
            "sudo",
            "systemctl",
            action,
            "odoo.service",
        ],
        check=False,
    )


def odoo_logs():
    """Tail Odoo Logs"""
    print("Odoo Logs")
    subprocess.run(
        [
            "sudo",
            "journalctl",
            "-u",
            "odoo.service",
            "-f",
        ],
        check=False,
    )


def odoo_backup(configfile):
    bkp_prefix = f"{time.strftime('%Y_%m_%d_%H_%M_%S')}"
    # main database and filestore
    print("odoo:", _dump_db_tar(configfile, bkp_prefix, "/opt/odoo/backups"))
    # addons
    print("addons:", _dump_addons_tar(configfile, bkp_prefix, "/opt/odoo/backups"))


def odoo_restore(configfile, backup_files):
    """Restore from backup file"""
    if not are_you_sure("restore from backup"):
        return
    # stop
    files = parse_multi(backup_files)
    files.sort()
    for dump_file in files:
        fname = os.path.splitext(os.path.basename(dump_file))[0].split(".")[0]
        bfile = os.path.splitext(fname)[0].split("__")
        if len(bfile) == 2:
            print(f"restore from dump file {dump_file}")
            _restore_db_tar(configfile, dump_file)
        elif len(bfile) == 3 and bfile[-1] == "addons":
            print(f"restore addons file {dump_file}")
            _restore_addons_tar(dump_file)
        else:
            print("invalid backup filename")


# Admin password
def change_password(new_password):
    new_password = new_password.strip()
    if new_password == "":
        return
    ctx = CryptContext(schemes=["pbkdf2_sha512"])
    pw_hash = ctx.hash(new_password)
    print(pw_hash)
    return


def are_you_sure(action):
    """Double Prompt"""
    text = input(f"Are you sure you want to {action} [YES/N] ")
    if text != "YES":
        return False
    text = input(f"Are you really sure you want to {action} [YES/N] ")
    if text != "YES":
        return False
    return True


def parse_multi(multi):
    """parse multiples list"""
    multi_list = []
    for m in multi:
        multi_list.extend(m.split(","))
    return multi_list


# =============================================================================
class ArgParser(argparse.ArgumentParser):
    """ArgParser modified to output help on error"""

    def error(self, message):
        print(f"error: {message}\n")
        self.print_help()


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
    # ===================
    # backup         Backup database filestore and addons
    subparsers.add_parser("backup", help="Backup database filestore and addons")
    # config
    parser.add_argument("-b", "--backup", action="store_true", help="backup database")
    # restore        Restore database and filestore or addons
    restore_parser = subparsers.add_parser(
        "restore", help="Restore database and filestore or addons"
    )
    restore_parser.add_argument("file", help="Path to backup file", nargs="+")
    # trim           Trim database backups
    subparsers.add_parser("trim", help="Trim database backups")
    # install        Install module(s)
    install_parser = subparsers.add_parser("install", help="Install module(s)")
    install_parser.add_argument("module", help="module(s) to install", nargs="+")
    # upgrade        Upgrade module(s)
    upgrade_parser = subparsers.add_parser("upgrade", help="Upgrade module(s)")
    upgrade_parser.add_argument("module", help="module(s) to upgrade", nargs="+")
    # scaffold        Scaffold module
    scaffold_parser = subparsers.add_parser("scaffold", help="Scaffold module")
    scaffold_parser.add_argument("module", help="module to scaffold", nargs="*")
    # start          start odoo server
    subparsers.add_parser("start", help="start odoo server")
    # stop           stop odoo server
    subparsers.add_parser("stop", help="top odoo server")
    # restart        restart odoo server
    subparsers.add_parser("restart", help="restart odoo server")
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
        "-p",
        "--password",
        action="store",
        default="",
        help="change admin password",
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

    if args.command == "backup" or args.backup:
        odoo_backup(args.config)
    elif args.command == "restore" and args.file:
        odoo_restore(args.config, args.file)
    elif args.command == "trim":
        trim(args.d, int(args.n))
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
    elif args.command == "logs":
        odoo_logs()
    elif args.command == "password":
        change_password(args.password)
    return


if __name__ == "__main__":
    main()
