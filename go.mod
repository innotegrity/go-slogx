module go.innotegrity.dev/slogx

go 1.18

require (
	github.com/fatih/color v1.15.0
	github.com/mattn/go-colorable v0.1.13
	github.com/slack-go/slack v0.12.2
	go.innotegrity.dev/async v0.1.1
	go.innotegrity.dev/errorx v1.0.15
	go.innotegrity.dev/generic v0.1.0
	go.innotegrity.dev/runtimex v0.1.0
	golang.org/x/exp v0.0.0-20230801115018-d63ba01acd4b
)

require (
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	golang.org/x/sys v0.6.0 // indirect
)

// TODO: remove before release
replace go.innotegrity.dev/generic => /Users/joshhogle/workspace/src/github.com/innotegrity/go-generic
