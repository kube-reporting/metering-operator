# This script post-processes helm template output

# Removes headers of the following form
# Example helm template header:
# ---
# # Source: foo-chart/templates/deployment.yaml
s/^---$//g
s/# Source.*$//g
# Delete trailing whitespace (spaces, tabs) from end of each line
s/[ 	]*$//
# Deletes all blank lines from top and end of file
/./,/^$/!d
