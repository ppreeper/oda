#!/usr/bin/env python3
"""Odoo Administration DB tool for Servers"""
import argparse
import json
import os
import subprocess
import sys
import time
from contextlib import closing
from shutil import rmtree, which

import psycopg2


# ==============================================================================
# Backup
def odoo_backup(configfile):
    bkp_prefix = f"{time.strftime('%Y_%m_%d_%H_%M_%S')}"
    # main database and filestore
    print("odoo:", _dump_db_tar(configfile, bkp_prefix, "/opt/odoo/backups"))
    # addons
    print("addons:", _dump_addons_tar(configfile, bkp_prefix, "/opt/odoo/backups"))


def _dump_addons_tar(configfile, bkp_prefix, bkp_dest="/opt/odoo/backups"):
    """Backup Odoo DB addons folders"""
    db_name = get_odoo_conf(configfile, "db_name")
    addons = get_odoo_conf(configfile, "addons_path").split(",")[2:]
    for addon in addons:
        folder = addon.replace("/opt/odoo/", "")
        dirlist = os.listdir(addon)
        if len(dirlist) != 0:
            tar_cmd = "tar"
            bkp_file = f"{bkp_prefix}__{db_name}__{folder}.tar.zst"
            file_path = os.path.join(bkp_dest, bkp_file)
            tar_args = ["ahcf", file_path, "-C", addon, "."]
            r = subprocess.run(
                [tar_cmd, *tar_args],
                stdout=subprocess.DEVNULL,
                stderr=subprocess.STDOUT,
                check=True,
            )
            if r.returncode != 0:
                raise OSError(f"could not backup addons {addon}")
            return file_path


def _dump_db_tar(configfile, bkp_prefix, bkp_dest="/opt/odoo/backups"):
    """Backup Odoo DB to dump file"""
    db_host = get_odoo_conf(configfile, "db_host")
    db_port = get_odoo_conf(configfile, "db_port")
    db_name = get_odoo_conf(configfile, "db_name")
    db_user = get_odoo_conf(configfile, "db_user")
    db_password = get_odoo_conf(configfile, "db_password")
    data_dir = get_odoo_conf(configfile, "data_dir")
    bkp_file = f"{bkp_prefix}__{db_name}.tar.zst"
    dump_dir = os.path.abspath(os.path.join(bkp_dest, f"{bkp_prefix}__{db_name}"))
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
    r = subprocess.run(
        pg_cmd,
        env=PG_ENV,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.STDOUT,
        check=True,
    )
    if r.returncode != 0:
        raise OSError(f"could not backup postgresql database {db_name}")

    # filestore
    filestore = os.path.join(data_dir, "filestore", db_name)
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


# ==============================================================================
# Restore
def odoo_restore(configfile, backup_files, copy=False):
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
            _restore_db_tar(configfile, dump_file, copy=copy)
        elif len(bfile) == 3 and bfile[-1] == "addons":
            print(f"restore addons file {dump_file}")
            _restore_addons_tar(dump_file)
        else:
            print("invalid backup filename")


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


def _restore_db_tar(configfile, dump_file, bkp_dir="/opt/odoo/backups", copy=True):
    """Restore Odoo DB from dump file"""
    db_host = get_odoo_conf(configfile, "db_host")
    db_port = get_odoo_conf(configfile, "db_port")
    db_name = get_odoo_conf(configfile, "db_name")
    db_user = get_odoo_conf(configfile, "db_user")
    db_password = get_odoo_conf(configfile, "db_password")

    bkp_file = os.path.basename(dump_file)

    db_name = get_odoo_conf(configfile, "db_name")
    data_dir = get_odoo_conf(configfile, "data_dir")

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
    tar = subprocess.run(
        tar_cmd, stdout=subprocess.DEVNULL, stderr=subprocess.STDOUT, check=True
    )
    if tar.returncode != 0:
        raise OSError(f"could not restore filestore for {db_name}")

    #########
    # odoo database neutralize if not copy
    _neutralize_db(configfile, copy)
    return


