#### Результатом выполнения следующих домашних заданий является сервис «Календарь»:
- [Домашнее задание №12 «Заготовка сервиса Календарь»](./docs/12_README.md)
- [Домашнее задание №13 «Реализация Rest API Календаря»](./docs/13_README.md)
- [Домашнее задание №14 «Интеграция Apache Kafka в Календарь»](./docs/14_README.md)
- [Домашнее задание №15 «Докеризация и интеграционное тестирование Календаря»](./docs/15_README.md)
- [Домашнее задание №16 «Мониторинг Календаря»](./docs/16_README.md)

#### Ветки при выполнении
- `hw12_calendar` (от `master`) -> Merge Request в `master`
- `hw13_calendar` (от `hw12_calendar`) -> Merge Request в `hw12_calendar` (если уже вмержена, то в `master`)
- `hw14_calendar` (от `hw13_calendar`) -> Merge Request в `hw13_calendar` (если уже вмержена, то в `master`)
- `hw15_calendar` (от `hw14_calendar`) -> Merge Request в `hw14_calendar` (если уже вмержена, то в `master`)
- `hw16_calendar` (от `hw15_calendar`) -> Merge Request в `hw15_calendar` (если уже вмержена, то в `master`)

**Домашнее задание не принимается, если не принято ДЗ, предшедствующее ему.**

## API

API описано в спецификации OpenAPI V3: [`api/openapi.yaml`](api/openapi.yaml)

### Запуск сервера

```bash
make run
# или
go run cmd/calendar/main.go -config configs/config.toml
```

### Генерация кода

```bash
make generate
```

### Endpoints

- `GET /api/v1/events` - Получить все события
- `POST /api/v1/events` - Создать событие
- `GET /api/v1/events/{id}` - Получить событие по ID
- `PUT /api/v1/events/{id}` - Обновить событие
- `DELETE /api/v1/events/{id}` - Удалить событие
- `GET /api/v1/events/day/{date}` - Получить события на день
- `GET /api/v1/events/week/{date}` - Получить события на неделю
- `GET /api/v1/events/month/{date}` - Получить события на месяц

### Формат запросов

#### Создание события
```json
{
  "title": "Встреча",
  "start_time": "2026-03-27T10:00:00Z",
  "end_time": "2026-03-27T11:00:00Z",
  "user_id": 1,
  "description": "Описание встречи",
  "notify_before": 30
}
```

### Тесты

```bash
go test -v ./...
```
