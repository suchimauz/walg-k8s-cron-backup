# walg-k8s-cron-backup
Service for cron backup Postgres database with WAL-G in k8s cluster

## Usage

Create **.env** file and set support variables

## Kubernetes token

If you not have kubernetes Bearer Token but have access to kubectl cluster - you may get token from secret

```shell
kubectl --namespace kube-system get secret
```

If you use default user for access cluster - search **default-token-XXXXX**, or use your custom token
After get name you token and execute this command

```shell
kubectl --namespace kube-system describe secret default-token-XXXXX
```

And get **token** from response data
It's your Bearer Token for access Kubernetes API

## Support environment variables

```
APP_TIMEZONE=<tz_string> # default: UTC, example: Europe/Paris
APP_SAVE_LOGS=<boolean> # default: false

K8S_HOST=<host> # example: kube.domain.com or kube.domain.com:6443
K8S_INSECURE=<boolean> # default = true
K8S_AUTH_TOKEN=<token> # bearer token
K8S_NAMESPACE=<namespace> # namespace
K8S_LABEL_SELECTOR=<label_selector> # example: app=db,env=dev
K8S_POD_CONTAINER_NAME=<container_name> # example: backend

# Filestorage: use for example Minio
# For save backups info log file
# Required if APP_SAVE_LOGS is true
FS_HOST=<fs_host>
FS_BUCKET=<fs_bucket>
FS_ACCESS_KEY=<fs_access_key>
FS_SECRET_KEY=<fs_secret_key>
FS_SECURE=<boolean> # default = true

# example: wal-g backup-push <path_to_postgres_data>
# required
EXEC_BACKUP=<exec_backup> 
# example: wal-g backup-list --json --pretty --detail
# required when APP_SAVE_LOGS is true or TG_INFO_NOTIFICATION_ENABLED is true
EXEC_INFO=<exec_info>


TG_BOT_TOKEN=<bot_token>
TG_BACKUP_NOTIFICATION_ENABLED=true # default=false
# example: -1232345,2910434
TG_BACKUP_NOTIFICATION_CHATS=<chat_ids>

TG_INFO_NOTIFICATION_ENABLED=true # default=false
# optional | example: -1232345,2910434
TG_INFO_NOTIFICATION_CHATS=<chat_ids> 

# cron: Second | Minute | Hour | Dom | Month | Dow
# for execute EXEC_BACKUP command
# example: 0 0 21 * * *
# required
CRON_BACKUP=<cron_backup>
# for execute EXEC_BACKUP command
# example: 0 30 * * * *
# required when APP_SAVE_LOGS is true or TG_INFO_NOTIFICATION_ENABLED is true
CRON_INFO=<cron_info>
```

## Cron documentation

Field name   | Mandatory? | Allowed values  | Allowed special characters
----------   | ---------- | --------------  | --------------------------
Seconds      | Yes        | 0-59            | * / , -
Minutes      | Yes        | 0-59            | * / , -
Hours        | Yes        | 0-23            | * / , -
Day of month | Yes        | 1-31            | * / , - ?
Month        | Yes        | 1-12 or JAN-DEC | * / , -
Day of week  | Yes        | 0-6 or SUN-SAT  | * / , - ?
