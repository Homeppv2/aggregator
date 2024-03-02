#!/bin/bash

set -e

hostpostgers="postgres"
portpostgres="5432"

hostredis="redis"
portredis="6380"

hostrabbitmq="rabbitmq"
portrabbitmq="5672"




cmd="$@"




# >&2 echo "!!!!!!!! Check conteiner_a for available !!!!!!!!"

until curl http://"$hostpostgers":"$portpostgres"; do
  # >&2 echo "Conteiner_A is unavailable - sleeping"
  sleep 3
done

until curl http://"$hostredis":"$portredis"; do
  # >&2 echo "Conteiner_A is unavailable - sleeping"
  sleep 3
done

until curl http://"$hostrabbitmq":"$portrabbitmq"; do
  # >&2 echo "Conteiner_A is unavailable - sleeping"
  sleep 3
done


# >&2 echo "Conteiner_A is up - executing command"

exec $cmd