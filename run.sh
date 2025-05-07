#!/bin/sh

while true; do
    # Run the Go script
    /root/go-app >> /root/logs/script.log 2>&1
    
    # Run logrotate
    /usr/sbin/logrotate /etc/logrotate.conf
    
    # Sleep for 5 minutes
    sleep 100
done
