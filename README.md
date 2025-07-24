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

## Prerequsites
- Go (v1.20+)
- Ginkgo CLI (for running the tests)

Before running the app please download the dependencies:
```bash
make vendor
```
Generate server certificates and jwt keys:
```bash
make generate-certs-and-keys
```

## How to run the server
To run the server using docker:
```bash
make run-docker
```

## How to run the tests
To run all the tests:

```bash
make test
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

