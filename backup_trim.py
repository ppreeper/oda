#!/usr/bin/env python3
import os


def trim(mpath=".", limit=10):
    onlyfiles = [f for f in os.listdir(mpath) if os.path.isfile(os.path.join(mpath, f))]
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
        # print(k)
        backups[k].sort()
        destroy = backups[k][:-limit]
        for d in destroy:
            rmlist.append("__".join([d, k]) + ".tar.zst")
    akeys = list(addons.keys())
    akeys.sort()
    for k in akeys:
        # print(k)
        addons[k].sort()
        destroy = addons[k][:-limit]
        for d in destroy:
            rmlist.append("__".join([d, k, "addons"]) + ".tar.zst")
    rmlist.sort()
    for r in rmlist:
        print("rm -f ", r)
        if os.path.exists(r):
            os.remove(r)


if __name__ == "__main__":
    trim("/share/backups", 10)
