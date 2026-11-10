#!/bin/bash
set -eu

# ==================================================================================== #
# VARIABLES
# ==================================================================================== #

# Set the timezone for the server. A full list of available timezones can be found by 
# running timedatectl list-timezones.
TIMEZONE=America/Los_Angeles

# Set the name of the new user to create.
USERNAME=ubuntu

# Prompt to enter a password for the PostgreSQL ubuntu user (rather than hard-coding
# a password in this script).
read -p "Enter password for ubuntu DB user: " DB_PASSWORD

# Force all output to be presented in en_US for the duration of this script. This avoids  
# any "setting locale failed" errors while this script is running, before we have 
# installed support for all locales. Do not change this setting!
export LC_ALL=en_US.UTF-8 

# ==================================================================================== #
# SCRIPT LOGIC
# ==================================================================================== #

# Enable the "universe" repository.
sudo add-apt-repository --yes universe

# Update all software packages.
sudo apt update

# Set the system timezone and install all locales.
sudo timedatectl set-timezone ${TIMEZONE}
sudo apt --yes install locales-all

# Force a password to be set for the new user the first time they log in.
sudo passwd --delete "${USERNAME}"
sudo chage --lastday 0 "${USERNAME}"

# Configure the firewall to allow SSH, HTTP and HTTPS traffic.
sudo ufw allow 22
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw --force enable

# Install fail2ban.
sudo apt --yes install fail2ban

# Install the migrate CLI tool.
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate.linux-amd64 /usr/local/bin/migrate

# Install PostgreSQL.
sudo apt --yes install postgresql postgresql-contrib

# Install prerequisites from vector
sudo apt install build-essential postgresql-server-dev-all -y

# Clone pgvector repo
cd /tmp
git clone https://github.com/pgvector/pgvector.git
cd pgvector

# Build and install pgvector
make
sudo make install

# Clean up
cd ..
rm -rf /tmp/pgvector

# Set up the music DB and create a user account with the password entered earlier.
sudo -i -u postgres psql -c "CREATE DATABASE music"
sudo -i -u postgres psql -d music -c "CREATE EXTENSION IF NOT EXISTS citext"
sudo -i -u postgres psql -d music -c "CREATE EXTENSION IF NOT EXISTS vector"
sudo -i -u postgres psql -d music -c "CREATE ROLE ubuntu WITH LOGIN PASSWORD '${DB_PASSWORD}'"
sudo -i -u postgres psql -d music -c "GRANT ALL PRIVILEGES ON DATABASE music TO ubuntu"

# Add a DSN for connecting to the greenlight database to the system-wide environment 
# variables in the /etc/environment file.
echo "MUSIC_DB_DSN='postgres://ubuntu:${DB_PASSWORD}@localhost/music'" | sudo tee -a /etc/environment

# Install Caddy (see https://caddyserver.com/docs/install#debian-ubuntu-raspbian).
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt --yes install caddy

# Upgrade all packages. Using the --force-confnew flag means that configuration 
# files will be replaced if newer ones are available.
sudo apt --yes -o Dpkg::Options::="--force-confnew" upgrade

echo "Script complete! Rebooting..."
sudo reboot
