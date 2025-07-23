# Feature Flags API Requirements

## Objective
Build a fully functional CRUD API in Go for managing feature flags, with:
- OAuth 2.0 token validation (no mocks)
- Scope-based authorization (read:flags, write:flags)
- Relational database support
- Required unit and integration tests

## Domain
Your application will allow teams to toggle features dynamically using flags.
```json
{
  "id": "uuid",
  "key": "string",              // e.g. "enable_dark_mode"
  "description": "string",
  "enabled": "boolean",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

## Authentication & Authorization
Tokens must be validated (e.g., via introspection or public key verification).
Enforce scope-based access:
- `read:flags`: for GET endpoints
- `write:flags`: for POST, PUT, DELETE endpoints
You may use a public or self-hosted OAuth provider (e.g. Auth0, Keycloak, Okta…)

## Required API Endpoints
Method Path Description Required Scope
- GET /flags - List all flags, scope read:flags
- GET /flags/:id  - Get a flag by ID, scope read:flags
- POST /flags - Create a new flag, scope write:flags
- PUT /flags/:id - Update a flag, scope write:flags
- DELETE /flags/:id - Delete a flag, scope write:flags

## Testing (Required)
- Unit Tests
    - Core business logic and repository layer
    - Coverage of edge cases and errors
- Integration Tests
    - API tests including:
        - Token validation
        - Authz scope enforcement
        - Database interactions

## Tech Requirements
- Go 1.20+
- PostgreSQL or SQLite (for local/dev use)
- Any router (gin, echo, etc.)
- Real OAuth 2.0 token validation (JWT RS256 preferred)
- Clean, modular structure: handlers/, services/, repositories/, etc.

## Optional Enhancements
- Swagger/OpenAPI documentation
- Docker + Docker Compose
- Static code analysis
- CI pipeline (GitHub Actions, etc.)

## Deliverables
- Git repository
- README.md with:
    - Setup instructions
    - OAuth provider setup (client ID/secret, etc.)
    - How to run tests
    - Sample token and curl examples