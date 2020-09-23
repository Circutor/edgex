#!/bin/sh -e
### BEGIN EdgeX Foundry
# Provides:          EdgeX Foundry
# Required-Start:    none
# Required-Stop:     none
# Should-Start:      none
# Should-Stop:       none
# Default-Start:     5
# Default-Stop:      0 6
# Short-Description: EdgeX Foundry.
### END INIT INFO
PATH="/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin"

DAEMON=/usr/bin/edgex
DAEMONOPT="-confdir /etc/edgex"
PIDFILE=/var/run/edgex.pid

# source function library
. /etc/init.d/functions

case "$1" in
  start)
	echo "Starting custom EdgeX Foundry"
	start-stop-daemon -b -m -S -p $PIDFILE -x $DAEMON -- $DAEMONOPT
	sleep 1

	;;
  stop)
	echo "Stoping custom EdgeX Foundry"
	start-stop-daemon -K -p $PIDFILE -x $DAEMON

	;;
  *)
	echo "Usage: /etc/init.d/edgex.sh {start|stop}"
	exit 1
esac

exit 0
