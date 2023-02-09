# deprem-yardim-backend-go

# Proje Mimarisi

![architecture](/docs/architecture.png)

# Endpointler

### /feeds/areas

**Query Params**: `sw_lat` `sw_lng` `ne_lat` `ne_lng` `timestamp`

İşlenmiş lokasyon verisini afetharita.com adresine lokasyon ve timestamp bilgisine göre döner. Eğer timestamp alanı boş geçilirse son 1 yıla ait kayıtlar döner.

### /feeds/:id

**Path variable**: `id (int64)`

Tekil bir işlenmemiş twitter verisini döner.

### Run Locally

Redis: `docker run -d --name redis-stack-server -p 6379:6379 redis/redis-stack-server:latest`

Grafana: `docker run --name grafana -i -p 3000:3000 grafana/grafana`
[Dashboard](https://grafana.com/grafana/dashboards/6671-go-processes/)

Prometheus: `docker run -it -d --name prometheus -p 9090:9090 -v $PWD:/etc/prometheus prom/prometheus --config.file=/etc/prometheus/prometheus.yml`

### /monitor

![monitor](/docs/fiber-monitor.png)

### /metrics

![metrics](/docs/metrics.png)