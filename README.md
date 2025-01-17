# Chirpy

<!--toc:start-->
- [Chirpy](#chirpy)
  - [API](#api)
    - [Users](#users)
      - [POST /api/users](#post-apiusers)
      - [PUT /api/users](#put-apiusers)
      - [POST /api/login](#post-apilogin)
    - [Posts (Chirps)](#posts-chirps)
      - [GET /api/chirps?author_id={id}?sort=asc|desc](#get-apichirpsauthorididsortascdesc)
      - [GET /api/chirps/{id}](#get-apichirpsid)
      - [POST /api/chirps](#post-apichirps)
<!--toc:end-->

Chirpy is a simple blog engine written in Go:

- It allows storing users and their posts using a REST API.
- Builded using standard library
- Posts and users are stored in a [PostgreSQL](https://www.postgresql.org/) database.
- Passwords are hashed using [`bcrypt`](https://pkg.go.dev/golang.org/x/crypto/bcrypt).
- Authorization is done using [JSON Web Tokens](https://github.com/golang-jwt/jwt), that are refreshed every hour.
- Handle 'Polka' Webhook with authorization.

## API

### Users

#### POST /api/users

Creates a new user.

Parameters:

```json
{
  "email": "email@example.com",
  "password": "password"
}
```

Returns `201` if successful:

```json
{
  "id": "123e4567-e89b-12d3-a456-426655440000",
  "createdAt": "2021-01-01T00:00:00Z",
  "updatedAt": "2021-01-01T00:00:00Z",
  "email": "email@example.com"
}
```

#### PUT /api/users

Updates a user.

Parameters:

Headers: `Authorization: Bearer {token}`

```json
{
  "email": "email@example.com",
  "password": "password"
}
```

Returns `200` if successful:

```json
{
  "id": "123e4567-e89b-12d3-a456-426655440000",
  "createdAt": "2021-01-01T00:00:00Z",
  "updatedAt": "2021-01-01T00:00:00Z",
  "email": "email@example.com",
  "token": "JWT.Token.Structure",
  "refresh_token": "generated_string",
  "is_chirpy_red": true
}
```

#### POST /api/login

Logs in a user.

Parameters:

```json
{
  "email": "email@example.com",
  "password": "password"
}
```

Returns `200` if successful:

```json
{
  "id": "123e4567-e89b-12d3-a456-426655440000",
  "createdAt": "2021-01-01T00:00:00Z",
  "updatedAt": "2021-01-01T00:00:00Z",
  "email": "email@example.com",
  "token": "JWT.Token.Structure",
  "refresh_token": "generated_string"
}
```

### Posts (Chirps)

#### GET /api/chirps?author_id={id}?sort=asc|desc

Returns a list of posts:

- All if `author_id` is not set.
- By author if `author_id` is set.
- Sorted by creation date in `sort` order (`asc` by default).

```json
[
  {
    "id": "123e4567-e89b-12d3-a456-426655440000",
    "createdAt": "2021-01-01T00:00:00Z",
    "updatedAt": "2021-01-01T00:00:00Z",
    "body": "Hello, world!",
    "user_id": "123e4567-e89b-12d3-a456-426655440000"
  }
]
```

#### GET /api/chirps/{id}

Returns a post for current user.

```json
{
  "id": "123e4567-e89b-12d3-a456-426655440000",
  "createdAt": "2021-01-01T00:00:00Z",
  "updatedAt": "2021-01-01T00:00:00Z",
  "body": "Hello, world!",
  "user_id": "123e4567-e89b-12d3-a456-426655440000"
}
```

#### POST /api/chirps

Creates a new post for current user.

Headers: `Authorization: Bearer {token}`

Parameters:

```json
{
  "body": "Hello, world!"
}
```

Returns `201` if successful:

```json
{
  "id": "123e4567-e89b-12d3-a456-426655440000",
  "createdAt": "2021-01-01T00:00:00Z",
  "updatedAt": "2021-01-01T00:00:00Z",
  "body": "Hello, world!",
  "user_id": "123e4567-e89b-12d3-a456-426655440000"
}
```
