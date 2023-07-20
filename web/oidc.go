package web

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

const (
	accessToken    = "access_token"
	refreshToken   = "refresh_token"
	idToken        = "id_token"
	ocidConfigPath = "/.well-known/openid-configuration"
)

func init() {
	Flags = append(Flags, &cli.StringFlag{
		Name:    "oidc.client",
		Usage:   "clientId of oauth",
		EnvVars: []string{"OIDC_CLIENT"},
	},
		&cli.StringFlag{
			Name:    "oidc.secret",
			Usage:   "Client Secret of oauth",
			EnvVars: []string{"OIDC_SECRET"},
		},
		&cli.StringFlag{
			Name:    "oidc.site",
			Usage:   "site of oauth server",
			EnvVars: []string{"OIDC_SITE"},
		},
	)
}

type oidcAuth struct {
	config    *oauth2.Config
	site      string
	keySet    jwk.Set
	logoutUrl *url.URL
}

func initOIDC(ctx *cli.Context) {
	site := ctx.String("oidc.site")
	clientId := ctx.String("oidc.client")
	clientSecret := ctx.String("oidc.secret")
	if site == "" || clientId == "" {
		return
	}
	configUrl := site + ocidConfigPath
	oidc, err := newOIDC(configUrl)
	if err != nil {
		log.Printf("init oauth failed")
		return
	}
	auth = oidc
	oidc.config.ClientID = clientId
	oidc.config.ClientSecret = clientSecret
	oidc.config.Scopes = []string{"openid"}
	engine.GET("/callback", oidc.callback)
}
func newOIDC(configUrl string) (*oidcAuth, error) {
	res, err := http.Get(configUrl)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(res.Body)
	content, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	configs := make(map[string]interface{})
	err = json.Unmarshal(content, &configs)
	if err != nil {
		return nil, err
	}
	authUrl, ok := configs["authorization_endpoint"].(string)
	if !ok {
		return nil, fmt.Errorf("get oauth info from %s failed", configUrl)
	}
	tokenUrl, _ := configs["token_endpoint"].(string)
	config := &oauth2.Config{
		Endpoint: oauth2.Endpoint{
			AuthURL:  authUrl,
			TokenURL: tokenUrl,
		},
	}
	jwksURL, _ := configs["jwks_uri"].(string)
	keySet, err := jwk.Fetch(context.Background(), jwksURL)
	if err != nil {
		return nil, err
	}
	endSessionEndpoint, _ := configs["end_session_endpoint"].(string)
	logoutUrl, err := url.Parse(endSessionEndpoint)
	if err != nil {
		return nil, err
	}
	return &oidcAuth{
		config:    config,
		keySet:    keySet,
		logoutUrl: logoutUrl,
	}, nil
}
func (oidc *oidcAuth) isLogin(ctx *gin.Context) bool {
	session := sessions.Default(ctx)
	token := session.Get(accessToken)
	rawToken, ok := token.(string)
	if !ok {
		return false
	}
	_, err := oidc.parseAndValidateToken(rawToken)
	if ve, ok := err.(*jwt.ValidationError); ok && ve.Errors&jwt.ValidationErrorExpired != 0 {
		err = oidc.refresh(ctx)
	}
	return err == nil
}

func (oidc *oidcAuth) redirectToLogin(ctx *gin.Context) {
	session := sessions.Default(ctx)
	state := getState(6)
	session.Set("state", state)
	err := session.Save()
	if err != nil {
		responseError(ctx, err)
		return
	}
	oidc.config.RedirectURL = getRedirectUrl(ctx.Request)
	authUrl := oidc.config.AuthCodeURL(state)
	redirect(ctx, authUrl)
}
func getRedirectUrl(req *http.Request) string {
	host := req.Host
	proto := req.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		proto = "http"
	}
	return fmt.Sprintf("%s://%s/callback", proto, host)
}
func (oidc *oidcAuth) parseAndValidateToken(rawToken string) (*jwt.Token, error) {

	token, err := jwt.Parse(rawToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("no kid found in token")
		}

		keys, ok := oidc.keySet.LookupKeyID(kid)
		if !ok {
			return nil, fmt.Errorf("no key found with kid: %s", kid)
		}
		cert := keys.X509CertChain()
		if len(cert) < 1 {
			return nil, fmt.Errorf("no key found with kid: %s", kid)
		}
		return cert[0].PublicKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}
	return token, nil
}
func (oidc *oidcAuth) refresh(c *gin.Context) error {
	session := sessions.Default(c)
	token, ok := session.Get(refreshToken).(string)
	if !ok {
		return fmt.Errorf("no refresh token found")
	}

	tokenSource := oidc.config.TokenSource(context.Background(), &oauth2.Token{
		RefreshToken: token,
	})

	newToken, err := tokenSource.Token()
	if err != nil {
		return err
	}
	session.Set(refreshToken, newToken.RefreshToken)
	session.Set(accessToken, newToken.AccessToken)
	session.Set(idToken, newToken.Extra(idToken))
	return session.Save()
}

func (oidc *oidcAuth) callback(ctx *gin.Context) {
	session := sessions.Default(ctx)
	state := ctx.Query("state")
	sessionState := session.Get("state")
	if state != sessionState {
		oidc.redirectToLogin(ctx)
		return
	}
	code := ctx.Query("code")
	token, err := oidc.config.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("oauthConfig.Exchange() failed with '%s'\n", err)
		oidc.redirectToLogin(ctx)
		return
	}
	session.Set(accessToken, token.AccessToken)
	session.Set(refreshToken, token.RefreshToken)
	session.Set(idToken, token.Extra(idToken))
	err = session.Save()
	if err != nil {
		responseError(ctx, err)
		return
	}
	ctx.Redirect(http.StatusFound, "/")
}

func getState(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func (oidc *oidcAuth) logout(ctx *gin.Context) {
	session := sessions.Default(ctx)
	id := session.Get(idToken)
	session.Delete(accessToken)
	session.Delete(refreshToken)
	session.Delete(idToken)
	session.Delete("state")
	session.Save()
	if idToken, ok := id.(string); ok {
		query := oidc.logoutUrl.Query()
		query.Add("id_token_hint", idToken)
		logoutUrl := *oidc.logoutUrl
		logoutUrl.RawQuery = query.Encode()
		res, err := http.Get(logoutUrl.String())
		if err != nil {
			log.Printf("logout from oauth server failed:%v", err)
		} else if err = res.Body.Close(); err != nil {
			log.Printf("close response body failed:%v", err)
		}
	}
	ctx.Redirect(http.StatusFound, "/")
}
