#!/bin/bash

get_metadata() {
    mediainfo --Inform="General;artist=%Performer%\ntitle=%Title%\nalbum=%Album%" "$1"
}

get_lyrics() {
    mediainfo --Inform='General;%Lyrics%' "$1"
}

write_lyrics() {
    echo "$1"
}

here="$(pwd)"
dir="${1:-$here}"
shift

handle_file() {
    local file="$1"
    shift

    if [[ ! "$(file --brief --mime-type "$file")" =~ ^audio/ ]]
    then
        return
    fi

    if [[ ! -z "$(get_lyrics "$file")" ]]
    then
        echo "Found lyrics!"
    else
        echo "No lyrics, search them!"
        for provider in "$@"
        do
            local lyrics="$(get_metadata "$file" | xargs -d"\n" lyrics2go "$provider")"
            if [[ $? == 0 ]]
            then
                write_lyrics "$lyrics"
                break
            else
                echo "Lyrics lookup failed for ${provider}!"
            fi
        done
    fi
}

export -f handle_file

find "$dir" -type f -print0 | while IFS= read -r -d '' file
do
    handle_file "$file" "$@"
done
