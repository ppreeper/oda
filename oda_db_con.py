#!/usr/bin/env python3
"""Odoo Administration DB tool for Containers"""
import argparse
import time
import json
import os
import sys
import shutil
import subprocess
from contextlib import closing
from passlib.context import CryptContext
import psycopg2

sys.path.append("odoo")
import odoo

sys.path.append("odoo")


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


def _dump_db_tar(configfile, bkp_prefix, bkp_dest="/opt/odoo/backups"):
    """Backup Odoo DB to dump file"""
    db_name = get_odoo_conf(configfile, "db_name")
    bkp_file = f"{bkp_prefix}__{db_name}.tar.zst"
    dump_dir = os.path.abspath(os.path.join(bkp_dest, bkp_prefix))
    file_path = os.path.join(bkp_dest, bkp_file)

    # create dump_dir
    try:
        os.mkdir(dump_dir)
    except FileExistsError as e:
        raise OSError("directory already exists") from e

    # postgresql database
    pg_cmd = [
        shutil.which("pg_dump"),
        "--no-owner",
        "--file",
        os.path.join(dump_dir, "dump.sql"),
        db_name,
    ]
    r = subprocess.run(
        pg_cmd,
        env=odoo.tools.exec_pg_environ(),
        stdout=subprocess.DEVNULL,
        stderr=subprocess.STDOUT,
        check=True,
    )
    if r.returncode != 0:
        raise OSError(f"could not backup postgresql database {db_name}")

    # manifest.json
    with open(os.path.join(dump_dir, "manifest.json"), "w", encoding="UTF-8") as fh:
        db = odoo.sql_db.db_connect(db_name)
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
        tar_cmd, stdout=subprocess.DEVNULL, stderr=subprocess.STDOUT, check=True
    )
    if r.returncode != 0:
        raise OSError(f"could not backup database {db_name}")

    # cleanup dump_dir
    if os.path.exists(dump_dir):
        shutil.rmtree(dump_dir)
    return file_path


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


# Restore


