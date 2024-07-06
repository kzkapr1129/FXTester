package service

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"fxtester/internal/config"
	"fxtester/internal/security"
	"fxtester/middleware"
	"fxtester/openapi/gen"
	"net/http"
	"net/url"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/labstack/echo/v4"
)

type BarService struct {
	serviceProvider saml.ServiceProvider
	requestTracker  samlsp.CookieRequestTracker
}

func NewBarService() (*BarService, error) {
	// idPのメタデータを取得するURLをパースする
	idpMetadataURL, err := url.Parse(config.GetConfig().IdpMetadataURL)
	if err != nil {
		return nil, err
	}
	// idPのメタデータを取得
	idpMetadata, err := samlsp.FetchMetadata(context.Background(), http.DefaultClient, *idpMetadataURL)
	if err != nil {
		return nil, err
	}

	rootURL, err := url.Parse(config.GetConfig().RootUrl)
	if err != nil {
		return nil, err
	}

	keyPair, err := tls.LoadX509KeyPair(config.GetConfig().SamlCertPath, config.GetConfig().SamlKeyPath)
	if err != nil {
		return nil, err
	}
	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		return nil, err
	}

	opts := samlsp.Options{
		EntityID:           config.GetConfig().EntityID,
		URL:                *rootURL,
		IDPMetadata:        idpMetadata,
		Key:                keyPair.PrivateKey.(*rsa.PrivateKey),
		DefaultRedirectURI: config.GetConfig().RedirectUrlAfterLogin,
		Certificate:        keyPair.Leaf,
		SignRequest:        false,
	}
	serviceProvider := samlsp.DefaultServiceProvider(opts)
	serviceProvider.AcsURL = *rootURL.ResolveReference(&url.URL{Path: "/api/saml/acs"})
	requestTracker := samlsp.DefaultRequestTracker(opts, &serviceProvider)

	return &BarService{
		serviceProvider: serviceProvider,
		requestTracker:  requestTracker,
	}, nil
}

// (DELETE /api/auth)
func (s *BarService) DeleteApiAuth(ctx echo.Context) error {
	fmt.Println("DeleteApiAuth: ", ctx.Request().Cookies())
	security.DeleteSession(ctx.Response())
	return nil
}

// (GET /api/auth)
func (s *BarService) GetApiAuth(ctx echo.Context) error {
	return nil
}

// (POST /api/auth)
func (s *BarService) PostApiAuth(ctx echo.Context) error {
	logger := middleware.GetLogger(ctx)

	var reqBody gen.PostApiAuthJSONBody
	decoder := json.NewDecoder(ctx.Request().Body)
	decoder.Decode(&reqBody)

	if reqBody.Id != "admin" || reqBody.Password != "admin" {
		return ctx.JSON(http.StatusNotFound, struct{}{})
	}

	// TODO refreshTokenをDBに格納

	expires, err := security.CreateSession(ctx.Response().Writer, reqBody.Id)
	if err != nil {
		logger.WithError(err).Error("failed to create session")
		return ctx.JSON(http.StatusInternalServerError, struct{}{})
	}

	var res gen.AuthStatus
	res.Status = gen.LoggedIn
	res.Expires = expires
	return ctx.JSON(http.StatusCreated, res)
}

// (GET /api/saml/login)
func (s *BarService) GetApiSamlLogin(ctx echo.Context) error {
	bindingLocation := s.serviceProvider.GetSSOBindingLocation(saml.HTTPPostBinding)

	authReq, err := s.serviceProvider.MakeAuthenticationRequest(bindingLocation, saml.HTTPPostBinding, saml.HTTPPostBinding)
	if err != nil {
		ctx.Logger().Error("failed to MakeAuthenticationRequest")
		return ctx.JSON(http.StatusInternalServerError, struct{}{})
	}

	relayState, err := s.requestTracker.TrackRequest(ctx.Response().Writer, ctx.Request(), authReq.ID)
	if err != nil {
		ctx.Logger().Error("failed to TrackRequest")
		return ctx.JSON(http.StatusInternalServerError, struct{}{})
	}

	ctx.Response().Header().Add("Content-Security-Policy", ""+
		"default-src; "+
		"script-src 'sha256-AjPdJSbZmeWHnEc5ykvJFay8FTWeTeRbs9dutfZ0HqE='; "+
		"reflected-xss block; referrer no-referrer;")
	ctx.Response().Header().Add("Content-type", "text/html")
	var buf bytes.Buffer
	buf.WriteString(`<!DOCTYPE html><html><body>`)
	buf.Write(authReq.Post(relayState))
	buf.WriteString(`</body></html>`)
	if _, err := ctx.Response().Write(buf.Bytes()); err != nil {
		ctx.Logger().Error("failed to TrackRequest")
		return ctx.JSON(http.StatusInternalServerError, struct{}{})
	}

	return nil
}

// (POST /api/auth/refresh)
func (s *BarService) PostApiAuthRefresh(ctx echo.Context) error {
	fmt.Println("PostApiAuthRefresh: ", ctx.Request().Cookies())
	return nil
}

// (GET /api/saml/logout)
func (s *BarService) GetApiSamlLogout(ctx echo.Context) error {
	body, err := s.serviceProvider.MakePostLogoutRequest("admin", "/login")
	if err != nil {
		ctx.Logger().Error("failed to MakeAuthenticationRequest")
		return ctx.JSON(http.StatusInternalServerError, struct{}{})
	}

	ctx.Response().Header().Add("Content-Security-Policy", ""+
		"default-src; "+
		"script-src 'sha256-AjPdJSbZmeWHnEc5ykvJFay8FTWeTeRbs9dutfZ0HqE='; "+
		"reflected-xss block; referrer no-referrer;")
	ctx.Response().Header().Add("Content-type", "text/html")
	var buf bytes.Buffer
	buf.WriteString(`<!DOCTYPE html><html><body>`)
	buf.Write(body)
	buf.WriteString(`</body></html>`)
	if _, err := ctx.Response().Write(buf.Bytes()); err != nil {
		ctx.Logger().Error("failed to TrackRequest")
		return ctx.JSON(http.StatusInternalServerError, struct{}{})
	}

	return nil
}

