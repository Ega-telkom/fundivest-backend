# REST API untuk fundivest

Status CI (Jangan digunakan di produksi jika gagal):
![CI](https://github.com/Ega-telkom/fundivest-backend/actions/workflows/ci.yml/badge.svg)

## Dokumentasi

Untuk membuat dokumentasi interaktif:
```
$ swag init -g cmd/api/main.go
```
```
$ make dev-up
$ make dev-worker
$ make dev-api
```

## TODO
### Tambah Logger
- [x] cmd/* (all)
- [x] handler/* (all)
- [ ] service/* (all)
- [x] worker/* (all)
- [ ] storage/* (all)
- [ ] queue/* (all)