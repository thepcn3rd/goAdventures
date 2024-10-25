#!/bin/bash

# All folders sync forward 
userName="kali"
#remoteIP="10.27.20.215"
remoteIP="10.0.2.4"

for folder in parrotWork ansible obsidian ensign
do
	rsync -arhvut $folder/ $userName@$remoteIP:/home/$userName/parrotWork
	#rsync -arhv $userName@$remoteIP:/home/$userName/parrotWork/ parrotWork
done

#rsync -arhv thepcn3rd@10.27.20.81:/home/thepcn3rd/homeVPN/ homeVPN

