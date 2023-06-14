#!/usr/bin/env python3
import argparse
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

import psycopg2

import odoo
from odoo import SUPERUSER_ID
import odoo.release
import odoo.sql_db
import odoo.tools
from odoo.tools import find_pg_tool, exec_pg_environ

_logger = logging.getLogger(__name__)


class DatabaseExists(Warning):
    pass


def _create_empty_database(name):
    db = odoo.sql_db.db_connect('postgres')
    with closing(db.cursor()) as cr:
        chosen_template = odoo.tools.config['db_template']
        cr.execute("SELECT datname FROM pg_database WHERE datname = %s",
                   (name, ),
                   log_exceptions=False)
        if cr.fetchall():
            raise DatabaseExists("database %r already exists!" % (name, ))
        else:
            # database-altering operations cannot be executed inside a transaction
            cr.rollback()
            cr._cnx.autocommit = True

            # 'C' collate is only safe with template0, but provides more useful indexes
            collate = sql.SQL("LC_COLLATE 'C'" if chosen_template ==
                              'template0' else "")
            cr.execute(
                sql.SQL("CREATE DATABASE {} ENCODING 'unicode' {} TEMPLATE {}"
                        ).format(sql.Identifier(name), collate,
                                 sql.Identifier(chosen_template)))

    # TODO: add --extension=trigram,unaccent
    try:
        db = odoo.sql_db.db_connect(name)
        with db.cursor() as cr:
            cr.execute("CREATE EXTENSION IF NOT EXISTS pg_trgm")
            if odoo.tools.config['unaccent']:
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
        pid_col = 'pid' if cr._cnx.server_version >= 90200 else 'procpid'

        cr.execute(
            """SELECT pg_terminate_backend(%(pid_col)s)
                      FROM pg_stat_activity
                      WHERE datname = %%s AND
                            %(pid_col)s != pg_backend_pid()""" %
            {'pid_col': pid_col}, (db_name, ))
    except Exception:
        pass


def exp_drop(db_name):
    if db_name not in list_dbs(True):
        return False
    odoo.modules.registry.Registry.delete(db_name)
    odoo.sql_db.close_db(db_name)

    db = odoo.sql_db.db_connect('postgres')
    with closing(db.cursor()) as cr:
        # database-altering operations cannot be executed inside a transaction
        cr._cnx.autocommit = True
        _drop_conn(cr, db_name)

        try:
            cr.execute(
                sql.SQL('DROP DATABASE IF EXISTS {}').format(
                    sql.Identifier(db_name)))
        except Exception as e:
            _logger.info('DROP DB: %s failed:\n%s', db_name, e)
            raise Exception("Couldn't drop database %s: %s" % (db_name, e))
        else:
            _logger.info('DROP DB: %s', db_name)

    fs = odoo.tools.config.filestore(db_name)
    if os.path.exists(fs):
        shutil.rmtree(fs)
    return True


def dump_db_manifest(cr):
    pg_version = "%d.%d" % divmod(cr._obj.connection.server_version / 100, 100)
    cr.execute(
        "SELECT name, latest_version FROM ir_module_module WHERE state = 'installed'"
    )
    modules = dict(cr.fetchall())
    manifest = {
        'odoo_dump': '1',
        'db_name': cr.dbname,
        'version': odoo.release.version,
        'version_info': odoo.release.version_info,
        'major_version': odoo.release.major_version,
        'pg_version': pg_version,
        'modules': modules,
    }
    return manifest


def _dump_db(db_name, bkp_name, folder="./backups", backup_format='zip'):
    """Dump database `db` into file-like object `stream` if stream is None
    return a file object with the dump """
    bkp_file = f"{bkp_name}.zip"
    file_path = os.path.join(folder, bkp_file)
    with open(file_path, 'wb') as stream:
        _logger.info('DUMP DB: %s format %s', db_name, backup_format)

        cmd = [find_pg_tool('pg_dump'), '--no-owner', db_name]
        env = exec_pg_environ()

        if backup_format == 'zip':
            with tempfile.TemporaryDirectory() as dump_dir:
                filestore = odoo.tools.config.filestore(db_name)
                if os.path.exists(filestore):
                    shutil.copytree(filestore,
                                    os.path.join(dump_dir, 'filestore'))
                with open(os.path.join(dump_dir, 'manifest.json'), 'w') as fh:
                    db = odoo.sql_db.db_connect(db_name)
                    with db.cursor() as cr:
                        json.dump(dump_db_manifest(cr), fh, indent=4)
                cmd.insert(-1, '--file=' + os.path.join(dump_dir, 'dump.sql'))
                subprocess.run(cmd,
                               env=env,
                               stdout=subprocess.DEVNULL,
                               stderr=subprocess.STDOUT,
                               check=True)
                if stream:
                    odoo.tools.osutil.zip_dir(
                        dump_dir,
                        stream,
                        include_dir=False,
                        fnct_sort=lambda file_name: file_name != 'dump.sql')
                else:
                    t = tempfile.TemporaryFile()
                    odoo.tools.osutil.zip_dir(
                        dump_dir,
                        t,
                        include_dir=False,
                        fnct_sort=lambda file_name: file_name != 'dump.sql')
                    t.seek(0)
                    return t
        else:
            cmd.insert(-1, '--format=c')
            stdout = subprocess.Popen(cmd,
                                      env=env,
                                      stdin=subprocess.DEVNULL,
                                      stdout=subprocess.PIPE).stdout
            if stream:
                shutil.copyfileobj(stdout, stream)
            else:
                return stdout
        return file_path


