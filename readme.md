# Deploy Go to ubuntu snippets

## Login to Ubuntu VPS

```
ssh root@123.123.123.123
```

## Update packages

```
apt-get update
apt-get upgrade
```

## Create a new user

```
adduser robby
adduser robby sudo
```

## Enable firewall

```
ufw app list
ufw allow OpenSSH
ufw enable
ufw status
```

## Add authorized ssh key

```
cd /home/robby
mkdir .ssh
nano .ssh/authorized_keys
```

Then past in your public ssh key.

## Login as new user

```
ssh robby@123.123.123.123
```

## Install go

```
cd /tmp
wget https://go.dev/dl/go1.19.linux-amd64.tar.gz
tar -xvf go1.11.linux-amd64.tar.gz
sudo mv go /usr/local
```

## Add go to path

```
nano ~/.profile
```

Then add the following line.

```
export PATH=$PATH:/usr/local/go/bin
```

# install postgres

```
sudo apt-get install postgresql postgresql-contrib
```

# Create a postgres user and database

```
sudo -i -u postgres
createdb notes
createuser --interactive
psql
```

```
ALTER USER robby WITH PASSWORD 'rk';
GRANT ALL PRIVILEGES ON DATABASE notes TO robby;
\q
```

## Copying local code to remote

```
rsync -a ./notes robby@123.123.123.123:/home/robby/go/src --exclude .env
```

## Build Go app on remote

```
cd /go/src/notes
go build
```

## Setup system.d service

```
sudo nano /lib/systemd/system/notes.service
```

Add the service code below.

```
[Unit]
Description=notes

[Service]
Environment=PORT=3000
Environment=GO_ENV=production
Environment=GIN_MODE=release
Environment=DB_URL=postgresql://robby:rk@123.123.123.123:5432/notes
Type=simple
Restart=always
RestartSec=5s
ExecStart=/home/robby/go/src/notes/notes

[Install]
WantedBy=multi-user.target
```

## Start the service and check status

```
sudo service notes start
sudo service notes statusy
```

# Setup coturn

allow these ports on your vps

To Action From

---

Nginx Full ALLOW Anywhere  
22 ALLOW Anywhere // for ssh
443 ALLOW Anywhere // ssl/tls
80 ALLOW Anywhere // default port for unencrypted web communication
3478,5349/tcp ALLOW Anywhere  
49160:49200/udp ALLOW Anywhere  
5349/tcp ALLOW Anywhere  
3478/udp ALLOW Anywhere

Nginx Full (v6) ALLOW Anywhere (v6)  
22 (v6) ALLOW Anywhere (v6) // for ssh
443 (v6) ALLOW Anywhere (v6) // ssl/tls
80 (v6) ALLOW Anywhere (v6) // efault port for unencrypted web communication
3478,5349/tcp (v6) ALLOW Anywhere (v6)  
49160:49200/udp (v6) ALLOW Anywhere (v6)  
5349/tcp (v6) ALLOW Anywhere (v6)  
3478/udp (v6) ALLOW Anywhere (v6)

## Installing nginx

```
sudo apt install nginx
```

## Enabling nginx on firewall

```
sudo ufw app list
sudo ufw allow "Nginx Full"
sudo ufw reload
sudo ufw status
```

## Restart nginx

```
sudo ln -s /etc/nginx/sites-available/notes.com /etc/nginx/sites-enabled/notes.com
sudo nginx -s reload
```

# Setup SSL with Certbot for Nginx

🔹 1. Install Certbot & Nginx Plugin
Ubuntu (20.04+):

```
sudo apt update
sudo apt install certbot python3-certbot-nginx -y
```

🔹 2. Obtain and Install SSL Certificate
Replace with your actual domain:

```
sudo certbot --nginx -d bincang-visual.cloud
```

✅ What this does:

- Verifies your domain
- Modifies your Nginx config to use SSL
- Sets up automatic HTTPS redirection

🔹 3. Test Automatic Renewal
Certbot sets up a cron job automatically, but test it:

```
sudo certbot renew --dry-run
```

🔹 4. Done! Visit:
https://bincang-visual.cloud

🔒 What Certbot Changed in Nginx
It usually updates your server block to look like this:

```
# Redirect HTTP to HTTPS
server {
   listen 80;
   server_name bincang-visual.cloud www.bincang-visual.cloud;
   return 301 https://$host$request_uri;
}

# HTTPS server block
server {
    #if ($host = www.bincang-visual.cloud) {
    #    return 301 https://$host$request_uri;
   # } # managed by Certbot

   listen 443 ssl http2;
   server_name bincang-visual.cloud www.bincang-visual.cloud;

   ssl_certificate /etc/letsencrypt/live/bincang-visual.cloud/fullchain.pem;
   ssl_certificate_key /etc/letsencrypt/live/bincang-visual.cloud/privkey.pem;
   include /etc/letsencrypt/options-ssl-nginx.conf;
   ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;

   location /ws {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "Upgrade";
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
   }
}
```

🛠 Troubleshooting Tips
Problem Fix
DNS not pointing Use https://dnschecker.org to verify A record
Port 80 blocked UFW: sudo ufw allow 80,443/tcp
