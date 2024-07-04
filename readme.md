## Features
- [x] Golang
  - [x] VS Code extension recommendations
- [x] REST API
  - [x] Handlers with dependency injection
  - [x] Default Health handler with DB ping check
  - [x] Swagger UI
  - [x] Swagger json generation with `swag init`
  - [x] Config from .env or environment variables
- [x] Auth
  - [x] Authentication with OAuth2 and JWT tokens
  - [x] Use .well-known/openid-configuration for configuration agnostic of provider
  - [x] Authorization via Open Policy Agent (OPA) policies
- [x] DB
  - ~~[x] GORM~~
  - [x] SQLC
  - [x] Automatic Migrations
  - [x] Postgres DB provider
  - ~~[x] SQLite DB provider~~
- [x] CI/CD
  - [x] Dockerfile
  - [x] Docker compose
  - [x] Kubernetes
  - [x] Github actions workflow

## How to use this template
This is a template repository so you can create a new repository based on this one. See instructions [here](https://docs.github.com/en/repositories/creating-and-managing-repositories/creating-a-repository-from-a-template).

## Getting started Locally
You should use the latest version available of Go. The current version used by this repository is 1.20.

### Install Go dependencies
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
DB_CONNECTION_STRING=postgres://postgres:94235CXcx@localhost:5432/goapitemplate?sslmode=disable
```
By default, the template uses Postres and thus you need it installed locally or available elsewhere.

### Run
```powershell
go run .\main.go
```

The API should now be available at http://localhost:8000/swagger/index.html

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
go build
```
In Windows, the command above yields a executable file `go-rest-template.exe`. In Linux, it yields an executable of same name but without extension.

## Deploy
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
docker compose up -d
```

### Kubernetes
You may use [minikube](https://minikube.sigs.k8s.io/docs/start/) locally to test kubernetes configuration.

```bash
kubectl create secret generic app-secrets --from-literal=AUTH_CONFIG_URL=<url> --from-literal=AUTH_AUDIENCE=<audience> --from-literal=DB_CONNECTION_STRING=<connection string>

kubectl create secret generic db-secrets --from-literal=POSTGRES_DB=<db name> --from-literal=POSTGRES_USER=<db user> --from-literal=POSTGRES_PASSWORD=<password>

kubectl apply -f ./db-pvc.yaml
kubectl apply -f ./db-pv.yaml
kubectl apply -f ./db-deployment.yaml
kubectl apply -f ./db-service.yaml
kubectl apply -f ./app-configmap.yaml
kubectl apply -f ./app-deployment.yaml
kubectl apply -f ./app-service.yaml
```

## Authentication
On startup, the application will execute an HTTP GET over the URL stored in `AUTH_CONFIG_URL` configuration. This variable should be a `.well-known/openid-configuration` endpoint which is typically provided by OAuth2 or OpenId providers such as:

|Provider|URL|Notes|
|-|-|-|
|Azure|https://login.microsoftonline.com/00000000-0000-0000-0000-000000000000/v2.0/.well-known/openid-configuration|The URL changes according to your Azure AD tenant id|
|Google|https://accounts.google.com/.well-known/openid-configuration||
|Facebook|https://www.facebook.com/.well-known/openid-configuration/||

The startup will automatically pull the issuer, jwks, and token signing algorithm. These fields are used to validate the JWT token. The jwks is also monitored for changes and is updated as needed. 

Additionally, the JWT is validated against the configured `AUTH_AUDIENCE` so only tokens intended for this API are accepted. The `AUTH_CLAIMS` configuration is used in order to lookup and add claims from the JWT body to the provided User interface so the app is aware of information such as user name, email, etc.

The authentication middleware will validate the JWT against the parameters set and allow (or not) the API pipeline to proceed. Any additional validation should be executed by the Authorization layer.

## Authorization
Authorization is provided via OPA policy with input fields method, path, and token. You may modify the `OpaMiddleware` to add more fields as necessary. The following basic policy is provided:

```opa
package authz

import future.keywords.if

default allow = false

allow if {
	endswith(payload.email, "@gmail.com")
	payload.verified
	startswith(input.path, "/person")
}

payload := {"verified": verified, "email": payload.email} if {
	[_, payload, _] := io.jwt.decode(input.token)
	verified := true
}
```

Note that the token input field is the full JWT provided by the consumer. You may decode it and use any of the provided fields such as Role, name, email, etc to validate whether the call is authorized or not.

The above basic policy enforces that the URL path must start with `/person` and the user email must end with `@gmail.com`. This is obviously just to get the authorization started and should be modified before using this template. For more information on OPA, please see https://www.openpolicyagent.org/.

The `OpaMiddleware` is a combined local PEP (Policy Enforcement Point) and PDP (Policy Decision Point). As such, any time your policy changes, you need a code change as the policy is stored locally, and a release. As your needs outgrow this approach, you should look into introducing a centralized PDP, adding a PIP (Policy Information Point) to enrich the policy inputs, and PAP (Policy Administration point) to create or modify policies without the need for a release.

## CI/CD
By default, this repository includes a single GitHub Actions workflow with 3 jobs that will:

1. Prepare and validate go code
2. Execute unit tests
3. Execute Linter
4. Determine Semver by using git commits
   1. See https://gitversion.net/docs/
5. Build, tag, and push docker image to docker hub
   1. You may want to change this step and push to a private repository
6. Deploy to Azure Kubernetes Service
7. Deploy to Azure Web App

## General recommendations
Before you push a similar solution to a production environment, keep the following recommendations in mind:

1. Kubernetes secrets are not really secret. If possible, you should leverage a secrets platform such as Azure Key Vault.
2. Favor managed database solutions instead of container based solution. Remember that if you use database in a container, you will need to take care of backups and other reliability related features that are typically available in managed solutions in cloud platforms such as AWS and Azure.
3. Avoid using the latest tag for releases. If a pod goes down and comes back up, it might use a different version of the image than the other containers of the same type.