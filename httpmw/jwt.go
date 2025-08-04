package httpmw

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/qoinlyid/qore"
)

// JwtConfig defines the config for JWT middleware.
type JwtConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper
	// SuccessHandler defines a function which is executed for a valid token.
	SuccessHandler func(c qore.HttpContext)
	// ErrorHandler defines a function which is executed when all lookups have been done and none of them passed Validator
	// function. ErrorHandler is executed with last missing (ErrExtractionValueMissing) or an invalid key.
	// It may be used to define a custom JWT error.
	//
	// Note: when error handler swallows the error (returns nil) middleware continues handler chain execution towards handler.
	// This is useful in cases when portion of your site/api is publicly accessible and has extra features for authorized users
	// In that case you can use ErrorHandler to set default public JWT token value to request and continue with handler chain.
	ErrorHandler func(c qore.HttpContext, err error) error
	// Context key to store user information from the token into context.
	// Optional. Default value "auth".
	ContextKey string
	// Signing key to validate token.
	// This is one of the three options to provide a token validation key.
	// The order of precedence is a user-defined KeyFunc, SigningKeys and SigningKey.
	// Required if neither user-defined KeyFunc nor SigningKeys is provided.
	SigningKey any
	// Map of signing keys to validate token with kid field usage.
	// This is one of the three options to provide a token validation key.
	// The order of precedence is a user-defined KeyFunc, SigningKeys and SigningKey.
	// Required if neither user-defined KeyFunc nor SigningKey is provided.
	SigningKeys map[string]any
	// Signing method used to check the token's signing algorithm.
	// Optional. Default value HS256.
	SigningMethod string
	// KeyFunc defines a user-defined function that supplies the public key for a token validation.
	// The function shall take care of verifying the signing algorithm and selecting the proper key.
	// A user-defined KeyFunc can be useful if tokens are issued by an external party.
	// Used by default ParseTokenFunc implementation.
	//
	// When a user-defined KeyFunc is provided, SigningKey, SigningKeys, and SigningMethod are ignored.
	// This is one of the three options to provide a token validation key.
	// The order of precedence is a user-defined KeyFunc, SigningKeys and SigningKey.
	// Required if neither SigningKeys nor SigningKey is provided.
	// Not used if custom ParseTokenFunc is set.
	// Default to an internal implementation verifying the signing algorithm and selecting the proper key.
	KeyFn jwt.Keyfunc
	// TokenLookup is a string in the form of "<source>:<name>" or "<source>:<name>,<source>:<name>" that is used
	// to extract token from the request.
	// Optional. Default value "header:Authorization".
	// Possible values:
	// - "header:<name>" or "header:<name>:<cut-prefix>"
	// 			`<cut-prefix>` is argument value to cut/trim prefix of the extracted value. This is useful if header
	//			value has static prefix like `Authorization: <auth-scheme> <authorisation-parameters>` where part that we
	//			want to cut is `<auth-scheme> ` note the space at the end.
	//			In case of JWT tokens `Authorization: Bearer <token>` prefix we cut is `Bearer `.
	// If prefix is left empty the whole value is returned.
	// - "query:<name>"
	// - "param:<name>"
	// - "cookie:<name>"
	// - "form:<name>"
	// Multiple sources example:
	// - "header:Authorization:Bearer ,cookie:myowncookie"
	TokenLookup string
	// ParseTokenFunc defines a user-defined function that parses token from given auth. Returns an error when token
	// parsing fails or parsed token is invalid.
	// Defaults to implementation using `github.com/golang-jwt/jwt` as JWT implementation library
	ParseTokenFn func(c qore.HttpContext, auth string) (any, error)
	// Claims are extendable claims data defining token content. Used by default ParseTokenFunc implementation.
	// Not used if custom ParseTokenFunc is set.
	// Optional. Defaults to function returning jwt.MapClaims
	NewClaimsFn func(c qore.HttpContext) jwt.Claims
}

