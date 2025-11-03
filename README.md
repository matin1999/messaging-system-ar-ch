
# messaging system

## requirements :

- kafka broker 
- mysql


## endpoints

```bash
/health
/account/createuser
/account/:user_id/services/status
/account/:user_id/services/create
/account/:user_id/services/charge
/account/:user_id/services/:service_id/messages
/sms/:user_id/:service_id/express/send
/sms/:user_id/:service_id/async/send

```
## Envs

```bash

LOG_LEVEL="info"
APP_PORT="8080"
PROMETHEUS_PORT="8181"
LOG_LEVEL="info"
KAVENEGAR_SMS_NUMBER=""
KAVENEGAR_SMS_API_KEY=""
MYSQL_ROOT_PASSWORD=test
MYSQL_USER=root
MYSQL_PASSWORD=test
MYSQL_DATABASE=postchi
DB_DSN=app:app@tcp(mysql:3306)/postchi?parseTime=true&loc=UTC&charset=utf8mb4,utf8
NATS_URL=nats://nats:4222
KAFKA_BROKERS=kafka:9092
KAFKA_TOPIC_SMS=sms_send
SMS_WORKER_COUNT=10
COST_PER_CHAR_EXPRESS=3
COST_PER_CHAR_ASYNC=1

```


