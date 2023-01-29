
$scriptPath = $PSScriptRoot

Write-Host @"
#=====================================================================
# Import basic schemas to DataService01
#=====================================================================
"@

python $scriptPath\..\..\lib\python\submitDsData.py -m $scriptPath\..\data\dsMap.json -d $scriptPath\..\data\vmComputeSchema.json

