# run godoc for repo outside of gopath
# ref: https://github.com/golang/go/issues/26827
function godoc() {
    if [ ! -f go.mod ]
    then
        echo "error: go.mod not found" >&2
        return
    fi

    module=$(sed -n 's/^module \(.*\)/\1/p' go.mod)
    docker run \
           --rm \
           -e "GOPATH=/tmp/go" \
           -p 127.0.0.1:6060:6060 \
           -v $PWD:/tmp/go/src/$module \
           golang \
           bash -c "echo http://localhost:6060/pkg/$module && godoc -http=:6060"
}
