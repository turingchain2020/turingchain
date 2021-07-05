go env -w CGO_ENABLED=0
go build -o build/turingchain.exe github.com/turingchain2020/turingchain/cmd/turingchain
go build -o build/turingchain-cli.exe github.com/turingchain2020/turingchain/cmd/cli
