
$release_dir="bin"

If (!(test-path $release_dir)){
    md $release_dir
}

go env -w GOOS="linux"
go build -o $release_dir/gate ./cmd/gate

go env -w GOOS="windows"
# go build -o $release_dir/gate.exe ./gate