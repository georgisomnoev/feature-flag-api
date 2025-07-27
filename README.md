# Feature Flags API

## Description
FeatureFlagsAPI is an app that lets you manage feature flags. You can perform basic CRUD operations over the feature flag data. Access to this data is controlled by a predefined list of users with different permission levels.

## Architecture
The project is built using:
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Counterfeiter](https://github.com/maxbrunsfeld/counterfeiter) for mocks/fakes generation.
- [Ginkgo](https://onsi.github.io/ginkgo/) and [Gomega](https://onsi.github.io/gomega/) as a testing framework. 
- [Echo](https://echo.labstack.com) for a router.
- [PostgreSQL](https://www.postgresql.org) for a database store.
- [Migrate](https://github.com/golang-migrate/migrate/) for managing DB migrations.

## Application Workflow Overview
```
User --> [Post /auth] --> Auth Service (Generates and returns JWT)
       --> [Get/Post/Put /flags] --> Feature Flags API
             --> Middleware (Validates Token)
                   --> CRUD Operations (Role/Scope filtering)
```

## Prerequsites
- Go (v1.24 was used during development)
- Ginkgo CLI (for running the tests)


## Initial setup
- Download the dependencies:
```bash
make vendor
```
- Generate server certificates and jwt keys:
```bash
make generate-certs-and-keys
```
- Create a `.env` file. Feel free to use the `.env.example`.

## How to run the app
To run the app with docker:
```bash
make run-app
```

## How to run the tests
To run the unit tests:

```bash
make test-unit
```

To run the integration tests:

```bash
make test-integration
```

## Example requests
To use the feature flags API, you need to authenticate first.
Each user has a different permission level. For testing, you can use one of these users:

| Username   | Role   |
| ---------- | ------ |
| ted        | viewer |
| mike       | viewer |
| john       | editor |
| uncle_bob  | editor |

The password is the same as the username.

### To get a token with read-only access:
```bash
curl -X POST https://127.0.0.1:8443/auth \
  -H "Content-Type: application/json" \
  -k \
  -d '{
    "username": "mike",
    "password": "mike"
  }'
```

### To get a token with write access:
```bash
curl -X POST https://127.0.0.1:8443/auth \
  -H "Content-Type: application/json" \
  -k \
  -d '{
    "username": "john",
    "password": "john"
  }'
```

Note: Use the generated token as a Bearer token in the Authorization header when calling the feature flag endpoints.

### View Feature Flags (Read Access)

#### Get all feature flags:
```bash
curl -X GET https://127.0.0.1:8443/flags \
  -H "Authorization: Bearer <TOKEN>" \
  -k
```

#### Get a single feature flag by ID:
```bash
curl -X GET https://127.0.0.1:8443/flags/<ID> \
  -H "Authorization: Bearer <TOKEN>" \
  -k
```

### Manage Feature Flags (Write Access)

#### Create a new feature flag:
```bash
curl -X POST https://127.0.0.1:8443/flags \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -k \
  -d '{
    "key": "new_feature_flag",
    "enabled": true,
    "description": "Description of the new feature flag"
  }'
```

#### Update an existing feature flag:
```bash
curl -X PUT https://127.0.0.1:8443/flags/<ID> \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -k \
  -d '{
    "key": "updated_feature_flag",
    "enabled": false,
    "description": "Updated description"
  }'
```

#### Delete a feature flag:
```bash
curl -X DELETE https://127.0.0.1:8443/flags/<ID> \
  -H "Authorization: Bearer <TOKEN>" \
  -k
```

## Future Enhancements:
- Token refresh functionality
- Depending on the environment, the app's web API communication could use just `http`.

Feedback and questions are welcome!