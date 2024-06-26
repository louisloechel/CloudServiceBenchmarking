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

# Start a new subshell with the new group
newgrp docker << EOF

# Start docker
sudo systemctl start docker

# Wait for docker to start
while ! docker info >/dev/null 2>&1; do sleep 1; done

# Build and run docker-compose
docker compose  -f "docker-compose.yml" up -d --build grpc-server

EOF