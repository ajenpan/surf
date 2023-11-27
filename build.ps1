
$release_dir="bin"

If (!(test-path $release_dir)){
    md $release_dir
}

go env -w GOOS="linux"
go build -o $release_dir/surf ./cmd

go env -w GOOS="windows"
go build -o $release_dir/surf.exe ./cmd
