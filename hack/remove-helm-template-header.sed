# This script removes headers of the following form, trailing whitespace, and
# lines composing only of whitespace or newlines

# Example helm template header:
# ---
# # Source: helm-operator/templates/deployment.yaml

s/^---$//g
s/# Source.*$//g
s/ *$//; /^$/d; /^\s*$/d
