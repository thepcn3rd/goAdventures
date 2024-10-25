#!/bin/bash

ip=172.31.1.10
user=thebabyn3rd

ssh -X -fT $user@$ip 'firefox -P default-esr &'
ssh -X -fT $user@$ip 'firefox -P Proxy &'
ssh -X -fT $user@$ip 'java -jar /work/burpsuite/burpsuite_community.jar &'

# Wait for firefox to launch
#ssh -X -fT $user@$ip 'jupyter-notebook --notebook-dir=/work/jnotebook --allow-root &'
