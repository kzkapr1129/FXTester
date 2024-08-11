package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"fxtester/internal"
	"fxtester/internal/gen"
	"fxtester/internal/middleware"
	"net/http"
	"time"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/labstack/echo/v4"
)

type BarService struct {
	serviceProvider saml.ServiceProvider
	requestTracker  samlsp.CookieRequestTracker
}

func NewBarService() (*BarService, error) {
	return &BarService{}, nil
}

// (DELETE /api/auth)
func (s *BarService) DeleteApiAuth(ctx echo.Context) error {
	fmt.Println("DeleteApiAuth: ", ctx.Request().Cookies())
	internal.DeleteSession(ctx.Response())
	return nil
}

// (GET /api/auth)
func (s *BarService) GetApiAuth(ctx echo.Context) error {
	accessToken, err := internal.GetAccessToken(*ctx.Request())
	if err != nil {
		return ctx.JSON(http.StatusNotFound, struct{}{})
	}
	claims, err := internal.VerifyAccessToken(accessToken)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, struct{}{})
	}
	expiresAt := time.Unix(claims.ExpiresAt, 0)
	return ctx.JSON(http.StatusOK, gen.AuthStatus{
		Expires: expiresAt.UTC().Format(time.RFC3339),
	})
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

	expires, err := internal.CreateSession(ctx.Response().Writer, reqBody.Id)
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
	return nil
}

// (POST /api/saml/slo)
func (s *BarService) PostApiSamlSlo(ctx echo.Context) error {
	return nil
}

// (DELETE /api/data/features)
func (s *BarService) DeleteApiDataFeatures(ctx echo.Context) error {
	logger := middleware.GetLogger(ctx)
	logger.Println("called DeleteDataFeatures", internal.GetConfig())

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
