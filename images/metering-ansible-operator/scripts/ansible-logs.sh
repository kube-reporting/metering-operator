#!/bin/bash

watch_dir=${1:-/tmp/ansible-operator/runner}
filename=${2:-stdout}
mkdir -p ${watch_dir}


inotifywait -r -m -e create "${watch_dir}" | while read dir op file
do
  if [[ "${file}" = "${filename}" ]] ; then
    echo "${dir}/${file}"
    (tail --follow=name "${dir}/${file}" || true) &
  fi
done
