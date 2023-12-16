module github.com/ywak/zouni/testdata/case01

go 1.20

replace (
  example.com/lib => "./lib"
)

require golang.org/x/exp v0.0.0-20231127185646-65229373498e // indirect
require example.com/lib v0.0.0-00010101000000-000000000000
