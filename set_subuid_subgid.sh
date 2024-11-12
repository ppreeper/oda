# subuid
echo "root:1000:1" | sudo tee -a /etc/subuid
echo "root:1001:1" | sudo tee -a /etc/subuid
# subgid
echo "root:1000:1" | sudo tee -a /etc/subgid
echo "root:1001:1" | sudo tee -a /etc/subgid

