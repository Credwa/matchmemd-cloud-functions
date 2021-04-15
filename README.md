# MatchMeMD Cloud functions

## Project setup and usage.

### Setup files for local development:

```
cp .env.example .env && cp matchmemd-staging-credentials.example.json matchmemd-staging-credentials.json
```
Replace the variable values

```
source /workspaces/matchmemd-cloud-functions/.env
```

```
export GOOGLE_APPLICATION_CREDENTIALS="/workspaces/matchmemd-cloud-functions/matchmemd-staging-credentials.json"
```

### Install dependencies:

```
go install
```

Run development server:

```
make dev
```

### Enabling unauthenticated access to cloud function after deployment

```
gcloud functions add-iam-policy-binding --member=allUsers --role=roles/cloudfunctions.invoker <FUNCTION-NAME> --region=<REGION>
```