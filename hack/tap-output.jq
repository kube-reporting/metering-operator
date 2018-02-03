(
    "1.." + (. | map(select(.Test != null and .Action == "run")) | length | tostring)
),

(
    map(select(.Test != null))
    |

    foreach .[] as $item ([1, empty];
        if $item.Action == "fail" then
          [.[0] + 1, "not ok \(.[0]) - \($item.Test)"]
        elif $item.Action == "pass" then
          [.[0] + 1,  "ok \(.[0]) - \($item.Test)" ]
        elif $item.Action == "skip" then
          [.[0] + 1, "ok \(.[0]) \($item.Test) # SKIP"]
        elif $item.Action == "output" then
          [.[0], ("# \($item.Output)" | rtrimstr("\n"))]
        else
          [.[0], empty]
        end;
    .[1]) | strings
)
