$link = "https://github.com/Plexcord/Installer/releases/latest/download/PlexcordInstallerCli.exe"

$outfile = "$env:TEMP\PlexcordInstallerCli.exe"

Write-Output "Downloading installer to $outfile"

Invoke-WebRequest -Uri "$link" -OutFile "$outfile"

Write-Output ""

Start-Process -Wait -NoNewWindow -FilePath "$outfile"

# Cleanup
Remove-Item -Force "$outfile"
