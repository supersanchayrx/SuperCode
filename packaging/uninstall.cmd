@echo off
powershell -NoProfile -ExecutionPolicy Bypass -Command "$d = Join-Path $env:LOCALAPPDATA 'Programs\SuperCode'; Remove-Item -Recurse -Force $d -ErrorAction SilentlyContinue; $p = ([Environment]::GetEnvironmentVariable('Path','User') -split ';' | Where-Object { $_ -and $_ -ne $d }) -join ';'; [Environment]::SetEnvironmentVariable('Path', $p, 'User'); Write-Host 'SuperCode removed. Your key file in %USERPROFILE%\.supercode was kept.'"
pause
