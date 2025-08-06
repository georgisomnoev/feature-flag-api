# Feature Flags API

## Description
FeatureFlagsAPI is an app that lets you manage feature flags. You can perform basic CRUD operations over the feature flag data. Access to this data is controlled by a predefined list of users with different permission levels.

## Architecture
The project is built using:
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Counterfeiter](https://github.com/maxbrunsfeld/counterfeiter) for mocks/fakes generation.
- [Gowrap](https://github.com/hexdigest/gowrap) for generating metric/trace decorators.
- [Ginkgo](https://onsi.github.io/ginkgo/) and [Gomega](https://onsi.github.io/gomega/) as a testing framework. 
- [Echo](https://echo.labstack.com) for a router.
- [PostgreSQL](https://www.postgresql.org) for a database store.
- [Migrate](https://github.com/golang-migrate/migrate/) for managing DB migrations.

## Application Workflow Overview
```
User --> [Post /auth] --> Auth Module
          --> Validate user credentials, generates and returns JWT
User --> [Get/Post/Put/Delete /flags] --> Feature Flags Module
          --> Middleware (Validates Token)
            --> CRUD Operations based on user scope
```

## Prerequsites
- Go (v1.24 was used during development)
- Ginkgo (for running the tests)


## Initial setup
- Download the dependencies:
```bash
make vendor
```
- Generate jwt keys:
```bash
make jwt-keys
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
curl -X POST http://127.0.0.1:8080/auth \
  -H "Content-Type: application/json" \
  -d '{
    "username": "mike",
    "password": "mike"
  }'
```

### To get a token with write access:
```bash
curl -X POST http://127.0.0.1:8080/auth \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john",
    "password": "john"
  }'
```

Note: Use the generated token as a Bearer token in the Authorization header when calling the feature flag endpoints.

### View Feature Flags (Read Access)

#### Get all feature flags:
```bash
curl -X GET http://127.0.0.1:8080/flags \
  -H "Authorization: Bearer <TOKEN>"
```

#### Get a single feature flag by ID:
```bash
curl -X GET http://127.0.0.1:8080/flags/<ID> \
  -H "Authorization: Bearer <TOKEN>"
```

### Manage Feature Flags (Write Access)

#### Create a new feature flag:
```bash
curl -X POST http://127.0.0.1:8080/flags \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "key": "new_feature_flag",
    "enabled": true,
    "description": "Description of the new feature flag"
  }'
```

#### Update an existing feature flag:
```bash
curl -X PUT http://127.0.0.1:8080/flags/<ID> \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "key": "updated_feature_flag",
    "enabled": false,
    "description": "Updated description"
  }'
```

#### Delete a feature flag:
```bash
curl -X DELETE http://127.0.0.1:8080/flags/<ID> \
  -H "Authorization: Bearer <TOKEN>"
```

## Future Enhancements:
- Proper validation for the api input fields.
- Group based access control for the feature flags. Currently all users have access to all the feature flags.
- Token refresh functionality.
- Token/user blacklisting to revoke access under specific conditions.
- Enable a feature toggle at a specific date.

Feedback and questions are welcome!