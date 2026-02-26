# Laravel Deploy Script

One command. Your Laravel app is live.

This is a bash script that takes care of the entire server setup for a Laravel application. You give it a domain name, a GitHub repo, and a database password — it handles everything else.

No more copy-pasting commands from your notes. No more forgetting a step and debugging for an hour.

## What it does

When you run `deploy.sh`, it will:

- Create the project directory under `/home/forge/`
- Clone your repository from GitHub
- Create a MySQL database and user
- Configure your `.env` file for production
- Install Composer dependencies
- Generate an app key, run migrations, and link storage
- Cache config, routes, and views
- Set the correct file permissions
- Set up an Nginx virtual host
- Install an SSL certificate with Certbot
- Configure a Supervisor queue worker
- Add the Laravel scheduler to cron

All of this from a single command.

## Requirements

- Ubuntu/Debian server
- Nginx
- PHP 8.3 (with FPM)
- MySQL
- Composer
- Certbot
- Supervisor
- A `forge` user on the server

## Installation

Download or clone this repo to your server:

```bash
git clone https://github.com/olakunlevpn/laravel-deploy-script.git
cd laravel-deploy-script
```

## Usage

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

You can turn off the queue worker or scheduler if your app doesn't need them:

```bash
ENABLE_QUEUE_WORKER=false
ENABLE_SCHEDULER=false
```

### SSL note

The script will ask if your DNS is already pointing to the server before attempting SSL. If it's not ready yet, you can skip that step and run Certbot manually later:

```bash
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com
```

## How it names things

Everything is derived from the domain name automatically:

| Domain | Database | DB User | Nginx Log |
|---|---|---|---|
| `myapp.com` | `myapp` | `myapp_user` | `myapp.com-access.log` |
| `client-site.com` | `client_site` | `client_site_user` | `client-site.com-access.log` |

You don't have to configure any of that.

## Contributing

Contributions are welcome! Here's how you can help:

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/my-improvement`)
3. Make your changes
4. Test the script on a fresh server if possible
5. Commit your changes (`git commit -m 'Add my improvement'`)
6. Push to the branch (`git push origin feature/my-improvement`)
7. Open a Pull Request

### Some ideas for contributions

- Support for PostgreSQL
- Support for Redis session/cache driver setup
- Automatic Node.js/NPM asset building
- Rollback functionality
- Multi-site deployment from a config file
- Support for other PHP versions

### Guidelines

- Keep it simple. The whole point is that this script is easy to read and modify.
- Stick to bash. No dependencies on external tools beyond what a typical Laravel server already has.
- Test your changes. If you can, run the script on a fresh Ubuntu server before submitting.
- One feature per PR. Makes it easier to review and merge.

## Credits

- [Olakunle](https://github.com/olakunlevpn)
- [All Contributors](https://github.com/olakunlevpn/laravel-deploy-script/graphs/contributors)

## License

The MIT License (MIT). Please see [License File](LICENSE) for more information.
