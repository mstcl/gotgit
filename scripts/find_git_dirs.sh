#!/bin/sh

fd --type d -a --hidden --follow --no-ignore --exclude '.local' --exclude '.cache' \
	--exclude 'sftpgo' --exclude '.steam' --base-directory="$HOME" --glob \.git | \
	sed -r "s/(^.+)\/.+$/\1/g"
