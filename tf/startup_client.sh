sudo apt-get update -y
sudo apt-get install ca-certificates curl gnupg -y
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

# Add the repository to Apt sources:
echo \
    "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
    "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | \
sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update -y

sudo apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin -y

# Clone the repository
sudo git clone https://github.com/louisloechel/CloudServiceBenchmarking /home/ubuntu/CloudServiceBenchmarking

# Navigate to the repository directory
cd /home/ubuntu/CloudServiceBenchmarking

# Add user to docker group
sudo usermod -aG docker louisloechel

# Overwrite client/config.yml server_address with the server's internal IP
SERVER_IP="$(gcloud compute instances describe 'server-vm' --zone='europe-west10-a' --format='get(networkInterfaces[0].networkIP)')"

# Replace the server_address in client/config.yml with the server's IP
sudo sed -i "s/server_address:.*/server_address: $SERVER_IP/" client/config.yml

# Ping the server-vm to make sure it's running
while ! ping -c 1 $SERVER_IP; do sleep 1; done

# Start a new subshell with the new group
newgrp docker << EOF

# Start docker
sudo systemctl start docker

# Wait for docker to start
while ! docker info >/dev/null 2>&1; do sleep 10; done

# Build and run docker-compose
sudo docker compose  -f "docker-compose.yml" up -d --build grpc-client

EOF