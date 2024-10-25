
$protocbin="protoc.exe"

Write-Output $protocbin
& $protocbin --version

# for golang 
Get-ChildItem -Path . -Recurse -Filter *.proto | ForEach-Object {    
    $outputPath = $_.DirectoryName    
    Write-Output  $_.FullName
    & $protocbin --proto_path=$outputPath --go_out=../ $_.FullName
}

# for csharp
# mkdir -p msg-cs
# Get-ChildItem -Path ./openproto -Recurse -Filter *.proto | ForEach-Object {    
#     $outputPath = $_.DirectoryName    
#     Write-Output  $_.FullName
#     & $protocbin --proto_path=$outputPath --csharp_out=./msg-cs/ $_.FullName
# }
