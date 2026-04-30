# backup-system

Backup management system for Mikrotik and database nodes, with a React dashboard.

---

## Running the Go backend

```bash
# From project root — copy dan isi .env dulu
cp .env.example .env

go build -o bin/web ./cmd/web
./bin/web
# Listens on WEB_LISTEN (default: 0.0.0.0:1909)
```

---

## Running the React frontend (dev)

```bash
cd frontend
npm install
npm run dev
# Buka http://localhost:1909
# /api/* dan /storage/* diproxy ke http://localhost:1909 (Go backend)
```

---

## Production deployment (nginx)

Build frontend dulu:

```bash
cd frontend
npm install
npm run build   # output ke frontend/dist/
```

Nginx config (`/etc/nginx/sites-available/backup-system`):

```nginx
server {
    listen 80;
    server_name backup.domain-kamu.com;

    root /root/opt/backup-system-new/frontend/dist;
    index index.html;

    # Proxy ke Go backend
    location /api/ {
        proxy_pass         http://127.0.0.1:1909;
        proxy_http_version 1.1;
        proxy_set_header   Host $host;
        proxy_set_header   X-Real-IP $remote_addr;
        proxy_set_header   X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
    }

    location /storage/ {
        proxy_pass         http://127.0.0.1:1909;
        proxy_http_version 1.1;
        proxy_set_header   Host $host;
    }

    # React SPA — semua route dikembalikan ke index.html
    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

```bash
sudo ln -s /etc/nginx/sites-available/backup-system /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
```

---

## Systemd service

```ini
# /etc/systemd/system/backup-web.service
[Unit]
Description=Backup System Web
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/root/opt/backup-system-new
ExecStart=/root/opt/backup-system-new/bin/web
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now backup-web
sudo systemctl status backup-web
```

---

## Update workflow (setelah git pull)

```bash
cd /root/opt/backup-system-new
git pull origin main

# Kalau ada perubahan Go:
go build -o bin/web ./cmd/web && sudo systemctl restart backup-web

# Kalau ada perubahan frontend:
cd frontend && npm install && npm run build
# nginx tidak perlu reload — langsung serve file baru
```

---

## First-time setup

1. Buat `.env` dari `.env.example`, isi semua nilai
2. Buat `nodes.json` dari `nodes.example.json`, isi credentials device
3. Jalankan Go backend
4. Buka browser → kamu akan diredirect ke `/setup`
5. Buat admin account pertama
6. Login dan mulai monitoring