// (POST /api/saml/acs)
func (s *BarService) PostApiSamlAcs(ctx echo.Context) error {
	err := ctx.Request().ParseForm()
	if err != nil {
		ctx.Logger().Error(err)
		return ctx.JSON(http.StatusBadRequest, struct{}{})
	}

	possibleRequestIDs := []string{}
	if s.serviceProvider.AllowIDPInitiated {
		possibleRequestIDs = append(possibleRequestIDs, "")
	}

	trackedRequests := s.requestTracker.GetTrackedRequests(ctx.Request())
	for _, tr := range trackedRequests {
		possibleRequestIDs = append(possibleRequestIDs, tr.SAMLRequestID)
	}

	assertion, err := s.serviceProvider.ParseResponse(ctx.Request(), possibleRequestIDs)
	if err != nil {
		ctx.Logger().Error(err)
		return ctx.JSON(http.StatusBadRequest, struct{}{})
	}

	fmt.Println("Issuer: ", assertion.Issuer.Value)
	for _, attributeStatement := range assertion.AttributeStatements {
		for _, attribute := range attributeStatement.Attributes {
			if attribute.FriendlyName == "username" {
				for _, value := range attribute.Values {
					fmt.Println("username is ", value.Value)
				}
			}
		}
	}

	redirectURI := s.serviceProvider.DefaultRedirectURI
	if trackedRequestIndex := ctx.Request().Form.Get("RelayState"); trackedRequestIndex != "" {
		trackedRequest, err := s.requestTracker.GetTrackedRequest(ctx.Request(), trackedRequestIndex)
		if err != nil {
			if err == http.ErrNoCookie && s.serviceProvider.AllowIDPInitiated {
				if uri := ctx.Request().Form.Get("RelayState"); uri != "" {
					redirectURI = uri
				}
			} else {
				ctx.Logger().Fatal("failed1 !!!!", err)
			}
		} else {
			if err := s.requestTracker.StopTrackingRequest(ctx.Response().Writer, ctx.Request(), trackedRequestIndex); err != nil {
				ctx.Logger().Fatal("failed2 !!!!", err)
			}

			redirectURI = trackedRequest.URI
		}
	}

	if url, err := url.Parse(redirectURI); err != nil {
		ctx.Logger().Fatal("failed3 !!!!", err)
	} else if _, exists := config.GetConfig().BlacklistRedirectUrls[url.Path]; exists {
		// redirectのループを回避するための処理
		redirectURI = s.serviceProvider.DefaultRedirectURI
	}

	return ctx.Redirect(http.StatusFound, redirectURI)
}

// (POST /api/saml/slo)
func (s *BarService) PostApiSamlSlo(ctx echo.Context) error {
	return ctx.Redirect(http.StatusFound, config.GetConfig().RedirectUrlAfterLogout)
}

// (DELETE /api/data/features)
func (s *BarService) DeleteApiDataFeatures(ctx echo.Context) error {
	logger := middleware.GetLogger(ctx)
	logger.Println("called DeleteDataFeatures", config.GetConfig())

	var reqBody gen.DeleteApiDataFeaturesJSONBody
	decoder := json.NewDecoder(ctx.Request().Body)
	decoder.Decode(&reqBody)
	fmt.Println(reqBody)

	return ctx.JSON(http.StatusBadRequest, gen.Error{
		Code:      200,
		ErrorName: "testError",
		Arguments: &[]string{"hoge"},
	})
}

// (GET /api/data/features)
func (*BarService) GetApiDataFeatures(ctx echo.Context, params gen.GetApiDataFeaturesParams) error {
	return nil
}

// (POST /api/data/features)
func (*BarService) PostApiDataFeatures(ctx echo.Context) error {
	return nil
}

// (GET /api/data/resource/features/ids)
func (*BarService) GetApiDataResourceFeaturesIds(ctx echo.Context, params gen.GetApiDataResourceFeaturesIdsParams) error {
	return nil
}

// (DELETE /api/resource/candles)
func (*BarService) DeleteApiResourceCandles(ctx echo.Context) error {
	return nil
}

// (GET /api/resource/candles)
func (*BarService) GetApiResourceCandles(ctx echo.Context, params gen.GetApiResourceCandlesParams) error {
	return nil
}

// (POST /api/resource/candles)
func (*BarService) PostApiResourceCandles(ctx echo.Context) error {
	return nil
}

// (GET /api/resource/candles/metadata)
func (*BarService) GetApiResourceCandlesMetadata(ctx echo.Context, params gen.GetApiResourceCandlesMetadataParams) error {
	return nil
}

// (DELETE /api/test)
func (*BarService) DeleteApiTest(ctx echo.Context) error {
	return nil
}

// (GET /api/test)
func (*BarService) GetApiTest(ctx echo.Context, params gen.GetApiTestParams) error {
	return nil
}

// (POST /api/test)
func (*BarService) PostApiTest(ctx echo.Context) error {
	return nil
}

// (GET /api/test/ids)
func (*BarService) GetApiTestIds(ctx echo.Context, params gen.GetApiTestIdsParams) error {
	return nil
}
