#!/bin/bash

# rsync the data from parrotWork to pWork
# One folder at a time...
#rsync -arhv parrotWork/burpsuite/ pWork/burpsuite

# Change permissions for the files in pWork accessible to parrot
sudo chown -R thepcn3rd:thepcn3rd pWork

# All folders in parrotWork to pWork
rsync -arhvut parrotWork/ pWork
rsync -arhvut parrotWork/ pWorkBackup

# If changes were made in pWork sync with parrotWork
# Note: you need to delete files in both folders to make them disappear...
rsync -arhvut pWork/ parrotWork

# Change permissions back on the pWork folder for accessibility inside of parrot
sudo chown -R 1001:1001 pWork
