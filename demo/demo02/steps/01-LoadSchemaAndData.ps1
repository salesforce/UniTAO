$scriptPath = $PSScriptRoot

Write-Host @"
#=====================================================================
# Import Schema all schemas without indexTemplate
#=====================================================================
"@

python $scriptPath\..\..\lib\python\submitDsData.py -m $scriptPath\..\data\dsMap.json -d $scriptPath\..\data\vmComputeSchema.json