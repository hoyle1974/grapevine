#!/bin/bash

export POSTGRES_PASSWORD=$(kubectl get secret --namespace default postgres-postgresql -o jsonpath="{.data.postgres-password}" | base64 -d)
export CMD=`cat schema.sql`

echo $CMD

kubectl run postgres-postgresql-client --rm --tty -i --restart='Never' --namespace default --image docker.io/bitnami/postgresql:15.3.0-debian-11-r0 --env="PGPASSWORD=$POSTGRES_PASSWORD" --env="CMD=$CMD" --command -- psql --host postgres-postgresql -U postgres -d postgres -p 5432 

