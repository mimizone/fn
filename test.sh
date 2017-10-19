#!/bin/bash
# Top level test script to start all other tests

set -ex

export DB_USER=funcs
export DB_PASS=funcpass
export DB_DB=funcs

function host {
    case ${DOCKER_LOCATION:-localhost} in
    localhost)
        echo "localhost"
        ;;
    docker_ip)
        if [[ !  -z  ${DOCKER_HOST}  ]]
        then
            DOCKER_IP=`echo ${DOCKER_HOST} | awk -F/ '{print $3}'| awk -F: '{print $1}'`
        fi

        echo ${DOCKER_IP}
        ;;
    container_ip)
        echo "$(docker inspect -f '{{.NetworkSettings.IPAddress}}' ${1})"
        ;;
    esac
}

DB_CONTAINER="func-mysql-test"
docker rm -fv ${DB_CONTAINER} || echo No prev mysql test db container
docker run --name ${DB_CONTAINER} -p 3306:3306 -e MYSQL_DATABASE=$DB_DB \
  -e MYSQL_ROOT_PASSWORD=root -e MYSQL_USER=$DB_USER -e MYSQL_PASSWORD=$DB_PASS -d mysql
sleep 15
MYSQL_HOST=`host ${DB_CONTAINER}`
MYSQL_PORT=3306
MYSQL_URL="mysql://${DB_USER}:${DB_PASS}@tcp(${MYSQL_HOST}:${MYSQL_PORT})/${DB_DB}"

DB_CONTAINER="func-postgres-test"
docker rm -fv ${DB_CONTAINER} || echo No prev test db container
docker run --name ${DB_CONTAINER} -e "POSTGRES_DB=$DB_DB" \
  -e "POSTGRES_PASSWORD=$DB_PASS" -e "POSTGRES_USER=$DB_USER" -p 5432:5432 -d postgres
sleep 15
POSTGRES_HOST=`host ${DB_CONTAINER}`
POSTGRES_PORT=5432
POSTGRES_URL="postgres://${DB_USER}:${DB_PASS}@${POSTGRES_HOST}:${POSTGRES_PORT}/${DB_DB}?sslmode=disable"

go test -v $(go list ./... | grep -v vendor | grep -v examples | grep -v test/fn-api-tests)
go vet -v $(go list ./... | grep -v vendor)
docker rm --force func-postgres-test 
docker rm --force func-mysql-test

# test middlware, extensions, examples, etc
# TODO: do more here, maybe as part of fn tests
cd examples/middleware
go build
cd ../..
cd examples/extensions
go build
cd ../..
