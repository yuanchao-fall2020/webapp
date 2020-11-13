#!/bin/bash

sudo /opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl \
    -a fetch-config \
    -m ec2 \
    -c file:/opt/cw-agent.json \
    -s

sudo /opt/gin_demo > /dev/null 2> /dev/null < /dev/null &
##/opt/gin_demo