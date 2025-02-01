#!/bin/sh

# Smallweb: https://smallweb.run

if [ $# -eq 0 ]; then
  jq -n '{
    title: "SmallWeb",
    root: [
      { title: "Search Apps", type: "run", command: "search-apps" }
    ],
    commands: [
      {
        name: "search-apps",
        mode: "filter"
      }
    ]
  }'
  exit 0
fi

COMMAND=$1

if [ "$COMMAND" = "search-apps" ]; then
    smallweb ls --dir ~/smallweb/smallweb.run --json | jq 'map({
        title: .name,
        accessories: [.url],
        actions: [
            { title: "Open in Browser", type: "open", url: .url },
            { title: "Open Dir", type: "open", url: .dir },
            { title: "Copy URL", type: "copy", text: .url }
        ]
    }) | { items: . }'
fi
