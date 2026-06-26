module github.com/wkqco/tdraw

go 1.26.1

require (
	golang.org/x/image v0.37.0
	golang.org/x/term v0.41.0
)

require (
	github.com/seoyc/wcli v0.0.0
	golang.org/x/sys v0.42.0 // indirect
)

replace github.com/seoyc/wcli => ./wcli
