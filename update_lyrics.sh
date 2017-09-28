#!/bin/bash

#dependencies: mediainfo, eyeD3, metaflac

get_metadata() {
    mediainfo --Inform="General;artist=%Performer%\ntitle=%Title%\nalbum=%Album%" "$1"
}

get_tag() {
    local file="$1"
    local name="$2"

    mediainfo --Inform="General;%${name}%" "$file"
}

flac_remove_tag() {
    local file="$1"
    local name="$2"

    metaflac --remove-tag="${name}" "$file"
}

get_lyrics() {
    get_tag "$1" Lyrics
}

add_lyrics() {
    local mime="$1"
    local file="$2"
    local lyrics="$3"

    echo -n "Writing lyrics to ${file}..."

    case "$mime" in
        audio/x-flac)
            metaflac --set-tag="LYRICS=${lyrics}" "$file"
            echo "done!"
            ;;
        audio/mpeg)
            eyeD3 --add-lyrics "$lyrics"
            local tempFile="$(mktemp)"
            echo "$lyrics" > "$tempFile"
            eyeD3 --add-lyrics "$tempFile" "$file"
            rm "$tempFile"
            echo "done!"
            ;;
        *)
            echo "unknown mime ${mime}!"
            ;;
    esac
}

handle_file() {
    local file="$1"
    shift

    local mime="$(file --brief --mime-type "$file")"
    if [[ ! "$mime" =~ ^audio/ ]]
    then
        echo "Not an audio file!"
        return 1
    fi

    if [[ ! -z "$(get_lyrics "$file")" ]]
    then
        echo "Found lyrics!"
        return 1
    fi

    local flac_alt_tag="UNSYNCEDLYRICS"
    local existingLyrics="$(get_tag "$file" "$flac_alt_tag")"
    if [[ ! -z "$existingLyrics" ]]
    then
        echo -n "Lyrics found in ${flac_alt_tag} tag, migrating..."
        flac_remove_tag "$file" "$flac_alt_tag"
        add_lyrics "$mime" "$file" "$existingLyrics"
        echo "done!"
        return 0
    fi

    echo "No lyrics, searching..."
    for provider in "$@"
    do
        local lyrics
        lyrics="$(get_metadata "$file" | xargs -d"\n" lyrics2go "$provider")"
        if [[ $? == 0 ]]
        then
            add_lyrics "$mime" "$file" "$lyrics"
            return 0
        else
            echo "Lyrics lookup failed for ${provider}!"
        fi
    done
}


here="$(pwd)"
dir="${1:-$here}"
shift

find "$dir" -type f -print0 | while IFS= read -r -d '' file
do
    echo "File: $file"
    handle_file "$file" "$@"
done
