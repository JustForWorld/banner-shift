## Запус сервиса
### 1. Локальный запуск
**1.1** Если необходимо запустить сервис с пользователем авторизованным как **администратор**, тогда в директории с проектом запустите команду:
```bash
make banner-admin-build
```
***
**1.2** Если необходимо запустить сервис с авторизованным **обычным** пользователем, тогда в директории с проектом запустите команду:
```bash
make banner-user-build
```
Так же реализована проверка получения информации о баннере (пара полей - фича и тег) и принадлежности пользователя к тегу (ограничение просмотра). При необхомости можно поменять у пользователя тег в файле user.yaml (`./config/mock/user.yaml`)
***

**1.3** Для получения JWT токен (каждый раз создается новый) необходимо в логах сервиса после сообщения `starting banner-shift` и до сообщения `starting server` найти и сохранить значение для дальнейшего использования в http запросах.

Пример JWT токена в логах сервиса:
```
level=DEBUG msg="current jwt token" jwt=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYWRtaW4iLCJ0YWciOjUsInVzZXJuYW1lIjoibi5uZXNtZXlhbm92YSJ9.ZHR4E2PlZhc1xeApNJGVtw7CV0zUxDm7B9tOfZs0vD4
```
***

**1.4** Перейдите в `postgres` контейнер с помощью команды:
```bash
sudo docker exec -it postgres psql -U postgres -d postgres
```
***

**1.5** Из директории `./scripts` скопируйте из файла `init.sql` код (или из сниппета ниже) и добавьте в `psql` для добавления значений по умолчанию в таблицы `tag` и `feature`:
```sql
DO $$
DECLARE
    i INT := 1;
BEGIN
    WHILE i <= 10 LOOP
        INSERT INTO feature DEFAULT VALUES;
        INSERT INTO tag DEFAULT VALUES;
        i := i + 1;
    END LOOP;
END $$;

```
***


### 2. Docker-compose
Убедитесь, что у вас не подняты контейнеры **redis** и **postgresql** после локальной сборки!

2.1 Запуск просиходит **администратора**, если необходимо запустить от **пользователя** поменяйте в `Dockerfile`:
- `CMD ["/banner-shift", "--config", "./config/docker.yaml", "--user", "./config/mock/admin.yaml"]`
- `CMD ["/banner-shift", "--config", "./config/docker.yaml", "--user", "./config/mock/user.yaml"]`

***

2.2 Запустить сервис с помощью команды и дождитесь когда контейнер `banner-shift` будет доступен:
```bash
sudo docker compose up --build -d
```
***

2.3 Для получения JWT токена пользовател зайдите в логи контейнера `banner-shift` с помощью команды:
```bash
sudo docker logs -f banner-shift
```
***

2.4 Перейдите в `postgres` контейнер с помощью команды:
```bash
sudo docker exec -it postgres psql -U postgres -d postgres
```
***

2.5 Из директории `./scripts` скопируйте из файла `init.sql` код (или из сниппета ниже) и добавьте в `psql` для добавления значений по умолчанию в таблицы `tag` и `feature`:
```sql
DO $$
DECLARE
    i INT := 1;
BEGIN
    WHILE i <= 10 LOOP
        INSERT INTO feature DEFAULT VALUES;
        INSERT INTO tag DEFAULT VALUES;
        i := i + 1;
    END LOOP;
END $$;
```
***

2.6 Остановить работу сервиса можно с помощью команды:
```bash
sudo docker compose down
```

## Работа с сервисом
### 1. Curl
**1.1** **GET** _/user_banner_ — получение баннера для пользователя:
```curl
curl -X GET "http://localhost:8080/user_banner?tag_id=7&feature_id=1&use_last_revision=true" \
-H "Authorization: Bearer {ВАШ_ТОКЕН_УБРАТЬ_CURLE_СКОБКИ}" 
```

**1.2** **GET** _/banner_ — получение всех баннеров c фильтрацией по фиче и/или тегу:
```bash
curl -X GET "http://localhost:8080/banner?feature_id=1&tag_id=7&limit=10&offset=10" \
-H "Authorization: Bearer {ВАШ_ТОКЕН_УБРАТЬ_CURLE_СКОБКИ}" 
```

**1.3** **POST** _/banner_ — создание нового баннера:
```bash
curl -X POST http://localhost:8080/banner \
-H "Content-Type: application/json" \
-H "Authorization: Bearer {ВАШ_ТОКЕН_УБРАТЬ_CURLE_СКОБКИ}" \
-d '{
    "tag_ids": [5, 6, 8],
    "feature_id": 2,
    "content": {
        "title": "some_title",
        "text": "some_text",
        "url": "some_url"
    },
    "is_active": true
}'
```

**1.4** **PATCH** _/banner/{id}_ — обновление содержимого баннера:
```bash
curl -X PATCH 'http://localhost:8080/banner/1' \
-H 'Content-Type: application/json' \
-H "Authorization: Bearer {ВАШ_ТОКЕН_УБРАТЬ_CURLE_СКОБКИ}" \
-d '{
    "tag_ids": [6, 8, 9],
    "feature_id": 2,
    "content": {
        "title": "some_title7",
        "text": "some_text7",
        "url": "some_url7"
    },
    "is_active": false
}'

```

**1.5** **PATCH** _/banner/{id}_ — обновление содержимого баннера:
```bash
curl -X DELETE "http://localhost:8080/banner/1" \
-H "Authorization: Bearer {ВАШ_ТОКЕН_УБРАТЬ_CURLE_СКОБКИ}"

```

### 2. Ход решения
Все возникшие проблемы (а лучше сказать задачи) и их решения описаны в каждом pull request:
- [Установка окружения и кофигурации](https://github.com/JustForWorld/banner-shift/pull/1)
- [Проектирование и работа с БД](https://github.com/JustForWorld/banner-shift/pull/2)
- [Разработка HTTP API handlers](https://github.com/JustForWorld/banner-shift/pull/3)
- [Mock авторизация с JWT](https://github.com/JustForWorld/banner-shift/pull/4)
- [Интеграция с Redis для кеширования](https://github.com/JustForWorld/banner-shift/pull/5)