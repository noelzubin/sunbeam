#!/bin/sh

# check if jq is installed
if ! [ -x "$(command -v jq)" ]; then
    echo "jq is not installed. Please install it." >&2
    exit 1
fi

if [ $# -eq 0 ]; then
    jq -n '{
        title: "Bitwarden Vault",
        description: "Search your Bitwarden passwords",
        commands: [
            {
                name: "list-passwords",
                title: "List Passwords",
                mode: "filter"
            }
        ]
    }'
    exit 0
fi

# check if bkt is installed
if ! [ -x "$(command -v bkt)" ]; then
    echo "bkt is not installed. Please install it." >&2
    exit 1
fi

if [ -z "$BW_SESSION" ]; then
    echo "Please set the BW_SESSION environment variable." >&2
    exit 1
fi

COMMAND=$(echo "$1" | jq -r '.command')
if [ "$COMMAND" = "list-passwords" ]; then
    bkt --ttl=1d -- bw --nointeraction list items --session "$BW_SESSION" | jq 'map({
        title: .name,
        subtitle: (.login.username // ""),
        actions: [
            {
                title: "Copy Password",
                type: "copy",
                text: (.login.password // ""),
            },
            {
                title: "Copy Username",
                type: "copy",
                text: (.login.username // ""),
            }
        ]
    }) | { items: .}'
fi
