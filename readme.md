Why?
For whom?
Possible modifications

## Features
- [x] Golang
- [x] REST API with gin
  - [x] Handlers with dependency injection
  - [x] Default Health handler with DB ping check
  - [x] Swagger UI
  - [x] Swagger json generation with `swag init`
- [x] Auth
  - [x] Authentication with OAuth2 and JWT tokens
  - [x] Authentication configuration agnostic of provider
  - [x] Authorization with Open Policy Agent (OPA)
- [x] DB
  - [x] GORM
  - [x] Auto Migrations
  - [x] Postgres DB provider
  - [x] SQLite DB provider
- [ ] CI/CD
  - [x] Dockerfile
  - [x] Docker compose
  - [x] Kubernetes yaml
  - [ ] Github actions workflow

## Getting started

### Install deps
```cmd
go mod download
```

### ENV
> **IMPORTANT**: You should replace `AUTH_CONFIG_URL` and `AUTH_AUDIENCE` value with actual values from an OpenID provider.

Create a .env file in the root of the project with the following configs:

```env
ENV=local
WEB_PORT=localhost:8000
AUTH_CONFIG_URL=https://login.microsoftonline.com/00000000-0000-0000-0000-000000000000/v2.0/.well-known/openid-configuration
AUTH_AUDIENCE=api://c571ab3c-0fde-43b2-b010-77e7bdd0d6f7/api/
ENABLE_SWAGGER=true
DB_PROVIDER=postgres
DB_CONNECTION_STRING=host=localhost user=postgres password=mysecretpassword dbname=goapitemplate port=5432 sslmode=disable TimeZone=America/Denver
```
If you'd rather run with sqlite change `DB_PROVIDER` and `DB_CONNECTION_STRING`:

```env
DB_PROVIDER=sqlite
DB_CONNECTION_STRING=mydb.db
```

### Run
```powershell
cd goapi-template
go run .\main.go
```

### Test
Without test coverage:
```powershell
go test ./...
```

With HTML coverage output:
```powershell
go test -coverprofile=coverage ./...
go tool cover -html=coverage
```

### Build
```powershell
go build -o ./goapi-template ./main.go
```

## Run as container
### Docker
> **IMPORTANT**: You should replace `AUTH_CONFIG_URL` and `AUTH_AUDIENCE` value with actual values from an OpenID provider.

> **NOTE**: You will need docker installed and running

First set the `AUTH_CONFIG_URL`, `AUTH_AUDIENCE`, and `POSTGRES_PASSWORD` environment variables.

On Windows using PowerShell:
```powershell
$env:AUTH_CONFIG_URL = "https://login.microsoftonline.com/00000000-0000-0000-0000-000000000000/v2.0/.well-known/openid-configuration"
$env:AUTH_AUDIENCE = "api://00000000-0000-0000-0000-000000000000/api/"
$env:POSTGRES_PASSWORD = "mysecretpassword"
```

On Linux using bash:
```bash
export AUTH_CONFIG_URL="https://login.microsoftonline.com/00000000-0000-0000-0000-000000000000/v2.0/.well-known/openid-configuration"
export AUTH_AUDIENCE="api://00000000-0000-0000-0000-000000000000/api/"
export POSTGRES_PASSWORD="mysecretpassword"
```

Then, run docker compose:
```powershell
docker compose up
```

### Kubernetes
You may use minikube locally to test kubernetes configuration.

```powershell
kubectl create secret generic prod-db-secret --from-literal=username=produser --from-literal=password=Y4nys7f11


kubectl apply -f .\db-configmap.yaml
kubectl apply -f .\db-pvp.yaml
kubectl apply -f .\db-pv.yaml
kubectl apply -f .\db-deployment.yaml
kubectl apply -f .\db-service.yaml
kubectl apply -f .\app-configmap.yaml
kubectl apply -f .\app-deployment.yaml
kubectl apply -f .\app-service.yaml
```

#### Azure
```bash
az ad sp create-for-rbac --name lpains_github_sp \
                         --role contributor \
                         --scopes /subscriptions/4ead1c66-a55e-4cd7-babf-2e23ad5a6f39
```

## Authentication

## Authorization

## CI/CD