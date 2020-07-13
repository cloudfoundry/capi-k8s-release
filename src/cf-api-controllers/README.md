# cf-api-controllers

hello

## How does this thing work???

Here you go: https://miro.com/app/board/o9J_kvqLTy0=/

## How do I run my changes???

1. Ensure the cluster you want this controller to manage is targeted in your Kubeconfig
1. Simply run the binary: `go run ./main.go`
    - **Note**: You (for now) might see a bunch of logs processing old builds on startup

## How do I run tests???

1. Ensure you have installed [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) assets somewhere
    - Run `make test` to install kubebuilder dependencies to `/usr/local/kubebuilder/bin` (this will run as root to install there)
    - If you installed in `/usr/local/kubebuilder/bin`, you're all set
    - If you install it elsewhere, ensure you set `KUBEBUILDER_ASSETS` to that path
1. Simply run the tests: `go test -v ./...`
    - Can also continue to use `make test` to run tests as well

