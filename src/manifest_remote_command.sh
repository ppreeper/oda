REMOTE_FS=odoofs
ssh ${REMOTE_FS} tar Oaxf /share/${args[file]} ./manifest.json > manifest.json