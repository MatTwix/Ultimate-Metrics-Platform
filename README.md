# Ultimate Metrics Platform

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![Kafka](https://img.shields.io/badge/Apache%20Kafka-231F20?style=for-the-badge&logo=apachekafka&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-4169E1?style=for-the-badge&logo=postgresql&logoColor=white)
![MongoDB](https://img.shields.io/badge/MongoDB-4EA94B?style=for-the-badge&logo=mongodb&logoColor=white)
![Prometheus](https://img.shields.io/badge/Prometheus-E6522C?style=for-the-badge&logo=prometheus&logoColor=white)
![Grafana](https://img.shields.io/badge/Grafana-F46800?style=for-the-badge&logo=grafana&logoColor=white)
![Status](https://img.shields.io/badge/Status-In%20Development-blue?style=for-the-badge)
![Observability](https://img.shields.io/badge/Observability-Enabled-brightgreen?style=for-the-badge)
![License](https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge)

Комплексная микросервисная платформа для сбора метрик и обеспечения наблюдаемости, построенная на Go и включающая полный стек мониторинга с Prometheus, Grafana и Loki.

![Скриншот дэшборда Grafana](/assets/system-dashboard.png)

## О проекте

Проект представляет собой сквозную платформу, разработанную для демонстрации полного конвейера данных с использованием микросервисов на Go. Он позволяет наблюдать за поведением каждого компонента враспределенной системы и тенденциями в изменении внешних данных.

Проект служит практическим примером работы с техническими метриками в микросервисной архитектуре с использованием актуальных open-source решений для мониторинга.

## Архитектура

Платформа состоит из нескольких микросервисов, которые взаимодействуют через gRPC и брокер сообщений Kafka. Данные поступают от коллекторов через различные этапы обработки и в конечном итоге хранятся и визуализируются.

**Диаграмма потока данных:**

```md
Внешние API (GitHub, OpenWeather)
       |
       v
[ Collector Service ] -> (HTTP Метрики) -> [ Kafka ]
       |                                      |
       |--------------------------------------|
       |                                      |
       v                                      v
[ Cache Service ] <- (gRPC) - [ API Service ] |
       | (Redis)                              | (PostgreSQL)
       |                                      |
       v                                      v
[ Persister Service ]                         [ Notification Service ]
       | (PostgreSQL)                         | (Email)
       |                                      |
       v                                      v
[ Analytics Service ] <- (gRPC) --- [ API Service ]
        (MongoDB)
```

### Сервисы

* **Collector Service:** Собирает метрики из внешних источников (звезды на репозиториях GitHub, данные о температуре в определенном городе, аптайм выбранного сервиса) и публикует их в топик Kafka.
* **Kafka:** Выступает в роли брокера сообщений, разделяя производителей и потребителей и обеспечивая отказоустойчивый буфер для входящих данных.
* **Cache Service:** Потребляет метрики из Kafka и хранит их в Redis для быстрого доступа. Предоставляет gRPC-интерфейс для извлечения данных.
* **Persister Service:** Потребляет метрики из Kafka и сохраняет их в базу данных PostgreSQL для долгосрочного хранения.
* **API Service:** Предоставляет gRPC-интерфейс для запроса метрик из кэша или долгосрочной базы данных.
* **Analytics Service:** Периодически запрашивает API Service для выполнения агрегаций (например, почасовые средние значения, мин/макс) и сохраняет результаты в MongoDB.
* **Notification Service:** Потребляет метрики из Kafka и отправляет уведомления (например, по электронной почте) на основе предопределенных правил.

## Технологический стек

* **Бэкенд:** Go
* **Контейнеризация:** Docker & Docker Compose
* **Брокер сообщений:** Kafka
* **Базы данных:**
  * PostgreSQL (Долгосрочное хранение метрик)
  * MongoDB (Агрегированные аналитические данные)
  * Redis (Кэширование)
* **Мониторинг и метрики:**
  * **Prometheus:** Сбор технических метрик
  * **Grafana:** Визуализация (дэшборды для метрик и логов)
  * **Loki:** Агрегация логов
  * **Promtail:** Сбор логов

## CI/CD

Проект использует [GitHub Actions](https://docs.github.com/actions) для автоматизации процессов непрерывной интеграции и непрерывной доставки. Каждый пуш в ветку `main` запускает пайплайн, включающий два этапа:

### 1. Сборка (Build and Push)

Отвечает за сборку Docker-образов для всех микросервисов.

* **Среда выполнения:** `self-hosted`
* **Действия:**
  * **Checkout code:** Клонирование репозитория
  * **Set up QEMU & Docker Buildx:** Настройка инструментов для сборки Docker-образов
  * **Build and push [Service Name]:** Для каждого микросервиса выполняетсяс сборка Docker-образа. Образы тегируются как `latest` и загружаются в локальный кэш Docker раннера

### 2. Развертывание (Deploy)

Отвечает за развертывание сервисов на сервере после успешной сборки образов

* **Зависимость:** Запускается только после успешного завершения этапа `build-and-push`
* **Среда выполнения:** `self-hosted`
* **Окружение:** `env` = `production`
* **Действия:**
  * **Create .env:** Создает файл `.env` на сервере, используя защищенные переменные из GitHub secrets
  * **Deploy to server:** Запускает все сервисы с помощью `docker-compose up -d`, используя свежесобранные и загруженные в локальный кэш Docker-образы

**Полная конфигурация пайплайна доступна для ознакомления в файле `./.github/workflows/ci-cd.yml`*

## Начало работы

Следуйте этим шагам, чтобы запустить проект на вашей локальной машине.

### Предварительные требования

* [Docker](https://www.docker.com/get-started)
* [Docker Compose](https://docs.docker.com/compose/install/)

### Конфигурация

1. **Клонируйте репозиторий:**

    ```sh
    git clone https://github.com/MatTwix/Ultimate-Metrics-Platform
    cd Ultimate-Metrics-Platform
    ```

2. **Создайте файл окружения:**

    Скопируйте пример файла окружения `.env.example` и заполните необходимые значения.

3. **Отредактируйте `.env`:**
    Откройте файл `.env` и укажите необходимые данные.

### Запуск стека

Запустите все сервисы с помощью Docker Compose. Флаг `--build` соберет образы для Go-сервисов.

```sh
docker-compose up --build -d
```

Сервисы запустятся, и стек наблюдаемости начнет собирать данные.

## Использование

После запуска стека вы можете получить доступ к различным компонентам через ваш браузер:

* **Grafana:** `http://localhost:3000`
  * Войдите с учетными данными по умолчанию: `admin` / `admin`.
  * Дэшборды для всех сервисов и общий системный обзор настраиваются автоматически.

  **Пример дэшборда "Collector-Service":**
  ![Collector service Dashboard](/assets/collector-dashboard.png)
* **Prometheus UI:** `http://localhost:9090`
  * Исследуйте метрики и проверяйте статусы целей.
* **Kafdrop:** `http://localhost:19000`
  * Веб-интерфейс для просмотра топиков и сообщений Kafka.

## Мониторинг

Платформа полностью инструментирована для мониторинга и логирования.

### Метрики (Prometheus)

* Все микросервисы на Go предоставляют свои метрики по эндпоинту `/metrics`.
* Prometheus настроен на автоматический сбор этих эндпоинтов.
* Предварительно созданные дэшборды Grafana доступны для каждого сервиса, визуализируя ключевые показатели производительности (KPI), такие как частота запросов, частота ошибок и задержки.

### Логирование (Loki)

* Все сервисы настроены на вывод структурированных логов (JSON).
* **Promtail** автоматически собирает логи со всех запущенных Docker-контейнеров.
* Логи отправляются в **Loki** и могут быть запрошены и визуализированы в Grafana.
