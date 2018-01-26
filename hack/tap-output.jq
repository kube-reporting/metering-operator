if .Test !=null then
  .
***REMOVED***
  empty
end

|

if .Action == "fail" then
  "not ok # \(.Test)"
elif .Action == "pass" then
  "ok # \(.Test)"
elif .Action == "skip" then
  "ok # skip \(.Test)"
elif .Action == "output" then
  "# \(.Output)" | rtrimstr("\n")
***REMOVED***
  empty
end
