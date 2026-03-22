# Laravel Deploy Script & Panel

One command. Your Laravel app is live.

This repo includes two ways to deploy Laravel applications to a fresh Ubuntu server:

1. **`deploy.sh`** — A bash script. Edit four variables, run it, done.
2. **Web Panel** — A browser-based UI. Fill in a wizard, click Deploy, watch it happen in real time.

The web panel is distributed as a **single Go binary** with the React frontend embedded. Copy one file to your server and you're running.

---

## What It Does

Both the script and the panel perform the same deployment steps:

1. Creates the project directory with correct ownership
2. Clones your GitHub repository
3. Creates the MySQL/PostgreSQL database, user, and grants privileges
4. Copies `.env.example` and sets production values
5. Runs `composer install`, `key:generate`, `migrate`, `storage:link`, and caches config/routes/views
6. Sets file permissions and ownership
7. Writes and enables an Nginx server block
8. Installs an SSL certificate via Certbot (optional — skipped if DNS is not yet pointed)
9. Sets up a Supervisor queue worker (optional)
10. Adds the scheduler cron job (optional)

The web panel additionally provides a dashboard to monitor services, run Laravel artisan commands, view logs, edit `.env`, and re-run any individual step.

---

## Server Setup

Run the following on a fresh Ubuntu 20.04 / 22.04 server before deploying.

**System packages**

```bash
sudo apt update && sudo apt upgrade -y
sudo apt install -y git curl unzip
```

**Nginx**

```bash
sudo apt install -y nginx
sudo systemctl enable nginx
```

**MySQL**

```bash
sudo apt install -y mysql-server
sudo systemctl enable mysql
```

**PHP (choose your version — 8.1, 8.2, or 8.3)**

```bash
sudo apt install -y software-properties-common
sudo add-apt-repository ppa:ondrej/php -y
sudo apt update

# Replace 8.3 with your preferred version
sudo apt install -y php8.3-fpm php8.3-cli php8.3-mysql php8.3-xml \
  php8.3-curl php8.3-mbstring php8.3-zip php8.3-bcmath php8.3-intl \
  php8.3-gd php8.3-redis php8.3-readline
```

**Composer**

```bash
curl -sS https://getcomposer.org/installer | php
sudo mv composer.phar /usr/local/bin/composer
```

**Certbot** (only needed for SSL)

```bash
sudo apt install -y certbot python3-certbot-nginx
```

**Supervisor** (only needed for queue workers)

```bash
sudo apt install -y supervisor
sudo systemctl enable supervisor
```

---

## Option A: Deploy Script

Open `deploy.sh` and change these four variables at the top:

```bash
DOMAIN="yourdomain.com"
GITHUB_REPO="https://github.com/you/your-laravel-app.git"
GITHUB_BRANCH="main"
DB_PASSWORD="a-strong-password-here"
```

Then run it:

```bash
sudo bash deploy.sh
```

That's it. The script will walk you through each step and confirm before doing anything destructive.

### Optional toggles

```bash
ENABLE_QUEUE_WORKER=false
ENABLE_SCHEDULER=false
```

### SSL note

The script will ask if your DNS is already pointing to the server before attempting SSL. If it's not ready yet, skip that step and run Certbot manually later:

```bash
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com
```

---

## Option B: Web Panel

### Installation

Download the latest `panel` binary from the [Releases](https://github.com/olakunlevpn/laravel-deploy-script/releases) page, then copy it to your server:

```bash
scp panel user@your-server:/opt/deploy-panel/panel
```

Run it as root (required — deployment steps need root access for Nginx, MySQL, Certbot, and Supervisor):

```bash
ssh user@your-server
sudo /opt/deploy-panel/panel
```

The panel is now available at `http://your-server-ip:4432`.

> **Firewall strongly recommended.** The panel has no authentication. Restrict port 4432 to trusted IP addresses only.

To use a different port:

```bash
sudo /opt/deploy-panel/panel --port 8080
```

### Building from source

The `panel` binary has no runtime dependencies — no Go, no Node.js. But to compile it yourself:

```bash
git clone https://github.com/olakunlevpn/laravel-deploy-script.git
cd laravel-deploy-script
./build.sh
```

This requires Go 1.21+ and Node.js 18+. The result is a single `panel` file — roughly 6 MB — that contains everything.

### Usage

**Wizard** — Open `http://your-server-ip:4432` in your browser. The 5-step wizard walks you through:

1. Domain & repository
2. Database credentials (auto-derived from domain)
3. PHP version & server user
4. Optional features (queue worker, scheduler)
5. Review & deploy with real-time SSE progress

**Dashboard** — After deployment, shows system info (hostname, OS, uptime, memory, disk, CPU), service status with green/red indicators, and software versions.

**Sidebar pages:**
- **Actions** — Laravel artisan commands, Nginx reload/restart, file permissions
- **Deploy Steps** — Re-run any individual deployment step
- **Logs** — View Laravel, Nginx access, and Nginx error logs with search
- **Environment** — Edit the `.env` file with backup-on-save
- **Settings** — Systemd service file for auto-start on boot

| Domain | Derived DB name | Derived DB user |
|--------|----------------|-----------------|
| `myapp.com` | `myapp` | `myapp_user` |
| `my-app.com` | `my_app` | `my_app_user` |
| `api.myapp.com` | `api_myapp` | `api_myapp_user` |

---

## Configuration

**Script** — Edit the variables at the top of `deploy.sh`.

**Panel** — Settings are saved to `config.json` in the same directory as the binary. The file is written when you click Deploy Now in the wizard.

---

## Security

- The web panel has **no authentication**. Access control relies entirely on your firewall. Restrict port 4432 to your own IP address.
- The binary **must run as root** because deployment operations require root-level access (Nginx config, MySQL root login, Certbot, Supervisor, chown).
- The MySQL setup assumes the root account is accessible without a password (the default on Ubuntu with `auth_socket`). If your root account has a password set, step 3 will fail.

---

## Development

To run the panel locally for development:

```bash
# Terminal 1 — start the Go backend
go run .

# Terminal 2 — start the Vite dev server (proxies /api to :4432)
cd frontend
npm run dev
```

The Vite dev server runs on `http://localhost:5173` and automatically proxies all `/api` requests to the Go backend on port 4432.

---

## Contributing

Pull requests are welcome. For significant changes, please open an issue first to discuss what you would like to change.

---

## Credits

- [Olakunle](https://github.com/olakunlevpn)
- [All Contributors](https://github.com/olakunlevpn/laravel-deploy-script/graphs/contributors)

---

## License

The MIT License (MIT). Please see [License File](LICENSE) for more information.
