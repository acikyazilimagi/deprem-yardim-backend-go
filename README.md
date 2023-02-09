# deprem-yardim-backend-go

## Proje Mimarisi

![architecture](/docs/architecture.png)

## Endpointler

### /feeds/areas

**Query Params**: `sw_lat` `sw_lng` `ne_lat` `ne_lng` `time_stamp`

İşlenmiş lokasyon verisini afetharita.com adresine lokasyon ve time_stamp bilgisine döner. Eğer timestamp alanı boş geçilirse son 1 yıla ait kayıtlar döner.

**Örnek Request** : `/feeds/areas?ne_lat=37.62633260711298&ne_lng=36.97311401367188&sw_lat=37.558254797440675&sw_lng=36.82479858398438&time_stamp=1675807028`

### /feeds/:id

**Path variable**: `id (int64)`

Tekil bir işlenmemiş twitter verisini döner.

### /monitor

![monitor](/docs/fiber-monitor.png)

### /metrics

![metrics](/docs/metrics.png)

### Refresh Swagger

```sh
swag init --output swagger
```

## Run Locally

Gerekli bağımlılıklar topluca
[docker-compose](https://github.com/docker/compose) kullanılarak
çalıştırılabilir bunun için `docker compose up -d` komutunu kullanabilirsiniz.

Sonrasında `source go-dev.env` komutu ile ortam değişkenlerini ayarlayıp `go run
.` ile de programı çalıştırabilirsiniz.
