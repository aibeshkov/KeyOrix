#!/bin/bash

OUTPUT_FILE="largest-oldest-files.txt"

if stat --version &>/dev/null; then
  # GNU (Linux)
  find . -type f -print0 | \
    xargs -0 stat --format='%s %Y %n' 2>/dev/null
else
  # BSD/macOS
  find . -type f -print0 | \
    xargs -0 stat -f '%z %m %N' 2>/dev/null
fi | \
awk '{size=$1; mtime=$2; file=""; for(i=3;i<=NF;++i) file=file $i " "; printf "%.2f\t%d\t%s\n", size/1024/1024, mtime, file}' | \
sort -k1,1nr -k2,2n | \
awk '{ cmd="date -r "$2" +\"%Y-%m-%d\""; cmd | getline date; close(cmd); printf "%7.2f MB\t%s\t%s\n", $1, date, substr($0, index($0,$3)) }' > "$OUTPUT_FILE"

echo "Готово. Список самых больших и старых файлов сохранён в файл: $OUTPUT_FILE"