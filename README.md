# OAuth Service
OAuth service is an service for authentication and authorization, use for all 200lab project

# How to run
Clone the project to your favorite folder

Note: Please use go mod. If you're using go 1.11 or above, it supports go mod as default.

```
GO111MODULE=on // actually we don't need it
git clone github.com/200lab/oauth-service.git YOUR_FOLDER_PATH
cd YOUR_FOLDER_PATH
```

Use one of following ways to start the services:
### Use Docker-Compose

```
chmod +x start-in-docker.sh
./start-in-docker.sh
```
In case we don't want to re-build project:
```
./start-in-docker.sh nobuild
```

 Two services will be started: OAuth (port `3000`) and Mongodb (port `27017`)
 
### Run manually
#### Local MongoDB (for dev)
```
docker run -d --name mongo \
  -e MONGODB_USERNAME=oauth \
  -e MONGODB_PASSWORD=AuidwEyf776GG2S \
  -e MONGODB_DATABASE=oauth \
  -e MONGODB_ROOT_PASSWORD=200Lab2019! \
-p 27017:27017 \
bitnami/mongodb
```

#### OAuth
```
go build -o app
./app
```
Run with remote mongodb, easy to setup development env
```
go build -o app
MDB_MGO_URI="mongodb://oauth:AuidwEyf776GG2S@dev.db.200lab.io:27017/oauth_dev" ./app
```

# Test with fosite example
Here we need to start one more service to run as a client in OAuth2 Protocol

```
cd YOUR_FOLDER_PATH/oauth2/fosite-example
go build -o client
./client
```

A browser will be open automatically at port `3846`.

# OAuth Service Environments
Two either way to show all environment:
#### Without docker
``` 
./app outenv
```
#### With docker (after run docker-compose)
``` 
 docker run --rm oauth-service_app outenv
```

The result will look like
``` 
## gin mode (-gin-mode)
#GIN_MODE=

## disable default gin logger middleware (-gin-no-logger)
#GIN_NO_LOGGER=

## gin server Port. If 0 => get a random Port (-ginPort)
#GINPORT=3000

## gin server bind address (-ginaddr)
#GINADDR=

## init client id for oauth (-init-client-id)
#INIT_CLIENT_ID="200lab"

## init client secret for oauth (-init-client-secret)
#INIT_CLIENT_SECRET="secret-cannot-tell"

## init root password for client oauth (-init-root-password)
#INIT_ROOT_PASSWORD="Admin@2019"

## init root username for client oauth (-init-root-username)
#INIT_ROOT_USERNAME="admin"

## Log level: panic | fatal | error | warn | info | debug | trace (-log-level)
#LOG_LEVEL="debug"

## MongoDB ping check interval (-mdb-mgo-ping-interval)
#MDB_MGO_PING_INTERVAL=5

## MongoDB connection-string. Ex: mongodb://... (-mdb-mgo-uri)
#MDB_MGO_URI=

## oauth system secret key (-secret)
#SECRET="mrFPTI7EYOzt8CbcQVcUo2rIoLg97HI2"
```