def _restore_db_tar(
    configfile,
    bkp_file,
    bkp_dir="/opt/odoo/backups",
    copy=True,
):
    """Restore Odoo DB from dump file"""
    db_name = get_odoo_conf(configfile, "db_name")
    data_dir = get_odoo_conf(configfile, "data_dir")

    #########
    # Database Drop and Restore
    # drop postgresql database
    _drop_database(db_name)

    # create new postgresql database
    _create_empty_database(db_name)

    # restore postgresql database
    # print("restore postgresql database")
    tarpg_cmd = ["tar", "Oaxf", os.path.join(bkp_dir, bkp_file), "./dump.sql"]
    pg_cmd = [shutil.which("psql"), "--dbname", db_name, "-q"]
    tarpg = subprocess.Popen(
        tarpg_cmd,
        stdout=subprocess.PIPE,
        stderr=subprocess.DEVNULL,
    )
    pg = subprocess.Popen(
        pg_cmd,
        env=odoo.tools.exec_pg_environ(),
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
    _drop_filestore(db_name)
    # restore filestore from backup
    ddirs = os.listdir(data_dir)
    if len(ddirs) != 0:
        for ddir in ddirs:
            shutil.rmtree(os.path.join(data_dir, ddir))
    # restore filestore: get filestore directory
    filestore = "/opt/odoo/data"
    # restore filestore: make dir
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
    tar = subprocess.run(
        tar_cmd, stdout=subprocess.DEVNULL, stderr=subprocess.STDOUT, check=True
    )
    if tar.returncode != 0:
        raise OSError(f"could not restore filestore for {db_name}")

    #########
    # odoo database registry
    # print("odoo database registry")
    if copy:
        db = odoo.sql_db.db_connect(db_name)
        with db.cursor() as cr:
            cr.execute(
                "delete from ir_config_parameter where key='database.enterprise_code'"
            )
            cr.execute(
                """update ir_config_parameter set value=(select gen_random_uuid())
                where key = 'database.uuid'"""
            )
            cr.execute(
                """insert into ir_config_parameter
                (key,value,create_uid,create_date,write_uid,write_date)
                values
                ('database.expiration_date',(current_date+'3 months'::interval)::timestamp,1,
                current_timestamp,1,current_timestamp)
                on conflict (key)
                do update set value = (current_date+'3 months'::interval)::timestamp;"""
            )
    return


def _restore_addons_tar(bkp_file, bkp_dir="/opt/odoo/backups"):
    """Restore Odoo DB addons folders"""
    addons = ""
    dest = addons if addons != "" else bkp_file.split("_")[-1:][0].split(".")[0]
    dest = os.path.join("/opt/odoo", dest)
    for filename in os.listdir(dest):
        file_path = os.path.join(dest, filename)
        try:
            if os.path.isfile(file_path):
                os.remove(file_path)
            elif os.path.isdir(file_path):
                shutil.rmtree(file_path)
        except Exception as e:
            raise OSError(f"Error deleting {file_path}") from e
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


# Helpers
class DatabaseExists(Warning):
    """Empty Database Class"""


def _create_empty_database(name):
    """Create empty postgres db"""
    db = odoo.sql_db.db_connect("postgres")
    with closing(db.cursor()) as cr:
        chosen_template = odoo.tools.config["db_template"]
        cr.execute(
            "SELECT datname FROM pg_database WHERE datname = %s",
            (name,),
            log_exceptions=False,
        )
        if cr.fetchall():
            raise DatabaseExists("database %r already exists!" % (name,))
        else:
            # database-altering operations cannot be executed inside a transaction
            cr.rollback()
            cr.connection.autocommit = True

            # 'C' collate is only safe with template0, but provides more useful indexes
            collate = psycopg2.sql.SQL(
                "LC_COLLATE 'C'" if chosen_template == "template0" else ""
            )
            cr.execute(
                psycopg2.sql.SQL(
                    "CREATE DATABASE {} ENCODING 'unicode' {} TEMPLATE {}"
                ).format(
                    psycopg2.sql.Identifier(name),
                    collate,
                    psycopg2.sql.Identifier(chosen_template),
                )
            )

    try:
        db = odoo.sql_db.db_connect(name)
        with db.cursor() as cr:
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


def _drop_database(db_name):
    """Drop PostgreSQL DB"""
    if db_name not in list_dbs(True):
        return False
    odoo.modules.registry.Registry.delete(db_name)
    odoo.sql_db.close_db(db_name)

    db = odoo.sql_db.db_connect("postgres")
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
        else:
            print(f"DROP DB: {db_name}")

    return True


def _drop_filestore(db_name):
    """Drop Filestore"""
    if db_name not in list_dbs(True):
        return False
    fs = odoo.tools.config.filestore(db_name)
    if os.path.exists(fs):
        shutil.rmtree(fs)
    return True


def list_dbs(force=False):
    """List PostgreSQL DB"""
    if not odoo.tools.config["list_db"] and not force:
        raise odoo.exceptions.AccessDenied()

    if not odoo.tools.config["dbfilter"] and odoo.tools.config["db_name"]:
        # In case --db-filter is not provided and --database is passed, Odoo will not
        # fetch the list of databases available on the postgres server and instead will
        # use the value of --database as comma seperated list of exposed databases.
        res = sorted(db.strip() for db in odoo.tools.config["db_name"].split(","))
        return res

    chosen_template = odoo.tools.config["db_template"]
    templates_list = tuple(set(["postgres", chosen_template]))
    db = odoo.sql_db.db_connect("postgres")
    with closing(db.cursor()) as cr:
        try:
            cr.execute(
                """select datname from pg_database where datdba=(select usesysid from pg_user
                where usename=current_user) and not datistemplate and datallowconn
                and datname not in %s order by datname""",
                (templates_list,),
            )
            res = [odoo.tools.ustr(name) for (name,) in cr.fetchall()]
        except Exception as e:
            raise OSError("Listing databases failed") from e
    return res


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


# Admin password
def change_password(new_password):
    new_password = new_password.strip()
    if new_password == "":
        return
    ctx = CryptContext(schemes=["pbkdf2_sha512"])
    pw_hash = ctx.hash(new_password)
    print(pw_hash)
    return


# =============================================================================


def main():
    """Odoo Administration Backup Restore"""
    parser = argparse.ArgumentParser()
    parser.add_argument("-b", "--backup", action="store_true", help="backup database")
    # parser.add_argument(
    #     "-r", "--restore", action="store", help="restore database", nargs="+"
    # )
    parser.add_argument(
        "-c",
        "--config",
        action="store",
        default="/opt/odoo/conf/odoo.conf",
        help="odoo.conf file location",
    )
    parser.add_argument(
        "-p", "--password", action="store", help="generate password hash"
    )

    args = parser.parse_args()

    if args.password:
        change_password(args.password)
        return

    if args.backup and args.restore:
        print("backup or restore cannot run both commands")
        return

    if args.backup:
        bkp_prefix = f"{time.strftime('%Y_%m_%d_%H_%M_%S')}"
        # main database and filestore
        print(_dump_db_tar(args.config, bkp_prefix, "/opt/odoo/backups"))
        # addons
        print(_dump_addons_tar(args.config, bkp_prefix, "/opt/odoo/backups"))
        return

    # for dump in args.restore:
    #     dump_file = dump.strip('"')
    #     fname = os.path.splitext(os.path.basename(dump_file))[0].split(".")[0]
    #     bfile = os.path.splitext(fname)[0].split("__")

    #     if len(bfile) == 2:
    #         print(f"restore from dump file {dump_file}")
    #         _restore_db_tar(args.config, dump_file)
    #     elif len(bfile) == 3 and bfile[-1] == "addons":
    #         print(f"restore addons file {dump_file}")
    #         _restore_addons_tar(dump_file)
    #     else:
    #         print("invalid backup filename")
    return


if __name__ == "__main__":
    main()
