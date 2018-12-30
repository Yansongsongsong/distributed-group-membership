### Usage:
For macOS, 
`./src/main/dgm -h`

### how to build?
```shell
# open your terminal for Unix, just copy that
cd distributed-group-membership
export GOPATH=$PWD:$GOPATH
# get the info of your system
go env | grep GOOS
# output: GOOS="darwin"
go env | grep GOARCH
# output: GOARCH="amd64"
# with result below to build
GOOS=darwin GOARCH=amd64 go build -o dgm ./src/main/main.go
```

then you can find it in `/distributed-group-membership/dgm`