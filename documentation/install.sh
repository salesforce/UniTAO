#!/bin/sh
# This script installs the necessary required components for the project. 
# 
# It assumes the following is installed
# * NPM: If not, please install first.
# * homebrew is already installed. 
# 
# If git is needed then uncomment next line
# brew install git 
brew install go
brew install --cask docker
brew install docker-compose
mkdir ../test/data/DynamoDB/docker-compose/my-dynamodb-data
pushd ../test/data/DynamoDB/docker-compose
docker-compose up -d dynamodb
sudo npm install -g dynamodb-admin
