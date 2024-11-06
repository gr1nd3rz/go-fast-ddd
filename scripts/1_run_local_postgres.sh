set -e

docker-compose -f ../deployments/docker-compose.yaml down --remove-orphans 
docker-compose -f ../deployments/docker-compose.yaml up -d --force-recreate 

source common.sh
success=0
while [ $success -eq 0 ]; do
    echo "Trying to apply the db schema..."
    cat ../assets/db.sql | $docker_exec && success=1 || success=0
    if [ $success -eq 0 ]; then
        echo "Waiting Postgres was started. Retry in 2 seconds..."
        sleep 2
    fi
done
echo "Postgres started and the db schema was initialized"