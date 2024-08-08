#!/bin/bash

# set -o allexport
# source /backup.env
# set +o allexport

CONFIG_FILE="/backup.yml"

mc alias set myminio $MINIO_ENDPOINT $MINIO_ACCESS_KEY $MINIO_SECRET_KEY

backup() {
    local name=$1
    local source=$2
    local type=$3
    local bucket=$4
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local archive_name="${name}_backup_$timestamp.tar.gz"

    if [ "$type" == "folder" ]; then
        tar -czf /tmp/$archive_name -C "$source" .
    elif [ "$type" == "volume" ]; then
        docker run --rm -v ${source}:/data -v /tmp:/backup alpine \
            tar -czf /backup/$archive_name -C /data .
    fi

    mc cp /tmp/$archive_name myminio/$bucket/
    rm /tmp/$archive_name
}

schedule_backup() {
    local cron_schedule=$1
    local backup_cmd=$2
    (crontab -l; echo "$cron_schedule $backup_cmd") | crontab -
}

manual_backup() {
    local target_name=$1
    for backup_config in $(yq e '.backups[] | @base64' "$CONFIG_FILE"); do
        _jq() {
            echo ${backup_config} | base64 --decode | jq -r ${1}
        }

        name=$(_jq '.name')
        if [ "$name" == "$target_name" ]; then
            source=$(_jq '.source')
            type=$(_jq '.type')
            bucket=$(_jq '.bucket')
            backup "$name" "$source" "$type" "$bucket"
            exit 0
        fi
    done
    echo "Backup block '$target_name' not found in $CONFIG_FILE."
    exit 1
}

if [ $# -eq 1 ]; then
    manual_backup "$1"
    exit 0
fi

for backup_config in $(yq e '.backups[] | @base64' "$CONFIG_FILE"); do
    _jq() {
        echo ${backup_config} | base64 --decode | jq -r ${1}
    }

    name=$(_jq '.name')
    source=$(_jq '.source')
    type=$(_jq '.type')
    bucket=$(_jq '.bucket')
    cron_schedule=$(_jq '.cron')

    backup_cmd="/usr/local/bin/backup.sh $name $source $type $bucket"
    
    if [ "$cron_schedule" != "null" ]; then
        schedule_backup "$cron_schedule" "$backup_cmd"
    else
        backup "$name" "$source" "$type" "$bucket"
    fi
done

crond -f
