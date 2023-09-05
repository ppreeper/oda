#!/usr/bin/env python3
import argparse
from tabnanny import check
import time
import json
import logging
import os
import sys
import shutil
import subprocess
import tempfile
import zipfile

sys.path.append("odoo")

from psycopg2 import sql
from contextlib import closing
from passlib.context import CryptContext

import psycopg2

import odoo
from odoo import SUPERUSER_ID
import odoo.release
import odoo.sql_db
import odoo.tools
from odoo.tools import find_pg_tool, exec_pg_environ

_logger = logging.getLogger(__name__)


# Backup


def dump_db_manifest(cr):
    pg_version = "%d.%d" % divmod(cr._obj.connection.server_version / 100, 100)
    cr.execute(
        "SELECT name, latest_version FROM ir_module_module WHERE state = 'installed'"
    )
    modules = dict(cr.fetchall())
    manifest = {
        "odoo_dump": "1",
        "db_name": cr.dbname,
        "version": odoo.release.version,
        "version_info": odoo.release.version_info,
        "major_version": odoo.release.major_version,
        "pg_version": pg_version,
        "modules": modules,
    }
    return manifest


def _dump_addons_tar(addons, bkp_name, bkp_dest="./bkpdir"):
    cwd = os.getcwd()
    for addon in addons:
        folder = addon.replace(cwd + "/", "")
        dir = os.listdir(folder)
        if len(dir) != 0:
            tar_cmd = "tar"
            bkp_file = f"{bkp_name}_{folder}.tar.zst"
            file_path = os.path.join(bkp_dest, bkp_file)
            tar_args = ["ahcf", file_path, "-C", folder, "."]
            # print([tar_cmd, *tar_args])
            r = subprocess.run(
                [tar_cmd, *tar_args],
                stdout=subprocess.DEVNULL,
                stderr=subprocess.STDOUT,
            )
            if r.returncode != 0:
                raise Exception(f"could not backup addons {addon}")
            return file_path


def _dump_db_tar(db_name, bkp_name, bkp_dest="./backups"):
    bkp_file = f"{bkp_name}.tar.zst"
    dump_dir = os.path.abspath(os.path.join(bkp_dest, bkp_name))
    file_path = os.path.join(bkp_dest, bkp_file)

    # create dump_dir
    try:
        os.mkdir(dump_dir)
    except FileExistsError as e:
        pass

    # postgresql database
    pg_cmd = [
        find_pg_tool("pg_dump"),
        "--no-owner",
        "--file",
        os.path.join(dump_dir, "dump.sql"),
        db_name,
    ]
    r = subprocess.run(
        pg_cmd,
        env=exec_pg_environ(),
        stdout=subprocess.DEVNULL,
        stderr=subprocess.STDOUT,
        check=True,
    )
    if r.returncode != 0:
        raise Exception(f"could not backup postgresql database {db_name}")

    # manifest.json
    with open(os.path.join(dump_dir, "manifest.json"), "w") as fh:
        db = odoo.sql_db.db_connect(db_name)
        with db.cursor() as cr:
            json.dump(dump_db_manifest(cr), fh, indent=4)

    # filestore
    filestore = odoo.tools.config.filestore(db_name)
    filestore_back = os.path.join(dump_dir, "filestore")
    if os.path.exists(filestore):
        try:
            os.symlink(filestore, filestore_back, target_is_directory=True)
        except FileExistsError as e:
            pass

    # create tar archive
    tar_cmd = ["tar", "achf", os.path.abspath(file_path), "-C", dump_dir, "."]
    r = subprocess.run(
        tar_cmd,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.STDOUT,
    )
    if r.returncode != 0:
        raise Exception(f"could not backup database {db_name}")

    # cleanup dump_dir
    if os.path.exists(dump_dir):
        shutil.rmtree(dump_dir)
    return file_path


# Restore


def _restore_addons_tar(bkp_file, addons=""):
    dest = addons if addons != "" else bkp_file.split("_")[-1:][0].split(".")[0]
    if os.path.isdir(dest):
        shutil.rmtree(dest)
    if not os.path.exists(dest):
        os.makedirs(dest)
    tar_cmd = "tar"
    tar_args = ["axf", bkp_file, "-C", dest, "."]
    r = subprocess.run(
        [tar_cmd, *tar_args],
        stdout=subprocess.DEVNULL,
        stderr=subprocess.STDOUT,
    )
    if r.returncode != 0:
        raise Exception(f"could not restore addons {dest}")


