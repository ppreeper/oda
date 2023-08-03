mkdir -p /share/projects/${args[projectname]}/${args[branch]}/{data,conf}
git clone ${args[projecturl]} -b ${args[branch]} /share/projects/${args[projectname]}/${args[branch]}/addons