module github.com/nsxdevx/nsxbot

go 1.24.1

require (
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.3
	github.com/lmittmann/tint v1.1.1
	github.com/stretchr/testify v1.10.0
	github.com/tidwall/gjson v1.18.0
	golang.org/x/image v0.27.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

retract [v0.1.0, v0.1.5] // Not compatible with v0.2.x
