# go-instaman

## HTTP endpoints

This is a list of all the endpoints served by the `api-server` command.

### GET /instagram/me

This endpoint returns information about the account that is currently logged in via the `instaproxy` service.
Invoking this endpoint for the first time (after `instaproxy` is started or restarted) might take up to a minute due to the authentication and login workflow the proxy needs to go through.

Example response:

```json
{
    "biography": "The account bio",
    "fullName": "John Doe",
    "handler": "johndoe",
    "id": 123,
    "pictureURL": "https://cdninstagram.example.com/picture.png"
}
```

### GET /instagram/account/{name}

This endpoint returns information about the specified account.
The name parameter must be a non-blank account handler.

Example response:

```json
{
    "fullName": "John Doe",
    "handler": "john_doe",
    "id": 123,
    "pictureURL": "https://cdninstagram.example.com/picture.png"
}
```

### GET /instagram/account-id/{id}

This endpoint returns information about the specified account.
The id parameter must be a valid account identifier (as a 64bit integer).

Example response:

```json
{
    "fullName": "John Doe",
    "handler": "john_doe",
    "id": 123,
    "pictureURL": "https://cdninstagram.example.com/picture.png"
}
```

### GET /instagram/followers/{id}

This endpoint returns a paginated list of the users that follow the specified account.
The `id` parameter must be a valid account identifier (as a 64bit integer).

The optional `next_cursor` parameter is read from the query arguments, eg:

> GET /instagram/followers/123?next_cursor=abcde

Example response:

```json
{
    "next": "wxyz123",
    "users": [
        {
            "fullName": "John Doe",
            "handler":  "johndoe",
            "id": 45,
            "pictureURL": "https://cdninstagram.example.com/picture.png"
        },
        {
            "fullName": "Jane Doe",
            "handler":  "janedoe",
            "id": 56
        },
        {
            "fullName": "Name Surname",
            "handler":  "name_surname",
            "id": 67
        }
    ]
}
```

The `next` field represents the `next_cursor` that can be used for paginated searches. It is null or undefined if the search does not have any more pages to serve.

### GET /instagram/following/{id}

This endpoint returns a paginated list of the users that follow the specified account.
The `id` parameter must be a valid account identifier (as a 64bit integer).

The optional `next_cursor` parameter is read from the query arguments, eg:

> GET /instagram/following/456?next_cursor=abcde

Example response:

```json
{
    "next": "abc123",
    "users": [
        {
            "fullName": "John Doe",
            "handler":  "johndoe",
            "id": 12
        },
        {
            "fullName": "Jane Doe",
            "handler":  "janedoe",
            "id": 23,
            "pictureURL": "https://cdninstagram.example.com/picture.png"
        },
        {
            "fullName": "Name Surname",
            "handler":  "name_surname",
            "id": 34
        }
    ]
}
```

The `next` field represents the `next_cursor` that can be used for paginated searches. It is null or undefined if the search does not have any more pages to serve.

### GET /instagram/picture

This endpoint returns **binary data**: it is utilised as a proxy between the clients and Instagram, since the latter implements a[Cross-Origin Resource Sharing](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS) mechanism and therefore refuses to serve images to the browsers.

The picture URL is mandatory and needs to be specified via query argument.
It is mandatory to use HTTPS protocol and a subdomain of `cdninstagram.com` or else the API will serve an empty response with either 400 or 403 status code.

Example usage:

```html
<img
    alt=""
    src="/instagram/picture?pictureURL=https%3A%2F%2Fscontent-fco2-1.cdninstagram..."
    title="User profile picture"
/>
```
