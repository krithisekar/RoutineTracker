Remove-Item function.zip
Remove-Item bootstrap
Write-Output "Deleted the files"
$env:GOOS = "linux"
$env:GOARCH = "amd64"  
$env:CGO_ENABLED = "0"
Write-Output "Environment variables have been set"

go build -ldflags="-s -w" -o bootstrap  main.go
Write-Output "Go build complete!"

Compress-Archive -Path .\bootstrap -DestinationPath .\function.zip
aws lambda update-function-code --function-name RoutineTracker --zip-file fileb://function.zip
Write-Output "Pushed build to AWS!"