
$release_dir="bin"

If (!(test-path $release_dir)){
    md $release_dir
}

go env -w GOOS="linux"
go build -o $release_dir/surf ./cmd

go env -w GOOS="windows"
go build -o $release_dir/surf.exe ./cmd

# go env -w GOOS="linux"
# go build -o $release_dir/route ./cmd/route

# go env -w GOOS="windows"
# go build -o $release_dir/route.exe ./cmd/route