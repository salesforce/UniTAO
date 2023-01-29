<#
# ************************************************************************************************************
# Copyright (c) 2022 Salesforce, Inc.
# All rights reserved.

# UniTAO was originally created in 2022 by Shai Herzog & Yi Huo as an 
# Universal No-Coding Heterogeneous Infrastructure Maintenance & Inventory system that is holistically driven by open/community-developed semantic models/schemas.

# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as published
# by the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.

# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.

# You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <https://www.gnu.org/licenses/>

# This copyright notice and license applies to all files in this directory or sub-directories, except when stated otherwise explicitly.
# ************************************************************************************************************
#>

$scriptRoot = $PSScriptRoot

$dirName = $(Split-Path -Path $scriptRoot -Leaf)

$imageName = "UniTAO/$($dirName):localbuild".ToLower()

pushd $PSScriptRoot\..\..\
Write-Host "run docker build @[$(Get-Location)] with image [$imageName]"

docker pull golang:1.18
docker pull centos:latest

$dockerFilePath = "$scriptRoot\dockerfile" 

$imageName

$dockerFilePath

# Create the docker image with tag localbuild the image with same tag will be set as empty
docker build --no-cache -t $imageName -f $dockerFilePath --progress=plain .
# remove the empty image from previous command

$danglingImage = $(docker images --filter "dangling=true" -q --no-trunc)

if ($danglingImage) {
    Write-Host "delete dangling Images $danglingImage"
    docker rmi $danglingImage
}

popd