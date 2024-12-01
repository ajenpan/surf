

# Get current script directory
$scriptPath = $PSScriptRoot
Write-Output "current script path: $scriptPath"

$workDir = Convert-Path "$scriptPath/.."

$protocbin = "protoc.exe"
Write-Output $protocbin
& $protocbin --version

$goOutDir = "$workDir"

#for csharp
$csharpOutDir = "$workDir/msg-cs"
mkdir $csharpOutDir -ErrorAction SilentlyContinue

$protofiles = Get-ChildItem -Path $(Join-Path $workDir "proto"), $(Join-Path $workDir "game") -Recurse -File -Filter *.proto
 
# for golang 
$protofiles | ForEach-Object {    
    $protoPath = $_.DirectoryName    
    Write-Output  $_.FullName
    & $protocbin --proto_path="$workDir"/tools/protoc/include/ --proto_path=$protoPath --go_out=$goOutDir $_.FullName
    & $protocbin --proto_path="$workDir"/tools/protoc/include/ --proto_path=$protoPath --csharp_out=$csharpOutDir --csharp_opt=file_extension=.pb.cs  $_.FullName
}
