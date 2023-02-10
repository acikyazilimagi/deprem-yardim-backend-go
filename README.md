# deprem-yardim-backend-go

# Proje Mimarisi

![architecture](/docs/architecture.png)

# Endpointler

### /feeds/areas

**Query Params**: `sw_lat` `sw_lng` `ne_lat` `ne_lng` `time_stamp`

İşlenmiş lokasyon verisini afetharita.com adresine lokasyon ve time_stamp bilgisine döner. Eğer timestamp alanı boş
geçilirse son 1 yıla ait kayıtlar döner.

**Örnek
Request** : `/feeds/areas?ne_lat=37.62633260711298&ne_lng=36.97311401367188&sw_lat=37.558254797440675&sw_lng=36.82479858398438&time_stamp=1675807028`

### /feeds/:id

**Path variable**: `id (int64)`

Tekil bir işlenmemiş twitter verisini döner.

### Run Locally

Redis: `docker run -d --name redis-stack-server -p 6379:6379 redis/redis-stack-server:latest`

Grafana: `docker run --name grafana -i -p 3000:3000 grafana/grafana`
[Dashboard](https://grafana.com/grafana/dashboards/6671-go-processes/)

Prometheus: `docker run -it -d --name prometheus -p 9090:9090 -v $PWD:/etc/prometheus prom/prometheus --config.file=/etc/prometheus/prometheus.yml`

## API vs Consumer Mode

Dockerfile contains 2 executables: `api` and `consumer`. One of the option can be selected via `--entrypoint` parameter.

After building docker image, in order to run api that contains fiber endpoints;

```shell
docker run --entrypoint "/api" <image_name>
```

In same way, if you want to run application in consumer mode, use following

```shell
docker run --entrypoint "/consumer" <image_name>
```

### /monitor

![monitor](/docs/fiber-monitor.png)

### /metrics

![metrics](/docs/metrics.png)

### Swagger

![swagger](/docs/swagger.png)
swagger klasörü altındaki dosyaları güncellemek için bash:

```
swag init -g cmd/api/main.go --output swagger .
```