def _restore_db_tar(
    db_name,
    bkp_file,
    remote=False,
    copy=True,
    neutralize_database=False,
):
    print(db_name,bkp_file,remote,copy,neutralize_database)
    # drop postgresql database
    # exp_drop(db_name)

    # create new postgresql database
    # _create_empty_database(db_name)

    # restore postgresql database
    # tarpg_cmd = ["tar", "Oaxvf", bkp_file, "./dump.sql"]
    # pg_cmd = [find_pg_tool("psql"), "--dbname", db_name, "-q"]
    # tarpg = subprocess.Popen(
    #     tarpg_cmd,
    #     stdout=subprocess.PIPE,
    #     stderr=subprocess.DEVNULL,
    # )
    # pg = subprocess.Popen(
    #     pg_cmd,
    #     env=exec_pg_environ(),
    #     stdin=tarpg.stdout,
    #     stdout=subprocess.DEVNULL,
    #     stderr=subprocess.DEVNULL,
    # )
    # tarpg.stdout.close()
    # pg.communicate()
    # if pg.returncode != 0:
    #     raise Exception(f"could not backup postgresql database {db_name}")

    # if not remote:
        # restore filestore
        # restore filestore: cleanup dump_dir
        # data_dir = odoo.tools.config["data_dir"]
        # ddirs = os.listdir(data_dir)
        # if len(ddirs) != 0:
        #     for ddir in ddirs:
        #         shutil.rmtree(os.path.join(data_dir, ddir))
        # # restore filestore: get filestore directory
        # filestore = odoo.tools.config.filestore(db_name)
        # # restore filestore: make dir
        # try:
        #     os.makedirs(filestore, exist_ok=True)
        # except FileExistsError as e:
        #     pass
        # # restore filestore: extract from archive
        # tar_cmd = [
        #     "tar",
        #     "axf",
        #     bkp_file,
        #     "-C",
        #     filestore,
        #     "--strip-components=2",
        #     "./filestore",
        # ]
        # tar = subprocess.run(
        #     tar_cmd,
        #     stdout=subprocess.DEVNULL,
        #     stderr=subprocess.STDOUT,
        # )
        # if tar.returncode != 0:
        #     raise Exception(f"could not restore filestore for {db_name}")

    # odoo database registry
    # registry = odoo.modules.registry.Registry.new(db_name)
    # with registry.cursor() as cr:
    #     env = odoo.api.Environment(cr, SUPERUSER_ID, {})
    #     if copy:
    #         # change database.uuid if a copy (default)
    #         # if it's a copy of a database, force generation of a new dbuuid
    #         env["ir.config_parameter"].init(force=True)
    #     if neutralize_database:
    #         # neutralize (remove all modules) if needed
    #         odoo.modules.neutralize.neutralize_database(cr)
    return


# Helpers
class DatabaseExists(Warning):
    pass


def _create_empty_database(name):
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
            cr._cnx.autocommit = True

            # 'C' collate is only safe with template0, but provides more useful indexes
            collate = sql.SQL(
                "LC_COLLATE 'C'" if chosen_template == "template0" else ""
            )
            cr.execute(
                sql.SQL("CREATE DATABASE {} ENCODING 'unicode' {} TEMPLATE {}").format(
                    sql.Identifier(name), collate, sql.Identifier(chosen_template)
                )
            )

    # TODO: add --extension=trigram,unaccent
    try:
        db = odoo.sql_db.db_connect(name)
        with db.cursor() as cr:
            cr.execute("CREATE EXTENSION IF NOT EXISTS pg_trgm")
            if odoo.tools.config["unaccent"]:
                cr.execute("CREATE EXTENSION IF NOT EXISTS unaccent")
                # From PostgreSQL's point of view, making 'unaccent' immutable is incorrect
                # because it depends on external data - see
                # https://www.postgresql.org/message-id/flat/201012021544.oB2FiTn1041521@wwwmaster.postgresql.org#201012021544.oB2FiTn1041521@wwwmaster.postgresql.org
                # But in the case of Odoo, we consider that those data don't
                # change in the lifetime of a database. If they do change, all
                # indexes created with this function become corrupted!
                cr.execute("ALTER FUNCTION unaccent(text) IMMUTABLE")
    except psycopg2.Error as e:
        _logger.warning("Unable to create PostgreSQL extensions : %s", e)


