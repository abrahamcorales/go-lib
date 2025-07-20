
### How to use the `Protected` middleware

The `Protected` middleware is used to protect routes from unauthorized access. It will check the `Authorization` header for a valid OAuth token and if it is valid, it will set the `user` property on the request object to the user that the token belongs to.

```go
import (
    "github.com/abraham-corales/go-lib/oauth"
)

// Initialize the middleware
oauth.Initialize()

// unprotected endpoint
app.Get("/liveness", ping)

// protected endpoint
app.Get("/protected", oauth.Protected, Handler)
api := app.Group("/api", oauth.Protected)
```

### How to use the `ExtractScopes` middleware
The `ExtractScopes` middleware is used to extract the scopes from the `Authorization` header and set them on the request object.
This middleware does not check the validity of the token, it only extracts the scopes from the token.
The intended use is to use this middleware in requests that are already validated by previous proxies (for example, the `Istio Sidecar Proxy` or the `AWS Lambda authorization proxy`).

this will inject into the fiber context of the request the following properties:
- `scopes`: the scopes of the token
- `client_id`: the client id of the Pomelo client
- `auth0_client`: the client id of the Auth0 client that was used to generate the token
```go
// Initialize the middleware
oauth.Initialize()

// unprotected endpoint
app.Get("/liveness", ping)

// protected endpoint
app.Get("/endpoint", oauth.ExtractScopes, Handler)
api := app.Group("/api", oauth.ExtractScopes)
```

#### How to use in local environment
In order to use this middleware in local environment, you can send in the request the following headers:
- `X-Auth0-Client`: the client id of the Auth0 client that was used to generate the token
- `X-Scopes`: the scopes of the token
- `X-Client-Id`: the client id of the Pomelo client

If any of these headers are present in the request, the middleware will inject its value into the fiber context.


### How to use the `ExtractHeaders` middleware
The `ExtractHeaders` middleware is used to extract specific headers from the request and set them on the fiber context.
This middleware does not check validity of any token and it does not check the Authorization header.

this will inject into the fiber context of the request the following properties:
- `scopes`: the scopes of the token
- `client_id`: the client id of the Pomelo client
- `auth0_client`: the client id of the Auth0 client that was used to generate the token

### Environment Variables for OAuth

The oauth package uses 3 environment variables to configure the OAuth client.

| Variable              | Description                                                                                                    | Examples                                 |
|-----------------------|----------------------------------------------------------------------------------------------------------------|------------------------------------------|
| `AUTH_ISS`            | The client ID for the OAuth server                                                                             | `"https://pomelo-prod.us.auth0.com/"`    |
| `AUTH_AUDIENCE`       | The audience URL of the authentication                                                                         | `"https://auth-prod-internal.pomelo.la"` |
| `AUTH_SCOPE_REQUIRED` | The scopes that are needed in the client token in order to allow the request. The client needs ALL the scopes. | `"ajcc:all"` `"ajcc:all ajcc:cbk"`    |
