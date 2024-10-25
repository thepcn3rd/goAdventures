#!/bin/bash
#
ip=172.31.1.10

#docker run --privileged --sysctl net.ipv6.conf.all.disable_ipv6=0 --rm -ti --name parrot -v $PWD/parrotWork:/work --net parrotNetwork --ip $ip parrot_v1

# Installed go lang - v4
# Installed the dependencies for go-windapsearch - v5
# Installed evil-winrm from github - v6
# Installed 'gem install wpscan' - v7
# Installed tshark - v8
# Installed crackmapexec - v9
	# Install CrackMapExec
	# https://www.crackmapexec.wiki/getting-started/installation/installation-on-unix

docker run --privileged -e DISPLAY=${DISPLAY} -v /tmp/.X11-unix:/tmp/.X11-unix --sysctl net.ipv6.conf.all.disable_ipv6=0 --rm -ti --name parrot -v $PWD/pWork:/work --net parrotNetwork --ip $ip parrot_v9
