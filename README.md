# deprem-yardim-backend-go

# Proje Mimarisi

![architecture](/docs/architecture.png)

# Endpointler

### /feeds/areas

**Query Params**: `sw_lat` `sw_lng` `ne_lat` `ne_lng`

İşlenmiş lokasyon verisini afetharita.com adresine döner.

### /feeds/:id

**Path variable**: `id (int64)`

Tekil bir işlenmemiş twitter verisini döner.

### Run Locally

Redis: `docker run -d --name redis-stack-server -p 6379:6379 redis/redis-stack-server:latest`

### /monitor

![monitor](/docs/fiber-monitor.png)

### /metrics

![metrics](/docs/metrics.png)