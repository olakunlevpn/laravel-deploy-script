# Laravel Deploy Panel

A self-hosted web panel for deploying Laravel applications to a fresh Ubuntu server — no terminal, no config files, no notepad. Configure everything through a browser, hit Deploy, and watch it happen in real time.

Distributed as a **single Go binary** with the React frontend embedded. Copy one file to your server and you're running.

---

## What It Does

The panel replaces the manual process of SSHing into a server and editing shell variables. Instead, you open a browser, fill in a wizard, and the panel handles the full deployment:

1. Creates the project directory with correct ownership
2. Clones your GitHub repository
3. Creates the MySQL/PostgreSQL database, user, and grants privileges
4. Copies `.env.example` and sets production values
5. Runs `composer install`, `key:generate`, `migrate`, `storage:link`, and caches config/routes/views
6. Sets file permissions and ownership
7. Writes and enables an Nginx server block with security headers
8. Installs an SSL certificate via Certbot (optional — skipped if DNS is not yet pointed)
9. Sets up a Supervisor queue worker (optional)
10. Adds the scheduler cron job (optional)
11. Runs a health check to verify the site responds

After deployment, the panel provides a full management dashboard:

- **Dashboard** — System info (hostname, OS, uptime, CPU, memory, disk usage), software versions (PHP, Nginx, MySQL/PostgreSQL, Composer), service status with live green/red indicators
- **Actions** — Run Laravel artisan commands (cache, config, routes, views, migrate, optimize), restart Nginx, re-apply file permissions
- **Deploy Steps** — Re-run any individual deployment step from the UI
- **Logs** — View Laravel, Nginx access, and Nginx error logs with search and filter
- **Environment** — Edit the `.env` file directly with backup-on-save
- **Settings** — Generate a systemd service file for auto-start on boot

---

## Server Setup

The panel deploys your Laravel application — it does not install the server stack itself. Run the following on a fresh Ubuntu 20.04 / 22.04 server before using the panel.

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

## Installation

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

---

## Building From Source

The `panel` binary has no runtime dependencies — no Go, no Node.js. But to compile it yourself you need Go 1.21+ and Node.js 18+:

```bash
git clone https://github.com/olakunlevpn/laravel-deploy-script.git
cd laravel-deploy-script
./build.sh
```

The result is a single `panel` file — roughly 6 MB — that contains everything.

---

## Usage

### Wizard

Open `http://your-server-ip:4432` in your browser. The 5-step wizard walks you through:

**Step 1 — Domain & Project** Enter your domain name, GitHub repository URL, and branch.

**Step 2 — Database** Enter a database password. Select MySQL or PostgreSQL. The database name and user are auto-generated from your domain and can be edited.

| Domain | Derived DB name | Derived DB user |
|--------|----------------|-----------------|
| `myapp.com` | `myapp` | `myapp_user` |
| `my-app.com` | `my_app` | `my_app_user` |
| `api.myapp.com` | `api_myapp` | `api_myapp_user` |

**Step 3 — PHP & Server** Select your PHP version (8.1, 8.2, or 8.3). The site user is auto-detected from the server.

**Step 4 — Features** Toggle the queue worker (Supervisor) and task scheduler (cron) on or off.

**Step 5 — Review & Deploy** Review all settings. Check the DNS confirmation box if you have already pointed your domain to this server's IP — this enables SSL installation. Click **Deploy Now** and watch the progress in real time via SSE streaming.

Preflight checks run automatically before deployment starts to verify all required software is installed and services are running.

### Dashboard

After deployment, the dashboard displays comprehensive server information:

- **System** — Hostname, OS, kernel version, uptime, CPU cores, load average
- **Resources** — Memory and disk usage with color-coded progress bars
- **Software** — PHP, Nginx, MySQL/PostgreSQL, and Composer versions with installation status
- **Project** — Domain, site root, repository, branch, database details, enabled features
- **Services** — Live status of Nginx, PHP-FPM, MySQL, Supervisor, SSL certificate, and queue worker with restart buttons

### Webhook

The panel exposes a webhook endpoint for CI/CD integration:

```bash
curl -X POST http://your-server-ip:4432/api/webhook/deploy
```

This runs `git pull`, `composer install`, `migrate`, and re-caches config/routes/views — a quick redeploy without going through the full wizard.

---

## Configuration

Settings are saved to `config.json` in the same directory as the binary. The file is written when you click **Deploy Now** in the wizard. A backup is automatically created before each save.

To edit settings after a deployment, click **Setup** in the sidebar to re-open the wizard pre-filled with current values.

---

## Security

- The panel has **no authentication**. Access control relies entirely on your firewall. Restrict port 4432 to your own IP address.
- The binary **must run as root** because deployment operations require root-level access (Nginx config, MySQL root login, Certbot, Supervisor, chown).
- The MySQL setup assumes the root account is accessible without a password (the default on Ubuntu with `auth_socket`). If your root account has a password set, step 3 will fail.
- Config validation enforces domain format, password minimum length, and safe characters for database names and system users.

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
