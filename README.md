# Go-Lib 🚀

A modern and comprehensive Go library that provides essential utilities for web application and backend service development.

## 📦 Included Packages

### 🔐 OAuth v2 (`oauthv2/`)
Complete OAuth2 authentication middleware with JWT, JWKS, and scope validation support.

```go
import "github.com/abraham-corales/go-aws/oauthv2"

// Initialize middleware
oauthv2.Initialize()

// Protect routes
app.Get("/protected", oauthv2.Protected, handler)
app.Get("/external", oauthv2.ProtectExternal, handler)
```

**Features:**
- ✅ Automatic JWT token validation
- ✅ JWKS (JSON Web Key Sets) support
- ✅ Scope and audience validation
- ✅ Local mode for development
- ✅ Automatic claims injection in context

### 🌐 REST Client (`rest/`)
Robust HTTP client with caching, retry, interceptors, and mocking support.

```go
import "github.com/abraham-corales/go-aws/rest"

// Default client
client := rest.NewDefaultRestClient()

// Custom client
config := rest.Config{
    BaseURL:         "https://api.example.com",
    TimeoutInMillis: 5000,
    Retries:         3,
}
client := rest.NewCustomRestClient(config)

// Make requests
response := client.Get("/users").
    WithCache(5 * time.Minute).
    WithHeader("Authorization", "Bearer token").
    Do()
```

**Features:**
- ✅ Automatic caching with configurable TTL
- ✅ Automatic retry system
- ✅ Request and response interceptors
- ✅ Mocking for testing
- ✅ Configurable timeouts
- ✅ Customizable headers

### 💾 Cache (`cache/`)
In-memory caching system with configurable TTL and expiration policies.

```go
import "github.com/abraham-corales/go-aws/cache"

// Create cache
cache := cache.NewMemoryCache("my-cache", 1000, 1*time.Hour, false)

// Basic operations
cache.Save(ctx, "key", "value")
expired, value := cache.Get(ctx, "key")
cache.Delete(ctx, "key")

// Cache with custom TTL
cache.SaveWithTTL(ctx, "key", "value", 30*time.Second)
```

**Features:**
- ✅ In-memory cache with configurable limit
- ✅ Default and custom TTL
- ✅ Expired item return policy
- ✅ Automatic operation logging

### 🔧 String Utils (`string_utils/`)
String manipulation utilities with template and tag support.

```go
import "github.com/abraham-corales/go-aws/string_utils"

// Replace tags in strings
template := "Hello {name}, your age is {age}"
data := []byte(`{"name": "John", "age": 30}`)
result, err := string_utils.ReplaceTags(ctx, template, data)
// Result: "Hello John, your age is 30"

// Extract tags from string
tags := string_utils.GetTagsFromString("Hello {name}, your age is {age}")
// Result: ["name", "age"]
```

### 📅 Date Utils (`date_utils/`)
Date and time handling utilities.

```go
import "github.com/abraham-corales/go-aws/date_utils"

// Get current time
now := date_utils.GetCurrentTime()

// Calculate difference in milliseconds
startTime := time.Now()
time.Sleep(100 * time.Millisecond)
diff := date_utils.GetTimeDiffInMillis(startTime)
```

### 🛠️ Helper (`helper/`)
Generic helper functions for common use cases.

```go
import "github.com/abraham-corales/go-aws/helper"

// Create pointers easily
strPtr := helper.Ptr("hello")
intPtr := helper.Ptr(42)
```

## 🚀 Installation

```bash
go get github.com/abraham-corales/go-aws
```

## 📋 Dependencies

### Main
- `github.com/gofiber/fiber/v2` - Web framework
- `github.com/golang-jwt/jwt/v4` - JWT handling
- `github.com/karlseguin/ccache` - In-memory cache
- `github.com/tidwall/gjson` - JSON parsing
- `github.com/gojek/heimdall/v7` - HTTP client with retry

### Testing
- `github.com/stretchr/testify` - Testing framework

## 🔧 Configuration

### Environment Variables for OAuth

| Variable | Description | Example |
|----------|-------------|---------|
| `ENV` | Execution environment | `local`, `development`, `production` |
| `AUTH_ISS` | OAuth server issuer | `https://your-domain.auth0.com/` |
| `AUTH_AUDIENCE` | Allowed audience | `https://api.example.com` |
| `AUTH_SCOPE_REQUIRED` | Required scopes | `read:users write:users` |

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Tests with coverage
go test -cover ./...

# Specific tests
go test ./cache
go test ./rest
```

## 📝 Usage Examples

### Web Server with OAuth

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/abraham-corales/go-aws/oauthv2"
)

func main() {
    app := fiber.New()
    
    // Initialize OAuth
    oauthv2.Initialize()
    
    // Public routes
    app.Get("/health", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{"status": "ok"})
    })
    
    // Protected routes
    api := app.Group("/api", oauthv2.Protected)
    api.Get("/users", getUsers)
    api.Post("/users", createUser)
    
    app.Listen(":3000")
}
```

### REST Client with Cache

```go
package main

import (
    "time"
    "github.com/abraham-corales/go-aws/rest"
)

func fetchUserData(userID string) {
    client := rest.NewDefaultRestClient()
    
    response := client.Get("/users/" + userID).
        WithCache(5 * time.Minute).
        WithHeader("Accept", "application/json").
        Do()
    
    if response.Error != nil {
        log.Printf("Error: %v", response.Error)
        return
    }
    
    log.Printf("Status: %d, Duration: %dms", response.StatusCode, response.Duration)
}
```

## 🤝 Contributing

1. Fork the project
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License. See the `LICENSE` file for details.

## 🐛 Reporting Bugs

If you find any bugs or have suggestions, please open an issue on GitHub.

---

**Built with ❤️ using Go**