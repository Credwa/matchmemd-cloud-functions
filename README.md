# MatchMeMD Cloud functions

## Project setup and usage.

### Setup files for local development:

```
cp .env.example .env && cp matchmemd-staging-credentials.example.json matchmemd-staging-credentials.json
```
Replace the variable values

```
source .env
```

### Install dependencies:

```
go install
```

Run development server:

```
go run cmd/main.go
```