
$protocbin = "protoc.exe"

Write-Output $protocbin
& $protocbin --version

# for golang 
Get-ChildItem -Path ../proto -Recurse -Filter *.proto | ForEach-Object {    
    $outputPath = $_.DirectoryName    
    Write-Output  $_.FullName
    & $protocbin --proto_path=$outputPath --go_out=. $_.FullName
}

#for csharp
mkdir -p msg-cs
Get-ChildItem -Path ../proto -Recurse -Filter *.proto | ForEach-Object {    
    $outputPath = $_.DirectoryName    
    Write-Output  $_.FullName
    & $protocbin --proto_path=$outputPath --csharp_out=./msg-cs/ --csharp_opt=file_extension=.pb.cs $_.FullName
}
