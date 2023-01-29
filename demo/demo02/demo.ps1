Write-Host @"
###############################################################################################################
# Demo 02: AutoIndexing and Journal Process
# in this demo, we extended the element of [contentMediaType] again
# by introduce a attribute together [indexTemplate]
# the idea is that contentMediaType is referring to other dataType
# so Array/Map of [contentMediaType] can act as a registry of the referred dataType
# and indexTemplate provide the meaning of registry change to reflect data change from reality change
###############################################################################################################
"@

Pause

powershell $PSScriptRoot\steps\01-LoadSchemaAndData.ps1

Pause

python $PSScriptRoot\steps\02-JournalDataAndProcess.py

Pause

python $PSScriptRoot\steps\03-InventoryServiceSchemaSync.py

Pause

python $PSScriptRoot\steps\04-CmtIndexManual.py

Pause

python $PSScriptRoot\steps\05-CmtIndexAuto.py

Pause

python $PSScriptRoot\steps\06-CmtIndexConditionalIndex.py