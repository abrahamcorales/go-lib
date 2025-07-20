package oauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/gofiber/fiber/v2"
)

type Response struct {
	Message string `json:"message"`
}

type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

var jwks = Jwks{}
var ScopesRequired = []string{}
var AudiencesAllowed = []string{}

type JSONWebKeys struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

// getEnvironment returns the current environment from ENV variable
func getEnvironment() string {
	return os.Getenv("ENV")
}

// isLocal checks if the current environment is local
func isLocal(env string) bool {
	return env == "local" || env == "development"
}

var environment = getEnvironment()

// Initialize will initialize the oauth middleware
// It will fetch the jwks and set the required scopes
func Initialize() {
	if isLocal(environment) {
		return
	}

	var err error

	// retry 3 times
	for i := 0; i < 3; i++ {
		if i > 0 {
			//delay 5 seconds
			time.Sleep(5 * time.Second)
			log.Printf("Retrying to fetch jwks.json")
		}
		resp, err := http.Get(os.Getenv("AUTH_ISS") + ".well-known/jwks.json")
		if err != nil {
			log.Printf("Error initializing oauth library: %v", err.Error())
			continue
		}
		defer resp.Body.Close()
		err = json.NewDecoder(resp.Body).Decode(&jwks)
		if err != nil {
			log.Printf("Error initializing oauth library: %v", err.Error())
			continue
		}
		break
	}

	if err != nil {
		log.Panic("Fatal initializing oauth library: ", err.Error())
	}
	osAudiencesAllowed := os.Getenv("AUTH_AUDIENCE")
	if len(osAudiencesAllowed) < 1 {
		log.Panic("Fatal initializing oauth library: audiences not configured")
	}
	AudiencesAllowed = strings.Split(osAudiencesAllowed, ",")

	osScopeRequired := os.Getenv("AUTH_SCOPE_REQUIRED")
	if len(osScopeRequired) > 0 {
		ScopesRequired = strings.Split(osScopeRequired, " ")
	}
}

// Protected does check your JWT token and validates it
// If the token is valid or if the environment is local it will call the next handler
// If the token is invalid, it will return a 401
func Protected(c *fiber.Ctx) error {
	if checkJWT(c) || isLocal(getEnvironment()) {
		return c.Next()
	}
	return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized.")
}

// ProtectExternal checks if the request is external or not, if it is external it will check the JWT token
// If is internal it will inject the headers and call the next handler
func ProtectExternal(c *fiber.Ctx) error {
	isExternal := c.Get("X-External") == "true"
	if !isExternal {
		injectHeaders(c)
		injectClaimsWithoutValidation(c)
		return c.Next()
	}

	isValid := checkJWT(c)

	if isValid || isLocal(getEnvironment()) {
		return c.Next()
	}
	return c.Status(fiber.StatusUnauthorized).JSON(Response{Message: "Unauthorized."})
}

// DenyExternal checks if the request is external or not, if it is external it will return a 403
func DenyExternal(c *fiber.Ctx) error {
	isExternal := c.Get("X-External") == "true"
	if isExternal {
		return c.Status(fiber.StatusForbidden).JSON(Response{Message: "Forbidden."})
	}
	return c.Next()
}

// ExtractHeaders does not check for JWT token, but it will extract the headers and inject them into the context
func ExtractHeaders(c *fiber.Ctx) error {
	injectContextValues(c, nil)
	return c.Next()
}

func validationKeyGetter(token *jwt.Token) (interface{}, error) {
	// Verify all audience
	validAudience := false
	for _, v := range AudiencesAllowed {
		checkAud := token.Claims.(jwt.MapClaims).VerifyAudience(v, true)
		if checkAud {
			validAudience = true
			break
		}
	}
	if !validAudience {
		return token, errors.New("invalid audience")
	}

	// Verify 'iss' claim
	iss := os.Getenv("AUTH_ISS")
	checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, false)
	if !checkIss {
		return token, errors.New("invalid issuer")
	}

	cert, err := getPemCert(token)
	if err != nil {
		return nil, err
	}

	result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
	return result, nil
}

