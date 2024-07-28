
$protocbin="protoc.exe"

Write-Output $protocbin
& $protocbin --version

Get-ChildItem -Path . -Recurse -Filter *.proto | ForEach-Object {    
    $outputPath = $_.DirectoryName    
    Write-Output  $_.FullName
    & $protocbin --proto_path=$outputPath --go_out=../msg $_.FullName
}