def _dump_addons(addons,bkp_name,bkp_dir="./backups", ):
    cwd = os.getcwd()
    for addon in addons:
        folder=addon.replace(cwd+"/","")
        bkp_file = f"{bkp_name}_{folder}.zip"
        file_path = os.path.join(bkp_dir, bkp_file)
        with open(file_path, 'wb') as stream:
            odoo.tools.osutil.zip_dir(
                folder,
                file_path,
                include_dir=False
            )
            print(file_path)


def _restore_db(db, dump_file, copy=False, neutralize_database=False):
    _logger.info('RESTORING DB: %s', db)

    exp_drop(db)
    _create_empty_database(db)

    filestore_path = None
    with tempfile.TemporaryDirectory() as dump_dir:
        if zipfile.is_zipfile(dump_file):
            # v8 format
            with zipfile.ZipFile(dump_file, 'r') as z:
                # only extract known members!
                filestore = [
                    m for m in z.namelist() if m.startswith('filestore/')
                ]
                z.extractall(dump_dir, ['dump.sql'] + filestore)

                if filestore:
                    filestore_path = os.path.join(dump_dir, 'filestore')

            pg_cmd = 'psql'
            pg_args = ['-q', '-f', os.path.join(dump_dir, 'dump.sql')]
        else:
            # <= 7.0 format (raw pg_dump output)
            pg_cmd = 'pg_restore'
            pg_args = ['--no-owner', dump_file]

        r = subprocess.run(
            [find_pg_tool(pg_cmd), '--dbname=' + db, *pg_args],
            env=exec_pg_environ(),
            stdout=subprocess.DEVNULL,
            stderr=subprocess.STDOUT,
        )
        if r.returncode != 0:
            raise Exception("Couldn't restore database")

        registry = odoo.modules.registry.Registry.new(db)
        with registry.cursor() as cr:
            env = odoo.api.Environment(cr, SUPERUSER_ID, {})
            if copy:
                # if it's a copy of a database, force generation of a new dbuuid
                env['ir.config_parameter'].init(force=True)
            if neutralize_database:
                odoo.modules.neutralize.neutralize_database(cr)

            if filestore_path:
                filestore_dest = env['ir.attachment']._filestore()
                shutil.move(filestore_path, filestore_dest)

    _logger.info('RESTORE DB: %s', db)


def _restore_addons(dump_file, addons=""):
    cwd = os.getcwd()
    dest = addons if addons != "" else dump_file.split('_')[-1:][0].split('.')[0]
    with zipfile.ZipFile(dump_file, 'r') as z:
    #     # only extract known members!
        z.extractall(dest)



def list_dbs(force=False):
    if not odoo.tools.config['list_db'] and not force:
        raise odoo.exceptions.AccessDenied()

    if not odoo.tools.config['dbfilter'] and odoo.tools.config['db_name']:
        # In case --db-filter is not provided and --database is passed, Odoo will not
        # fetch the list of databases available on the postgres server and instead will
        # use the value of --database as comma seperated list of exposed databases.
        res = sorted(db.strip()
                     for db in odoo.tools.config['db_name'].split(','))
        return res

    chosen_template = odoo.tools.config['db_template']
    templates_list = tuple(set(['postgres', chosen_template]))
    db = odoo.sql_db.db_connect('postgres')
    with closing(db.cursor()) as cr:
        try:
            cr.execute(
                "select datname from pg_database where datdba=(select usesysid from pg_user where usename=current_user) and not datistemplate and datallowconn and datname not in %s order by datname",
                (templates_list, ))
            res = [odoo.tools.ustr(name) for (name, ) in cr.fetchall()]
        except Exception:
            _logger.exception('Listing databases failed:')
            res = []
    return res


def main():
    argParser = argparse.ArgumentParser()
    argParser.add_argument("-b",
                           "--backup",
                           action="store_true",
                           help="backup database")
    argParser.add_argument("-r",
                           "--restore",
                           action="store_true",
                           help="restore database")
    argParser.add_argument("-d",
                           "--dump_file",
                           action="store",
                           help="database dump file")
    argParser.add_argument("-a",
                           "--addons",
                           action="store_true",
                           help="restore addons")
    argParser.add_argument("-f",
                           "--folder",
                           action="store",
                           help="addons folder")

    args = argParser.parse_args()
    odoo.tools.config._parse_config(["-c", "./conf/odoo.conf"])
    db_name = odoo.tools.config["db_name"]
    addons = odoo.tools.config["addons_path"].split(',')[2:]

    if not args.backup and not args.restore and not args.addons:
        print(argParser.print_help())
        return

    if args.backup and (args.restore or args.addons):
        print("backup or restore cannot run both commands")
        return

    if args.backup:
        dbName = f"{time.strftime('%Y_%m_%d_%H_%M_%S')}_{db_name}"
        print(_dump_db(db_name,bkp_name))
        _dump_addons(addons,bkp_name)
        return

    if args.restore and args.dump_file is None or args.dump_file == "":
        print("restore command requires a dump file to read")
        return

    if args.restore and args.dump_file:
        print(f"restore from dump file {args.dump_file}")
        _restore_db(db_name, args.dump_file)
        return

    if args.addons and args.dump_file is None or args.dump_file == "":
        print("addons restore command requires a dump file to read")
        return

    if args.addons and args.dump_file:
        print(f"addons restore from dump file {args.dump_file} ")
        _restore_addons(args.dump_file) if (args.folder is None or args.folder == "") else _restore_addons(args.dump_file,args.folder)
        return


if __name__ == "__main__":
    main()
