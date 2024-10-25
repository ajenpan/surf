
$release_dir="bin"

If (!(test-path $release_dir)){
    mkdir $release_dir
}

go version

$current_os = $(go env GOOS)

$svrlist = @("gate", "battle")

# go env -w GOOS="linux"
# foreach ($svr in $svrlist) {
#     go build -o "$release_dir/$svr" "./cmd/$svr"
# }

# Build for Windows
go env -w GOOS="windows"
foreach ($svr in $svrlist) {
    go build -o "$release_dir/$svr.exe" "./cmd/$svr"
}

go env -w GOOS=$current_os