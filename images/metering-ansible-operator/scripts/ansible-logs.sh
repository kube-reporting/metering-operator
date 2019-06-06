#!/bin/bash

watch_dir=${1:-/tmp/ansible-operator/runner}
***REMOVED***lename=${2:-stdout}
mkdir -p ${watch_dir}


inotifywait -r -m -e create "${watch_dir}" | while read dir op ***REMOVED***le
do
  if [[ "${***REMOVED***le}" = "${***REMOVED***lename}" ]] ; then
    echo "${dir}/${***REMOVED***le}"
    (tail --follow=name "${dir}/${***REMOVED***le}" || true) &
  ***REMOVED***
done
