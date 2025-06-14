#!/bin/bash

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo "Please run as root (use sudo)"
    exit 1
fi

# Enable memory overcommit
echo "Enabling memory overcommit..."
echo "vm.overcommit_memory = 1" >> /etc/sysctl.conf
sysctl vm.overcommit_memory=1

# Create Redis config directory if it doesn't exist
mkdir -p /etc/redis

# Create a basic Redis configuration
cat > /etc/redis/redis.conf << EOF
# Basic Redis configuration
port 6379
bind 127.0.0.1
protected-mode yes
daemonize yes
pidfile /var/run/redis/redis-server.pid
logfile /var/log/redis/redis-server.log
dir /var/lib/redis

# Memory management
maxmemory 256mb
maxmemory-policy allkeys-lru

# Persistence
save 900 1
save 300 10
save 60 10000
stop-writes-on-bgsave-error yes
rdbcompression yes
rdbchecksum yes
dbfilename dump.rdb
EOF

# Create necessary directories
mkdir -p /var/lib/redis
mkdir -p /var/log/redis
mkdir -p /var/run/redis

# Set proper permissions
chown -R redis:redis /var/lib/redis
chown -R redis:redis /var/log/redis
chown -R redis:redis /var/run/redis

echo "Redis setup completed. Please restart Redis server to apply changes."
echo "You can start Redis with: redis-server /etc/redis/redis.conf" 