func getPemCert(token *jwt.Token) (string, error) {
	cert := ""

	for k := range jwks.Keys {
		if token.Header["kid"] == jwks.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		err := errors.New("unable to find appropriate key")
		return cert, err
	}

	return cert, nil
}

func extractor(c *fiber.Ctx) (string, error) {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return "", nil
	}
	authHeaderParts := strings.Split(authHeader, " ")
	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
		return "", errors.New("Authorization header format must be Bearer {token}")
	}
	return authHeaderParts[1], nil
}

func checkJWT(c *fiber.Ctx) bool {

	if isLocal(environment) {
		injectHeaders(c)
	}
	// Exctract token from header
	token, err := extractor(c)
	if err != nil || token == "" {
		fmt.Printf("Error extracting token: %v", err)
		return false
	}

	// Parse the token
	parsedToken, err := jwt.Parse(token, validationKeyGetter)
	if err != nil {
		fmt.Printf("Error parsing token: %v", err)
		return false
	}

	// Validate token
	if !parsedToken.Valid {
		fmt.Printf("Invalid token.")
		return false
	}

	// Get claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		fmt.Printf("Cannot get claims from token: %v", err)
		return false
	}

	// Validate required scopes
	if !hasValidScopes(claims) {
		return false
	}

	// Adding scopes and client_id to context
	injectContextValues(c, claims)

	// call injectHeaders func if internal audience
	if strings.Contains(claims["aud"].(string), "internal") {
		injectHeaders(c)
	}
	return parsedToken.Valid
}

func hasValidScopes(claims jwt.MapClaims) bool {
	if len(ScopesRequired) < 1 {
		return true
	}

	scopes := fmt.Sprintf("%v", claims["scope"])
	scopesSplitted := strings.Split(scopes, " ")

	for _, sc := range scopesSplitted {
		if contains(ScopesRequired, sc) {
			return true
		}
	}

	return false
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func injectContextValues(c *fiber.Ctx, claims jwt.MapClaims) {

	if claims["scope"] != nil {
		c.Locals("scopes", claims["scope"])
	}

	if claims["https://pomelo.la/client_id"] != nil {
		c.Locals("client_id", claims["https://pomelo.la/client_id"])
	}

	if claims["azp"] != nil {
		c.Locals("auth0_client", claims["azp"])
	}

	if claims["https://pomelo.la/email"] != nil {
		c.Locals("email", claims["https://pomelo.la/email"])
	}

	if claims["aud"] != nil {
		c.Locals("audience", mapAudienceClaim(claims)) // ["https://dashboard-auth-dev.pomelo.la", "x"]
	}
}

func mapAudienceClaim(claims jwt.MapClaims) []string {
	var audience []string
	switch audienceTyped := claims["aud"].(type) {
	case string:
		audience = append(audience, audienceTyped)
	case []string:
		audience = audienceTyped
	case []interface{}:
		for _, a := range audienceTyped {
			vs, ok := a.(string)
			if ok {
				audience = append(audience, vs)
			}
		}
	}
	return audience
}

func injectHeaders(c *fiber.Ctx) {

	if c.Get("X-Scopes") != "" {
		c.Locals("scopes", c.Get("X-Scopes"))
	}
	if c.Get("X-Client-Id") != "" {
		c.Locals("client_id", c.Get("X-Client-Id"))
	}
	if c.Get("X-Auth0-Client") != "" {
		c.Locals("auth0_client", c.Get("X-Auth0-Client"))
	}
}

func injectClaimsWithoutValidation(c *fiber.Ctx) {
	token, err := extractor(c)
	if err != nil || token == "" {
		return
	}

	parsedToken, _ := jwt.Parse(token, nil)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return
	}

	injectContextValues(c, claims)
}
