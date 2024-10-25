#!/bin/bash
cd "$(dirname "$0")"

docker inspect fortimanagerhp >/dev/null 2>&1 || docker build . -t fortimanagerhp
# attaching pseudo terminal in interactive mode was only fix I could get working, --init also failed
docker run -it --rm -v ./:/var/run/fortimanagerhp -p 541:541 fortimanagerhp
