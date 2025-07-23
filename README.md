# Feature Flags API

## Description
FeatureFlagsAPI is an app that lets you manage feature flags. You can perform basic CRUD operation over the feature flag data. Access to this data is controlled by a predefined list of users with different permission levels.

## Architecture
The project is built using:
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Counterfeiter](https://github.com/maxbrunsfeld/counterfeiter) for mocks/fakes generation.
- [Ginkgo](https://onsi.github.io/ginkgo/) and [Gomega](https://onsi.github.io/gomega/) as a testing framework. 
- [Echo](https://echo.labstack.com) for a router.
- [PostgreSQL](https://www.postgresql.org) for a database store.
- [Migrate](https://github.com/golang-migrate/migrate/) for managing DB migrations.

## Prerequsites
Before running the app please download the dependencies:
```bash
make vendor
```
Generate certificates:
```bash
make generate-certs
```

## How to run the server

To run the server:
```bash
make run
```

To run the server using docker:
```bash
make run-docker
```

## How to run the tests

To run the unit tests:

```bash
make test
```
