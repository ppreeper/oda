# subuid
echo "root:1000:1" | sudo tee -a /etc/subuid /etc/subgid
echo "root:1001:1" | sudo tee -a /etc/subuid /etc/subgid
# subgid
echo "root:1000:1" | sudo tee -a /etc/subuid /etc/subgid
echo "root:1001:1" | sudo tee -a /etc/subuid /etc/subgid

