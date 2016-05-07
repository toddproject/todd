#!/bin/bash

# Copyright 2016 Matt Oswalt. Use or modification of this
# source code is governed by the license provided here:
# https://github.com/mierdin/todd:$branch/blob/master/LICENSE

# This script is designed to manage containers for ToDD. This could be start the basic infrastructure for ToDD like the etcd and rabbitmq containers,
# or you could run with the "integration" arg, and run integration tests as well.

DIR="$(cd "$(dirname "$0")" && pwd)"

branch="latest"

alias dtodd='docker run --rm --net todd-network --name="todd-client" mierdin/todd:$branch todd --host="todd-server.todd-network"'

# Clean up old containers
function cleanup {
    echo "Removing existing containers"
    for name in todd-server rabbit etcd influx grafana $(docker ps -aq --filter "label=toddtype=agent")
    do
        container=$(docker ps -q -f name="${name}" | wc -l)
        if [ $container -gt 0 ]; then
            docker kill "${name}" > /dev/null
        fi
        
        container=$(docker ps -aq -f name="${name}" | wc -l)
        if [ $container -gt 0 ]; then
            docker rm "${name}" > /dev/null
        fi
    done

    for id in $(docker ps -aq --filter "label=toddtype=agent")
    do
        container=$(docker ps -q -f id="${id}" | wc -l)
        if [ $container -gt 0 ]; then
            docker kill "${id}" > /dev/null
        fi
        
        docker rm "${id}" > /dev/null
    done
}

# Run infra containers
function startinfra {
    echo "Starting etcd"
    docker run -d --net todd-network -v /usr/share/ca-certificates/:/etc/ssl/certs -p 4001:4001 -p 2380:2380 -p 2379:2379 \
     --name etcd quay.io/coreos/etcd:v2.0.8 \
     -name etcd0 \
     -advertise-client-urls http://${HostIP}:2379,http://${HostIP}:4001 \
     -listen-client-urls http://0.0.0.0:2379,http://0.0.0.0:4001 \
     -initial-advertise-peer-urls http://${HostIP}:2380 \
     -listen-peer-urls http://0.0.0.0:2380 \
     -initial-cluster-token etcd-cluster-1 \
     -initial-cluster etcd0=http://${HostIP}:2380 \
     -initial-cluster-state new > /dev/null

    # I have a rabbitmq server at home, but in case I'm working on my laptop only, spin this up too:
    echo "Starting RabbitMQ"
    docker run -d \
        --net todd-network \
        --name rabbit \
        -p 8085:15672 \
        -p 5672:5672 \
        -e RABBITMQ_DEFAULT_USER=guest \
        -e RABBITMQ_DEFAULT_PASS=guest \
        rabbitmq:3-management > /dev/null

    echo "Starting InfluxDB"
    docker run -d --net todd-network --volume=/var/influxdb:/data --name influx -p 8083:8083 -p 8086:8086 tutum/influxdb:0.9 > /dev/null
    echo "Starting Grafana"
    docker run -d --net todd-network --volume=/var/lib/grafana:/var/lib/grafana --name grafana -p 3000:3000 grafana/grafana > /dev/null

}

# arg $1: number of agents to use
# arg $2: server config location
# arg $3: agent config location
function starttodd {
    echo "Starting todd-server"
    docker run -d -h="todd-server" -p 8081:8081 -p 8080:8080 -p 8090:8090 --net todd-network --name="todd-server" mierdin/todd:$branch todd-server --config="$2" > /dev/null

    i="0"
    while [ $i -lt $1 ]
    do
        echo "Starting todd-agent-${i}"
        docker run -d --label toddtype="agent" -h="todd-agent-$i" --net todd-network --name="todd-agent-$i" mierdin/todd:$branch todd-agent --config="$3" > /dev/null
        i=$[$i+1]
    done

}

function itsetup {

    # Upload grouping files
    cat $DIR/../docs/dsl/integration/group-inttest-red.yml | docker run -i --rm --net todd-network --name="todd-client" mierdin/todd:$branch todd --host="todd-server.todd-network" create
    cat $DIR/../docs/dsl/integration/group-inttest-blue.yml | docker run -i --rm --net todd-network --name="todd-client" mierdin/todd:$branch todd --host="todd-server.todd-network" create

    # Upload testrun files
    cat $DIR/../docs/dsl/integration/testrun-inttest-iperf.yml | docker run -i --rm --net todd-network --name="todd-client" mierdin/todd:$branch todd --host="todd-server.todd-network" create
    cat $DIR/../docs/dsl/integration/testrun-inttest-ping.yml | docker run -i --rm --net todd-network --name="todd-client" mierdin/todd:$branch todd --host="todd-server.todd-network" create
    
}

function runintegrationtests {

    # set -e

    sleep 20

    dtodd objects group

    dtodd objects testrun

    dtodd agents

    dtodd groups

    dtodd run inttest-ping -y -j

    dtodd run inttest-iperf -y -j

}

if hash docker-machaine 2> /dev/null
then
    echo "Starting docker machine"
    docker-machine start docker-dev
    eval $(docker-machine env docker-dev)

    export HostIP=$(docker-machine ip docker-dev)
    # Set HostIP to localhost if docker-machine doesn't run
    if [ $HostIP="" ]; then
        export HostIP="127.0.0.1"
    fi
fi

image=mierdin/todd:$branch
echo "Pulling image ${image}"
docker pull $image > /dev/null

if [ $(docker network ls | grep todd-network | wc -l) -lt 1 ]; then
    echo "Creating todd-network"
    docker network create todd-network
fi

cleanup

startinfra

if [ "$1" == "integration" ]
then
    sleep 5

    starttodd 6 /etc/todd/server-int.cfg /etc/todd/agent-int.cfg

    itsetup

    runintegrationtests
    exit $?
fi

if [ "$1" == "interop" ]
then
    sleep 5

    starttodd 3 /etc/todd/server-interop.cfg /etc/todd/agent-interop.cfg
    exit $?
fi

if [ -n "$1" ]
then
    sleep 5
    
    starttodd "$1" "$2" "$3"
fi