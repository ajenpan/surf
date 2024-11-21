

# Get current script directory
$scriptPath = $PSScriptRoot
Write-Output "current script path: $scriptPath"

$workDir = Convert-Path "$scriptPath/.."

$protoDir = Join-Path $workDir "proto"
Write-Output "workdir: $workDir, proto dir: $protoDir"

$protocbin = "protoc.exe"
Write-Output $protocbin
& $protocbin --version

$goOutDir = "$workDir/msg"

#for csharp
$csharpOutDir = "$workDir/msg-cs"
mkdir $csharpOutDir -ErrorAction SilentlyContinue

# for golang 
Get-ChildItem -Path $protoDir -Recurse -Filter *.proto | ForEach-Object {    
    $protoPath = $_.DirectoryName    
    Write-Output  $_.FullName
    & $protocbin --proto_path=$protoPath --go_out=$goOutDir $_.FullName
    & $protocbin --proto_path=$protoPath --csharp_out=$csharpOutDir --csharp_opt=file_extension=.pb.cs  $_.FullName
}
