#!/bin/bash

# Copyright 2021 Adevinta

set -e

function delete_dir {
	for var in "$@"
	do

		if [ -d "$var" ]
		then
            echo "Deleting folder $var/"
			rm -r "$var"/
        else
            if [ -e "$var" ]
    		then
                echo "Deleting file $var"
    			rm -r "$var"
    		fi
		fi
	done
}

# Delete auto-generated content
delete_dir app client tool swagger *.go
