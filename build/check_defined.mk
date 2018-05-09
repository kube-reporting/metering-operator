# https://stackoverflow.com/questions/10858261/abort-make***REMOVED***le-if-variable-not-set
# Check that given variables are set and all have non-empty values,
# die with an error otherwise.
#
# Params:
#   1. Variable name(s) to test.
#   2. (optional) Error message to print.
check_de***REMOVED***ned = \
    $(strip $(foreach 1,$1, \
        $(call __check_de***REMOVED***ned,$1,$(strip $(value 2)))))
__check_de***REMOVED***ned = \
    $(if $(value $1),, \
        $(error Unde***REMOVED***ned $1$(if $2, ($2))$(if $(value @), \
                required by target `$@')))
