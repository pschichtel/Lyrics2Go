#!/bin/bash
#set -x

unsyncedlyrics_tag="UNSYNCEDLYRICS"

get_metadata() {
    mediainfo --Inform="General;artist=%Performer%\ntitle=%Title%\nalbum=%Album%" "$1"
}

get_tag() {
    local mime="$1"
    local file="$2"
    local name="$3"

    case "$mime" in
        *)
            mediainfo --Inform="General;%${name}%" "$file"
            ;;
    esac

}

add_lyrics() {
    local mime="$1"
    local file="$2"
    local lyrics="$3"

    echo "Writing lyrics to ${file}..."

    case "$mime" in
        audio/x-flac)
            metaflac --set-tag=LYRICS="$lyrics" "$file"
            echo "done!"
            ;;
        *)
            echo "Unable to write to unknown mime ${mime}!"
            ;;
    esac
}

add_tag() {
    local mime="$1"
    local file="$2"
    local name="$3"
    local value="$4"

    echo "Adding tag ${name} to ${file}..."

    case "$mime" in
        audio/x-flac)
            metaflac --set-tag="${name}=${value}" "$file"
            echo "done!"
            ;;
        *)
            echo "Unable to add tag to unknown mime ${mime}!"
            ;;
    esac
}

remove_tag() {
    local mime="$1"
    local file="$2"
    local name="$3"

    echo "Removing tag ${name} from ${file}..."

    case "$mime" in
        audio/x-flac)
            metaflac --remove-tag="${name}" "$file"
            echo "done!"
            ;;
        *)
            echo "Unable to remove tag from unknown mime ${mime}!"
            ;;
    esac
}

get_lyrics() {
    get_tag "$1" "$2" Lyrics
}

handle_file() {
    local file="$1"
    shift

    local mime="$(file --brief --mime-type "$file")"
    if [[ ! "$mime" =~ ^audio/ ]]
    then
        return
    fi

    if [[ ! -z "$(get_lyrics "$mime" "$file")" ]]
    then
        echo "Found lyrics!"
    else
        echo "No lyrics, search them!"
        set -x
        for provider in "$@"
        do
            local existingLyrics="$(get_tag "$mime" "$file" "$unsyncedlyrics_tag")"
            if [[ ! -z "$existingLyrics" ]]
            then
                echo "Lyrics found in different tag: $unsyncedlyrics_tag"
                remove_tag "$mime" "$file" "$unsyncedlyrics_tag"
                add_lyrics "$mime" "$file" "$existingLyrics"
                return 0
            else
                local lyrics
                lyrics="$(get_metadata "$file" | xargs -d"\n" lyrics2go "$provider")"
                if [[ $? == 0 ]]
                then
                    add_lyrics "$mime" "$file" "$lyrics"
                    return 0
                else
                    echo "Lyrics lookup failed for ${provider}!"
                fi
            fi
        done
    fi
}


here="$(pwd)"
dir="${1:-$here}"
shift

find "$dir" -type f -print0 | while IFS= read -r -d '' file
do
    echo "File: $file"
    handle_file "$file" "$@"
done