// JwtTokenError defines error that return when error occured about JWT token.
type JwtTokenError struct {
	Token *jwt.Token
	Err   error
}

func (e *JwtTokenError) Error() string { return e.Err.Error() }
func (e *JwtTokenError) Unwrap() error { return e.Err }

var defaultJwtConfig = &JwtConfig{
	Skipper: DefaultSkipper,
	ErrorHandler: func(c qore.HttpContext, err error) error {
		return c.Api().ClientError(qore.HttpStatusUnauthorized, err).Response()
	},
	ContextKey:    qore.HTTP_CONTEXT_AUTH,
	SigningMethod: "HS256",
	TokenLookup:   "header:Authorization:Bearer ",
	NewClaimsFn: func(c qore.HttpContext) jwt.Claims {
		return jwt.MapClaims{}
	},
}

func (config *JwtConfig) defaultParseTokenFn(c qore.HttpContext, auth string) (any, error) {
	token, err := jwt.ParseWithClaims(auth, config.NewClaimsFn(c), config.KeyFn)
	if err != nil {
		return nil, &JwtTokenError{Token: token, Err: err}
	}
	if !token.Valid {
		return nil, &JwtTokenError{Token: token, Err: errors.New("invalid token")}
	}
	return token, nil
}

func jwtHandler(next qore.HttpHandler, config *JwtConfig) qore.HttpHandler {
	if config == nil {
		config = defaultJwtConfig
	}
	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultJwtConfig.ErrorHandler
	}
	if len(strings.TrimSpace(config.ContextKey)) == 0 {
		config.ContextKey = defaultJwtConfig.ContextKey
	}
	if len(strings.TrimSpace(config.SigningMethod)) == 0 {
		config.SigningMethod = defaultJwtConfig.SigningMethod
	}
	if len(strings.TrimSpace(config.TokenLookup)) == 0 {
		config.TokenLookup = defaultJwtConfig.TokenLookup
	}
	if config.ParseTokenFn == nil {
		config.ParseTokenFn = config.defaultParseTokenFn
	}
	if config.NewClaimsFn == nil {
		config.NewClaimsFn = defaultJwtConfig.NewClaimsFn
	}

	return func(c qore.HttpContext) error {
		// Skip process middleware.
		if config.Skipper(c) {
			return next(c)
		}

		// Value extractor.
		extractors, err := JwtExtractor(config.TokenLookup)
		if err != nil {
			return err
		}
		var (
			lastExtractorErr error
			lastTokenErr     error
		)
		for _, extractor := range extractors {
			auths, err := extractor(c)
			if err != nil {
				lastExtractorErr = err
				continue
			}
			for _, auth := range auths {
				token, err := config.ParseTokenFn(c, auth)
				if err != nil {
					lastTokenErr = err
					continue
				}

				// Store auth information from token into context.
				c.Set(config.ContextKey, token)
				if config.SuccessHandler != nil {
					config.SuccessHandler(c)
				}
				return next(c)
			}
		}

		// Error checker. Prioritize token errors over extracting value errors.
		if lastTokenErr != nil {
			err = &JwtTokenError{Err: lastTokenErr}
		} else if lastExtractorErr != nil {
			err = &JwtTokenError{Err: lastExtractorErr}
		}

		if lastTokenErr == nil {
			return config.ErrorHandler(c, fmt.Errorf("missing or malformed jwt: %w", err))
		}
		return config.ErrorHandler(c, fmt.Errorf("invalid or expired jwt: %w", err))
	}
}

// JwtWithConfig returns a Jwt middleware with config.
// See: `Jwt()`.
func JwtWithConfig(config *JwtConfig) qore.HttpMiddleware {
	// Return qore.HttpHandler.
	return func(next qore.HttpHandler) qore.HttpHandler {
		return jwtHandler(next, config)
	}
}

// Jwt returns a jwt auth middleware.
func Jwt(next qore.HttpHandler) qore.HttpHandler {
	return JwtWithConfig(defaultJwtConfig)(next)
}
