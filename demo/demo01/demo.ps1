Write-Host @"
########################################################################################################################
# Demo 01: Schema based Data Service and Inventory Service
# Requirement: python3 + Curl
########################################################################################################################
"@

Pause

python $PSScriptRoot\steps\01-DataService01-ListSchema.py

Write-Host @"
########################################################################################################################
# Try to install VirtualMachine Record without Schema defined.
# Output: schema of type VirtualMachine not found
########################################################################################################################
"@

Pause

python $PSScriptRoot\steps\02-DataService01-AddVirtualMachineRecord.py

Write-Host @"
########################################################################################################################
# Load all basic schemas including schema for VirtualMachine
########################################################################################################################
"@
Pause

powershell $PSScriptRoot\steps\03-loadBasicSchema.ps1

python $PSScriptRoot\steps\01-DataService01-ListSchema.py

Write-Host @"
########################################################################################################################
# Try to install VirtualMachine Record again.
########################################################################################################################
"@

Pause

python $PSScriptRoot\steps\02-DataService01-AddVirtualMachineRecord.py

Pause

python $PSScriptRoot\steps\04-DataService01-AddVhdAndLinkToVM.py

Pause

python $PSScriptRoot\steps\05-InventoryServiceSchemaSync.py

Pause

python $PSScriptRoot\steps\06-AcrossDataServiceRef.py

Pause

python $PSScriptRoot\steps\07-InventoryServiceBasic.py

Pause

python $PSScriptRoot\steps\08-DataServcice01-AddSecondVm.py

Pause

python $PSScriptRoot\steps\09-InventoryServiceExplorerData.py
