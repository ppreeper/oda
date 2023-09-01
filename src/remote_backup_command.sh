echo "# this file is located in 'src/remote_backup_command.sh'"
echo "# code for 'oda remote backup' goes here"
echo "# you can edit it freely and regenerate (it will not be overwritten)"
inspect_args

# what project is the node attached to
# cmd
# PROJECT=$(ssh ${args[node]} -- cat /etc/fstab | grep addons | awk '{print $1}' | awk -F':' '{print $2}' | sed 's,/addons$,,')
# returns
# /share/projects/quest15/issue_285

ssh ${args[node]} oda node backup