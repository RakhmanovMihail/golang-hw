# Собираем в гошке
FROM golang:1.25 as build

ENV CODE_DIR /go/src/
WORKDIR ${CODE_DIR}

# Кэшируем слои с модулями
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . ${CODE_DIR}

# Собираем статические бинарники (без CGO для alpine)
ARG LDFLAGS
RUN CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o /opt/calendar/calendar ./cmd/calendar && \
    CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o /opt/calendar/calendar_scheduler ./cmd/calendar_scheduler && \
    CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o /opt/calendar/calendar_storer ./cmd/calendar_storer

# На выходе тонкий образ
FROM alpine:3.19

LABEL ORGANIZATION="OTUS Online Education"
LABEL SERVICE="calendar"
LABEL MAINTAINERS="student@otus.ru"

WORKDIR /opt/calendar

COPY --from=build /opt/calendar/calendar .
COPY --from=build /opt/calendar/calendar_scheduler .
COPY --from=build /opt/calendar/calendar_storer .

# Конфиги
COPY ./configs/ /etc/calendar/

# По умолчанию запускаем API
ENV CONFIG_FILE /etc/calendar/config.toml
CMD ["./calendar", "-config", "/etc/calendar/config.toml"]
