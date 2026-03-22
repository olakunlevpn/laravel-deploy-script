#!/bin/bash
set -e

# ============================================================
# LARAVEL DEPLOYMENT SCRIPT
# Change ONLY these variables, then run: sudo bash deploy.sh
# ============================================================

DOMAIN="mychoicemyworld.in"
GITHUB_REPO="https://github.com/olakunlevpn/mychoiceworld.git"
GITHUB_BRANCH="main"
PHP_VERSION="8.3"
DB_PASSWORD="Green@1230"
ENABLE_QUEUE_WORKER=true
ENABLE_SCHEDULER=true

# ============================================================
# AUTO-GENERATED VARIABLES (no need to touch these)
# ============================================================
SITE_USER="root"
SITE_GROUP="www-data"
SITE_ROOT="/home/${SITE_USER}/${DOMAIN}"
DB_NAME=$(echo "${DOMAIN}" | sed 's/[^a-zA-Z0-9]/_/g' | sed 's/_com$//' | sed 's/_+/_/g')
DB_USER="root"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

print_step() {
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}  STEP $1: $2${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

print_info() {
    echo -e "${YELLOW}  → $1${NC}"
}

print_success() {
    echo -e "${GREEN}  ✓ $1${NC}"
}

print_error() {
    echo -e "${RED}  ✗ $1${NC}"
}

# ============================================================
# PRE-FLIGHT CHECKS
# ============================================================
echo ""
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}  DEPLOYING: ${DOMAIN}${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "  Domain:     ${YELLOW}${DOMAIN}${NC}"
echo -e "  Repo:       ${YELLOW}${GITHUB_REPO}${NC}"
echo -e "  Branch:     ${YELLOW}${GITHUB_BRANCH}${NC}"
echo -e "  PHP:        ${YELLOW}${PHP_VERSION}${NC}"
echo -e "  DB Name:    ${YELLOW}${DB_NAME}${NC}"
echo -e "  DB User:    ${YELLOW}${DB_USER}${NC}"
echo -e "  Site Root:  ${YELLOW}${SITE_ROOT}${NC}"
echo -e "  Queue:      ${YELLOW}${ENABLE_QUEUE_WORKER}${NC}"
echo -e "  Scheduler:  ${YELLOW}${ENABLE_SCHEDULER}${NC}"
echo ""

if [ "$EUID" -ne 0 ]; then
    print_error "Please run as root: sudo bash deploy.sh"
    exit 1
fi

if [ "$GITHUB_REPO" = "https://github.com/YOUR_USERNAME/YOUR_REPO.git" ]; then
    print_error "You forgot to change GITHUB_REPO. Edit deploy.sh first."
    exit 1
fi

if [ "$DB_PASSWORD" = "CHANGE_THIS_STRONG_PASSWORD" ]; then
    print_error "You forgot to change DB_PASSWORD. Edit deploy.sh first."
    exit 1
fi

read -p "  Proceed with deployment? (y/n): " CONFIRM
if [ "$CONFIRM" != "y" ]; then
    echo "  Aborted."
    exit 0
fi

# ============================================================
# STEP 1: Create folder & set ownership
# ============================================================
print_step "1/10" "Creating project directory"

if [ -d "$SITE_ROOT" ]; then
    print_info "Directory already exists: ${SITE_ROOT}"
    read -p "  Delete and recreate? (y/n): " DELETE_CONFIRM
    if [ "$DELETE_CONFIRM" = "y" ]; then
        rm -rf "$SITE_ROOT"
        print_info "Removed existing directory"
    else
        print_error "Aborted. Remove the directory manually or change DOMAIN."
        exit 1
    fi
fi

mkdir -p "$SITE_ROOT"
chown -R ${SITE_USER}:${SITE_GROUP} "$SITE_ROOT"
chmod -R 775 "$SITE_ROOT"
print_success "Created ${SITE_ROOT}"

# ============================================================
# STEP 2: Clone repository
# ============================================================
print_step "2/10" "Cloning repository"

sudo -u ${SITE_USER} git clone -b ${GITHUB_BRANCH} "${GITHUB_REPO}" "${SITE_ROOT}"
print_success "Cloned ${GITHUB_REPO} (${GITHUB_BRANCH})"

# ============================================================
# STEP 3: Create MySQL database & user
# ============================================================
print_step "3/10" "Setting up MySQL database"

mysql -e "CREATE DATABASE IF NOT EXISTS \`${DB_NAME}\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
mysql -e "CREATE USER IF NOT EXISTS '${DB_USER}'@'localhost' IDENTIFIED BY '${DB_PASSWORD}';"
mysql -e "GRANT ALL PRIVILEGES ON \`${DB_NAME}\`.* TO '${DB_USER}'@'localhost';"
mysql -e "FLUSH PRIVILEGES;"
print_success "Database: ${DB_NAME} | User: ${DB_USER}"

# ============================================================
# STEP 4: Set up .env file
# ============================================================
print_step "4/10" "Configuring .env"

cd "${SITE_ROOT}"

if [ -f ".env.example" ]; then
    sudo -u ${SITE_USER} cp .env.example .env
else
    print_info "No .env.example found, creating .env from scratch"
    sudo -u ${SITE_USER} touch .env
fi

# Update .env values — replace if line exists, append if missing
set_env_value() {
    local key="$1"
    local value="$2"
    local file="$3"
    if grep -q "^${key}=" "$file"; then
        sed -i "s|^${key}=.*|${key}=${value}|" "$file"
    else
        echo "${key}=${value}" >> "$file"
    fi
}

set_env_value "APP_URL" "https://${DOMAIN}" .env
set_env_value "APP_ENV" "production" .env
set_env_value "APP_DEBUG" "false" .env
set_env_value "DB_CONNECTION" "mysql" .env
set_env_value "DB_HOST" "127.0.0.1" .env
set_env_value "DB_PORT" "3306" .env
set_env_value "DB_DATABASE" "${DB_NAME}" .env
set_env_value "DB_USERNAME" "${DB_USER}" .env
set_env_value "DB_PASSWORD" "${DB_PASSWORD}" .env
set_env_value "QUEUE_CONNECTION" "database" .env
set_env_value "SESSION_DRIVER" "file" .env

print_success ".env configured for production"

# ============================================================
# STEP 5: Install Composer dependencies
# ============================================================
print_step "5/10" "Installing Composer dependencies"

cd "${SITE_ROOT}"
sudo -u ${SITE_USER} /usr/bin/php${PHP_VERSION} /usr/bin/composer install --no-dev --optimize-autoloader --no-interaction
print_success "Composer dependencies installed"

# Generate app key
sudo -u ${SITE_USER} /usr/bin/php${PHP_VERSION} artisan key:generate --force
print_success "App key generated"

# Run migrations
sudo -u ${SITE_USER} /usr/bin/php${PHP_VERSION} artisan migrate --force
print_success "Migrations complete"

# Storage link
sudo -u ${SITE_USER} /usr/bin/php${PHP_VERSION} artisan storage:link 2>/dev/null || true
print_success "Storage linked"

# Cache config, routes, views for production
sudo -u ${SITE_USER} /usr/bin/php${PHP_VERSION} artisan config:cache
sudo -u ${SITE_USER} /usr/bin/php${PHP_VERSION} artisan route:cache
sudo -u ${SITE_USER} /usr/bin/php${PHP_VERSION} artisan view:cache
print_success "Config, routes, views cached"

# ============================================================
# STEP 6: Set permissions
# ============================================================
print_step "6/10" "Setting file permissions"

chown -R ${SITE_USER}:${SITE_GROUP} "${SITE_ROOT}"
chmod -R 755 "${SITE_ROOT}"
chmod -R ug+rwx "${SITE_ROOT}/storage" "${SITE_ROOT}/bootstrap/cache"
print_success "Permissions set"

# ============================================================
# STEP 7: Configure Nginx
# ============================================================
print_step "7/10" "Configuring Nginx"

NGINX_CONF="/etc/nginx/sites-available/${DOMAIN}"

cat > "${NGINX_CONF}" <<NGINX
server {
    listen 80;
    listen [::]:80;
    server_name ${DOMAIN} www.${DOMAIN};
    root ${SITE_ROOT}/public;

    index index.php index.html;

    add_header X-Frame-Options "SAMEORIGIN";
    add_header X-XSS-Protection "1; mode=block";
    add_header X-Content-Type-Options "nosniff";

    charset utf-8;

    location / {
        try_files \$uri \$uri/ /index.php?\$query_string;
    }

    location = /favicon.ico { access_log off; log_not_found off; }
    location = /robots.txt  { access_log off; log_not_found off; }

    access_log /var/log/nginx/${DOMAIN}-access.log;
    error_log /var/log/nginx/${DOMAIN}-error.log error;

    location ~ \.php$ {
        fastcgi_split_path_info ^(.+\.php)(/.+)$;
        fastcgi_pass unix:/var/run/php/php${PHP_VERSION}-fpm.sock;
        fastcgi_index index.php;
        include fastcgi_params;
        fastcgi_param SCRIPT_FILENAME \$document_root\$fastcgi_script_name;
    }

    location ~ /\.(?!well-known).* {
        deny all;
    }
}
NGINX

# Enable site
ln -sf "${NGINX_CONF}" "/etc/nginx/sites-enabled/${DOMAIN}"

# Test nginx config
nginx -t
systemctl reload nginx
print_success "Nginx configured and reloaded"

# ============================================================
# STEP 8: SSL with Certbot
# ============================================================
print_step "8/10" "Setting up SSL certificate"

print_info "Make sure DNS for ${DOMAIN} and www.${DOMAIN} points to this server first!"
read -p "  DNS is pointed and ready? (y/n): " DNS_CONFIRM

if [ "$DNS_CONFIRM" = "y" ]; then
    apt-get install -y certbot python3-certbot-nginx -qq
    certbot --nginx -d ${DOMAIN} -d www.${DOMAIN} --non-interactive --agree-tos --email admin@${DOMAIN} --redirect
    print_success "SSL certificate installed"
else
    print_info "Skipping SSL. Run this later:"
    print_info "sudo certbot --nginx -d ${DOMAIN} -d www.${DOMAIN}"
fi

# ============================================================
# STEP 9: Queue worker (Supervisor)
# ============================================================
print_step "9/10" "Setting up Supervisor queue worker"

if [ "$ENABLE_QUEUE_WORKER" = true ]; then
    SUPERVISOR_NAME=$(echo "${DOMAIN}" | sed 's/\./-/g')

    cat > "/etc/supervisor/conf.d/${SUPERVISOR_NAME}-worker.conf" <<SUPERVISOR
[program:${SUPERVISOR_NAME}-worker]
command=/usr/bin/php${PHP_VERSION} ${SITE_ROOT}/artisan queue:work --sleep=3 --tries=3 --timeout=120
autostart=true
autorestart=true
user=${SITE_USER}
numprocs=1
redirect_stderr=true
stdout_logfile=${SITE_ROOT}/storage/logs/worker.log
stopwaitsecs=3600
SUPERVISOR

    supervisorctl reread
    supervisorctl update
    supervisorctl start "${SUPERVISOR_NAME}-worker:*" 2>/dev/null || true
    print_success "Queue worker configured and started"
else
    print_info "Queue worker skipped (ENABLE_QUEUE_WORKER=false)"
fi

# ============================================================
# STEP 10: Cron scheduler
# ============================================================
print_step "10/10" "Setting up Laravel scheduler cron"

if [ "$ENABLE_SCHEDULER" = true ]; then
    CRON_JOB="* * * * * cd ${SITE_ROOT} && /usr/bin/php${PHP_VERSION} artisan schedule:run >> /dev/null 2>&1"
    EXISTING_CRON=$(crontab -u ${SITE_USER} -l 2>/dev/null || true)

    if echo "$EXISTING_CRON" | grep -q "${DOMAIN}"; then
        print_info "Cron job already exists for ${DOMAIN}"
    else
        (echo "$EXISTING_CRON"; echo "$CRON_JOB") | crontab -u ${SITE_USER} -
        print_success "Scheduler cron job added"
    fi
else
    print_info "Scheduler skipped (ENABLE_SCHEDULER=false)"
fi

# ============================================================
# DONE
# ============================================================
echo ""
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}  DEPLOYMENT COMPLETE: ${DOMAIN}${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "  ${YELLOW}Site:${NC}       https://${DOMAIN}"
echo -e "  ${YELLOW}Root:${NC}       ${SITE_ROOT}"
echo -e "  ${YELLOW}Database:${NC}   ${DB_NAME}"
echo -e "  ${YELLOW}DB User:${NC}    ${DB_USER}"
echo -e "  ${YELLOW}PHP:${NC}        ${PHP_VERSION}"
echo ""
echo -e "  ${YELLOW}Check logs:${NC}"
echo -e "    tail -n 50 ${SITE_ROOT}/storage/logs/laravel.log"
if [ "$ENABLE_QUEUE_WORKER" = true ]; then
echo -e "    tail -n 50 ${SITE_ROOT}/storage/logs/worker.log"
fi
echo ""
echo -e "  ${YELLOW}Useful commands:${NC}"
echo -e "    cd ${SITE_ROOT}"
echo -e "    sudo -u ${SITE_USER} php${PHP_VERSION} artisan migrate --force"
echo -e "    sudo -u ${SITE_USER} php${PHP_VERSION} artisan config:cache"
echo -e "    sudo nginx -t && sudo systemctl reload nginx"
echo ""