def _neutralize_db(configfile, copy=False):
    if not copy:
        db = db_connect(configfile)
        with closing(db.cursor()) as cr:
            try:
                # -- remove the enterprise code, report.url and web.base.url
                cr.execute(
                    "delete from ir_config_parameter where key in ('database.enterprise_code', 'report.url', 'web.base.url.freeze')"
                )
            except:
                pass

            try:
                # -- deactivate crons
                cr.execute("""UPDATE ir_cron SET active = 'f';""")
                cr.execute(
                    """UPDATE ir_cron SET active = 't' WHERE id IN (SELECT res_id FROM ir_model_data WHERE name = 'autovacuum_job' AND module = 'base');"""
                )
            except:
                pass

            try:
                # -- remove platform ir_logging
                cr.execute("""DELETE FROM ir_logging WHERE func = 'odoo.sh';""")
            except:
                pass

            try:
                # -- reset db uuid
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
                    do UPDATE set value = (current_date+'3 months'::interval)::timestamp;"""
                )
            except:
                pass

            try:
                # -- disable prod environment in all delivery carriers
                cr.execute("""UPDATE delivery_carrier SET prod_environment = false;""")
            except:
                pass

            try:
                # -- disable delivery carriers from external providers
                cr.execute(
                    """UPDATE delivery_carrier SET active = false WHERE delivery_type NOT IN ('fixed', 'base_on_rule');"""
                )
            except:
                pass

            try:
                cr.execute(
                    """UPDATE iap_account SET account_token = REGEXP_REPLACE(account_token, '(\+.*)?$', '+disabled');"""
                )
            except:
                pass

            try:
                # -- deactivate mail template
                cr.execute("""UPDATE mail_template SET mail_server_id = NULL;""")
            except:
                pass

            try:
                # -- deactivate fetchmail server
                cr.execute("""UPDATE fetchmail_server SET active = false;""")
            except:
                pass

            try:
                # -- disable generic payment provider
                cr.execute(
                    """UPDATE payment_provider SET state = 'disabled' WHERE state NOT IN ('test', 'disabled');"""
                )
            except:
                pass

            try:
                # -- activate neutralization watermarks
                cr.execute(
                    """UPDATE ir_ui_view SET active = true WHERE key = 'web.neutralize_banner';"""
                )
            except:
                pass

            try:
                # -- delete domains on websites
                cr.execute("""UPDATE website SET domain = NULL;""")
            except:
                pass

            try:
                # -- activate neutralization watermarks
                cr.execute(
                    """UPDATE ir_ui_view SET active = true WHERE key = 'website.neutralize_ribbon';"""
                )
            except:
                pass

            try:
                # -- disable cdn
                cr.execute("""UPDATE website SET cdn_activated = false;""")
            except:
                pass

            try:
                # -- disable bank synchronisation links
                cr.execute(
                    """UPDATE account_online_link SET provider_data = '', client_id = 'duplicate';"""
                )
            except:
                pass

            try:
                cr.execute(
                    """DELETE FROM ir_config_parameter WHERE key IN ('odoo_ocn.project_id', 'ocn.uuid');"""
                )
            except:
                pass

            try:
                # -- delete Facebook Access Tokens
                cr.execute(
                    """UPDATE social_account SET facebook_account_id = NULL, facebook_access_token = NULL;"""
                )
            except:
                pass

            try:
                # -- delete Instagram Access Tokens
                cr.execute(
                    """UPDATE social_account SET instagram_account_id = NULL, instagram_facebook_account_id = NULL, instagram_access_token = NULL;"""
                )
            except:
                pass

            try:
                # -- delete LinkedIn Access Tokens
                cr.execute(
                    """UPDATE social_account SET linkedin_account_urn = NULL, linkedin_access_token = NULL;"""
                )
            except:
                pass

            try:
                # -- Unset Firebase configuration within website
                cr.execute(
                    """UPDATE website SET firebase_enable_push_notifications = false, firebase_use_own_account = false, firebase_project_id = NULL, firebase_web_api_key = NULL, firebase_push_certificate_key = NULL, firebase_sender_id = NULL;"""
                )
            except:
                pass

            try:
                # -- delete Twitter Access Tokens
                cr.execute(
                    """UPDATE social_account SET twitter_user_id = NULL, twitter_oauth_token = NULL, twitter_oauth_token_secret = NULL;"""
                )
            except:
                pass

            try:
                # -- delete Youtube Access Tokens
                cr.execute(
                    """UPDATE social_account SET youtube_channel_id = NULL, youtube_access_token = NULL, youtube_refresh_token = NULL, youtube_token_expiration_date = NULL, youtube_upload_playlist_id = NULL;"""
                )
            except:
                pass

            try:
                # -- Remove Map Box Token as it's only valid per DB url
                cr.execute(
                    """DELETE FROM ir_config_parameter WHERE key = 'web_map.token_map_box';"""
                )
            except:
                pass

            try:
                cr.execute(
                    """UPDATE ir_cron SET active = 't' WHERE id IN (SELECT res_id FROM ir_model_data WHERE name = 'ir_cron_module_update_notification' AND module = 'mail');"""
                )
            except:
                pass

            try:
                # -- deactivate mail servers but activate default "localhost" mail server
                cr.execute(
                    """DO $$
                    BEGIN
                        UPDATE ir_mail_server SET active = 'f';
                        IF EXISTS (SELECT 1 FROM ir_module_module WHERE name='mail' and state IN ('installed', 'to upgrade', 'to remove')) THEN
                            UPDATE mail_template SET mail_server_id = NULL;
                        END IF;
                    EXCEPTION
                        WHEN undefined_table OR undefined_column THEN
                    END;
                $$;"""
                )
            except:
                pass
    return


# ==============================================================================
# Trim
def trim(configfile, bkp_path=".", limit=10, all=False):
    """Trim database backups"""
    db_name = get_odoo_conf(configfile, "db_name")

    # Get all backup files
    backups, addons = get_backup_files(bkp_path)

    rmbkp = []
    bkeys = list(backups.keys())
    bkeys.sort()
    if all:
        for k in bkeys:
            backups[k].sort()
            dates = backups[k][:-limit]
            for d in dates:
                rmbkp.append(f"{d}__{k}.tar.zst")
    else:
        backups[db_name].sort()
        dates = backups[db_name][:-limit]
        for d in dates:
            rmbkp.append(f"{d}__{db_name}.tar.zst")

    rmaddons = []
    akeys = list(addons.keys())
    akeys.sort()
    if all:
        for k in akeys:
            addons[k].sort()
            dates = addons[k][:-limit]
            for d in dates:
                rmaddons.append(f"{d}__{k}__addons.tar.zst")
    else:
        addons[db_name].sort()
        dates = addons[db_name][:-limit]
        for d in dates:
            rmaddons.append(f"{d}__{db_name}__addons.tar.zst")

    rmlist = rmbkp + rmaddons

    for r in rmlist:
        if os.path.exists(os.path.join(bkp_path, r)):
            print(" ".join(["rm", "-f", os.path.join(bkp_path, r)]))
            os.remove(os.path.join(bkp_path, r))


def get_backup_files(bkp_path="."):
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
    return backups, addons


# ==============================================================================
# Database
class DatabaseExists(Warning):
    """Empty Database Class"""


def db_connect(configfile, postgres=False):
    db_host = get_odoo_conf(configfile, "db_host")
    db_port = "5432" if postgres else get_odoo_conf(configfile, "db_port")
    db_name = "postgres" if postgres else get_odoo_conf(configfile, "db_name")
    db_user = "postgres" if postgres else get_odoo_conf(configfile, "db_user")
    db_password = "postgres" if postgres else get_odoo_conf(configfile, "db_password")
    return psycopg2.connect(
        dbname=db_name, user=db_user, password=db_password, host=db_host, port=db_port
    )


def _create_empty_database(configfile, db_name):
    """Create empty postgres db"""
    db_template = get_odoo_conf(configfile, "db_template")
    db_user = get_odoo_conf(configfile, "db_user")
    dbc = db_connect(configfile, postgres=True)
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
            unaccent = get_odoo_conf(configfile, "unaccent")
            if unaccent != None:
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
    db = db_connect(configfile, postgres=True)
    with closing(db.cursor()) as cr:
        # database-altering operations cannot be executed inside a transaction
        cr.connection.autocommit = True
        _drop_conn(cr, db_name)
        try:
            cr.execute("""DROP DATABASE IF EXISTS %s""" % (db_name))
        except Exception as e:
            raise OSError(f"Couldn't drop database {db_name}: {e}") from e


# ==============================================================================
# Odoo Helpers
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


def odoo_hosts(domain):
    """Update hosts file"""
    if not domain:
        print("no domain provided")
        return
    hname = os.uname().nodename
    domain = domain[0]
    with open("/etc/hosts", "w") as f:
        f.write(f"127.0.1.1	{hname} {hname}.{domain}\n")
        f.write("127.0.0.1	localhost\n")
        f.write("::1		localhost ip6-localhost ip6-loopback\n")
        f.write("ff02::1		ip6-allnodes\n")
        f.write("ff02::2		ip6-allrouters\n")
    return


def odoo_caddy(domain):
    """Update Caddyfile"""
    if not domain:
        print("no domain provided")
        return
    hname = os.uname().nodename
    domain = domain[0]
    with open("/etc/caddy/Caddyfile", "w") as f:
        f.write(f"{hname}.{domain} " + "{\n")
        f.write("tls internal\n")
        f.write(f"reverse_proxy http://{hname}:8069\n")
        f.write(f"reverse_proxy /websocket http://{hname}:8072\n")
        f.write(f"reverse_proxy /longpolling/* http://{hname}:8072\n")
        f.write("encode gzip zstd\n")
        f.write("file_server\n")
        f.write("log\n")
        f.write("}\n")
    cmd = [
        "sudo",
        "caddy",
        "fmt",
        "--overwrite",
        "/etc/caddy/Caddyfile",
    ]
    subprocess.run(
        cmd,
        check=True,
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


# ==============================================================================
# CLI Helpers
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

    # backup         Backup database filestore and addons
    subparsers.add_parser("backup", help="Backup database filestore and addons")

    # restore        Restore database and filestore or addons
    restore_parser = subparsers.add_parser(
        "restore", help="Restore database and filestore or addons"
    )
    restore_parser.add_argument("--copy", help="copy database", action="store_true")
    restore_parser.add_argument("file", help="Path to backup file", nargs="+")

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
        odoo_restore(args.config, args.file, args.copy)
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


if __name__ == "__main__":
    main()