def _drop_conn(cr, db_name):
    # Try to terminate all other connections that might prevent
    # dropping the database
    try:
        # PostgreSQL 9.2 renamed pg_stat_activity.procpid to pid:
        # http://www.postgresql.org/docs/9.2/static/release-9-2.html#AEN110389
        pid_col = "pid" if cr._cnx.server_version >= 90200 else "procpid"

        cr.execute(
            """SELECT pg_terminate_backend(%(pid_col)s)
                      FROM pg_stat_activity
                      WHERE datname = %%s AND
                            %(pid_col)s != pg_backend_pid()"""
            % {"pid_col": pid_col},
            (db_name,),
        )
    except Exception:
        pass


def exp_drop(db_name):
    if db_name not in list_dbs(True):
        return False
    odoo.modules.registry.Registry.delete(db_name)
    odoo.sql_db.close_db(db_name)

    db = odoo.sql_db.db_connect("postgres")
    with closing(db.cursor()) as cr:
        # database-altering operations cannot be executed inside a transaction
        cr._cnx.autocommit = True
        _drop_conn(cr, db_name)

        try:
            cr.execute(
                sql.SQL("DROP DATABASE IF EXISTS {}").format(sql.Identifier(db_name))
            )
        except Exception as e:
            _logger.info("DROP DB: %s failed:\n%s", db_name, e)
            raise Exception("Couldn't drop database %s: %s" % (db_name, e))
        else:
            _logger.info("DROP DB: %s", db_name)

    fs = odoo.tools.config.filestore(db_name)
    if os.path.exists(fs):
        shutil.rmtree(fs)
    return True


def list_dbs(force=False):
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
                "select datname from pg_database where datdba=(select usesysid from pg_user where usename=current_user) and not datistemplate and datallowconn and datname not in %s order by datname",
                (templates_list,),
            )
            res = [odoo.tools.ustr(name) for (name,) in cr.fetchall()]
        except Exception:
            _logger.exception("Listing databases failed:")
            res = []
    return res


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
    argParser = argparse.ArgumentParser()
    argParser.add_argument(
        "-b", "--backup", action="store_true", help="backup database"
    )
    argParser.add_argument(
        "-f",
        "--destfolder",
        action="store",
        default="./backups",
        help="backup destination folder",
    )
    argParser.add_argument(
        "-r", "--restore", action="store_true", help="restore database"
    )
    argParser.add_argument("-s", "--remote", action="store_true", help="remote restore")
    argParser.add_argument(
        "-d", "--dump_file", action="store", help="database dump file"
    )
    argParser.add_argument(
        "-p", "--password", action="store", help="generate password hash"
    )
    argParser.add_argument(
        "-c",
        "--config",
        action="store",
        default="./conf/odoo.conf",
        help="odoo.conf file location",
    )

    args = argParser.parse_args()
    odoo.tools.config._parse_config(["-c", args.config])
    db_name = odoo.tools.config["db_name"]
    addons = odoo.tools.config["addons_path"].split(",")[2:]

    if not args.backup and not args.restore and not args.password:
        print(argParser.print_help())
        return

    if args.password:
        change_password(args.password)
        return

    if args.backup and args.restore:
        print("backup or restore cannot run both commands")
        return

    if args.backup:
        bkp_name = f"{time.strftime('%Y_%m_%d_%H_%M_%S')}_{db_name}"
        # main database and filestore
        print(_dump_db_tar(db_name, bkp_name, args.destfolder))
        # addons
        print(_dump_addons_tar(addons, bkp_name, args.destfolder))
        return

    if args.restore and args.dump_file is None or args.dump_file == "":
        print("restore command requires a dump file to read")
        return

    if args.restore and args.dump_file:
        dump_file = args.dump_file.strip('"')
        fname = os.path.splitext(os.path.basename(dump_file))[0].split(".")[0]
        bfile = os.path.splitext(fname)[0].split("_")
        if len(bfile) >= 7:
            if bfile[-1]=="addons":
                print(f"restore addons file {dump_file}")
                _restore_addons_tar(dump_file)
            else:
                print(f"restore from dump file {dump_file}")
                _restore_db_tar(db_name, dump_file, args.remote)
        else:
            print("invalid backup filename")
        return


if __name__ == "__main__":
    main()
