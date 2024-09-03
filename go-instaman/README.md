# go-instaman

## HTTP endpoints

This is a list of all the endpoints served by the `api-server` command.

### GET /instaman/instagram/me

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

### GET /instaman/instagram/account/{name}

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

### GET /instaman/instagram/account-id/{id}

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

### GET /instaman/instagram/followers/{id}

This endpoint returns a paginated list of the users that follow the specified account.
The `id` parameter must be a valid account identifier (as a 64bit integer).

The optional `next_cursor` parameter is read from the query arguments, eg:

> GET /instaman/instagram/followers/123?next_cursor=abcde

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

### GET /instaman/instagram/following/{id}

This endpoint returns a paginated list of the users that follow the specified account.
The `id` parameter must be a valid account identifier (as a 64bit integer).

The optional `next_cursor` parameter is read from the query arguments, eg:

> GET /instaman/instagram/following/456?next_cursor=abcde

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

### GET /instaman/instagram/picture

This endpoint returns **binary data**: it is utilised as a proxy between the clients and Instagram, since the latter implements a[Cross-Origin Resource Sharing](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS) mechanism and therefore refuses to serve images to the browsers.

The picture URL is mandatory and needs to be specified via query argument.
It is mandatory to use HTTPS protocol and a subdomain of `cdninstagram.com` or else the API will serve an empty response with either 400 or 403 status code.

Example usage:

```html
<img
    alt=""
    src="/instaman/instagram/picture?pictureURL=https%3A%2F%2Fscontent-fco2-1.cdninstagram..."
    title="User profile picture"
/>
```

### GET /instaman/jobs

This endpoint finds a job in the database. It returns `null` instead of erroring if the job is not found.
However, it returns an error if neither `checksum` nor `id` are specified.

Query arguments:

- `id`: the job's primary key.
- `checksum`: the job's checksum.
- `state`: filter by job state.
- `type`: filter by job type.

Example response:

```json
{
    "id": 456,
    "checksum": "copy-followers:123456",
    "label": "Test job",
    "lastRun": "2024-01-01T12:00:00Z",
    "metadata": {
        "nextCursor": null,
        "userID": 123456
    },
    "nextRun": null,
    "state": "new",
    "type": "jobtype"
}
```

### GET /instaman/jobs/copy

This endpoint finds a copy job in the database. It returns `null` instead of erroring if the job is not found.
However, it returns an error if `direction` or `userID` are not specified.

Query arguments:

- `direction`: the connection's direction: either `followers` or `following`.
- `page`: if non-null, returns a paginated list of users along the response (key: `results`).
- `userID`: the Instagram account's ID connections are copied from.

Example response:

```json
{
    "id": 456,
    "checksum": "copy-followers:123456",
    "label": "Test label",
    "lastRun": "2025-01-01T12:00:00Z",
    "metadata": null,
    "nextRun": "2025-01-01T12:00:00Z",
    "state": "paused",
    "results": [
        {
            "firstSeen": "2025-08-01T12:00:00Z",
            "fullName": "John Doe",
            "handler": "john_doe",
            "id": 11,
            "lastSeen": "2025-08-10T00:00:00Z",
            "pictureURL": "https://cdninstagram.example.com/picture.png"
        },
        {
            "firstSeen": "2025-08-01T12:00:00Z",
            "fullName": "Jane Doe",
            "handler": "jane_doe",
            "id": 22,
            "lastSeen": "2025-08-10T00:00:00Z",
            "pictureURL": "https://cdninstagram.example.com/picture.png"
        }
    ],
    "resultsCount": 2,
    "type": "copy-followers"
}
```
