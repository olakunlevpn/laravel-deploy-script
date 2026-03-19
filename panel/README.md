# Laravel Deploy Panel

A self-hosted web panel for deploying Laravel applications to a fresh Ubuntu server — no terminal, no config files, no notepad. Configure everything through a browser, hit Deploy, and watch it happen in real time.

Distributed as a **single Go binary** with the React frontend embedded. Copy one file to your server and you're running.

---

## What It Does

The panel replaces the manual process of SSHing into a server and editing shell variables. Instead, you open a browser, fill in a wizard, and the panel handles the full deployment:

1. Creates the project directory with correct ownership
2. Clones your GitHub repository
3. Creates the MySQL database, user, and grants privileges
4. Copies `.env.example` and sets production values
5. Runs `composer install`, `key:generate`, `migrate`, `storage:link`, and caches config/routes/views
6. Sets file permissions and ownership
7. Writes and enables an Nginx server block
8. Installs an SSL certificate via Certbot (optional — skipped if DNS is not yet pointed)
9. Sets up a Supervisor queue worker (optional)
10. Adds the scheduler cron job (optional)

After deployment, the dashboard lets you monitor services, run Laravel artisan commands, view logs, and re-run any individual step.

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

Once the stack is in place, proceed to Installation below.

---

## Requirements

The `panel` binary has no runtime dependencies of its own — no Go, no Node.js, nothing else to install.

**Building from source** (only if you want to compile it yourself):
- Go 1.21+
- Node.js 18+

---

## Building

Clone the repository and run the build script:

```bash
git clone https://github.com/olakunlevpn/laravel-deploy-script.git
cd laravel-deploy-script/panel
./build.sh
```

This builds the React frontend and compiles it into the Go binary. The result is a single `panel` file — roughly 6 MB — that contains everything.

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

## Usage

### Wizard

Open `http://your-server-ip:4432` in your browser. You will land on the deployment wizard.

**Step 1 — Domain & Project**

Enter your domain name, GitHub repository URL, and branch.

**Step 2 — Database**

Enter a database password. The database name and user are auto-generated from your domain and can be edited.

| Domain | Derived DB name | Derived DB user |
|--------|----------------|-----------------|
| `myapp.com` | `myapp` | `myapp_user` |
| `my-app.com` | `my_app` | `my_app_user` |
| `api.myapp.com` | `api_myapp` | `api_myapp_user` |

**Step 3 — PHP & Server**

Select your PHP version (8.1, 8.2, or 8.3). The site user is auto-detected from the server — this is the user that will own the files.

**Step 4 — Features**

Toggle the queue worker (Supervisor) and task scheduler (cron) on or off.

**Step 5 — Review & Deploy**

Review all settings. Check the DNS confirmation box if you have already pointed your domain to this server's IP — this enables SSL installation. Click **Deploy Now** and watch the progress in real time.

If any step fails, the deployment stops and shows the error output for that step.

### Dashboard

After a successful deployment, navigate to `/dashboard`.

**Service Status** shows the current state of Nginx, PHP-FPM, MySQL, Supervisor, SSL certificate expiry, and the queue worker. Each service has action buttons to start, stop, or restart it.

**Laravel Actions** lets you run artisan commands without a terminal:
- Clear cache, config, routes, and views
- Run migrations or roll back
- Optimize and re-link storage

**Nginx Actions** provides reload and restart buttons, and shows the current Nginx config for your site.

**Permissions** re-applies correct file ownership and permissions to the project directory.

**Laravel Log** shows the last 100 lines of `laravel.log` with a refresh button and a clear button.

**Deployment Steps** lists all 10 steps with their last run status and a Re-run button next to each. Use this to re-apply a specific step after making a change — for example, re-run step 7 after editing your Nginx config, or re-run step 9 after modifying your Supervisor settings.

---

## Configuration

Settings are saved to `config.json` in the same directory as the binary. The file is written when you click **Deploy Now** or **Save** in the wizard.

To edit settings after a deployment, click **Edit Config** in the dashboard header to re-open the wizard pre-filled with current values.

---

## Security

- The panel has **no authentication**. Access control relies entirely on your firewall. Restrict port 4432 to your own IP address.
- The binary **must run as root** because deployment operations require root-level access (Nginx config, MySQL root login, Certbot, Supervisor, chown).
- The MySQL setup assumes the root account is accessible without a password (the default on Ubuntu with `auth_socket`). If your root account has a password set, step 3 will fail.

---

## Development

To run the panel locally against a Go backend during development:

```bash
# Terminal 1 — start the Go backend
cd panel
go run .

# Terminal 2 — start the Vite dev server (proxies /api to :4432)
cd panel/frontend
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

The MIT License (MIT).
