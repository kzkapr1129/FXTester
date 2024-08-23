package main

import (
	"bytes"
	"fmt"
	"fxtester/internal/common"
	"fxtester/internal/keycloak"
	"fxtester/internal/saml"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/labstack/echo/v4"

	cs "github.com/crewjam/saml"
	cssp "github.com/crewjam/saml/samlsp"
)

const TestServerPort = 8001

type TestSpService struct {
	sp cs.ServiceProvider
}

func NewTestSpService() *TestSpService {
	samlClient := *saml.NewSamlClient(&saml.SamlClientReader{}, nil)
	idpMetadata, err := samlClient.FetchIdpMetadata()
	if err != nil {
		panic(err)
	}

	// バックエンドのURLをurl.URL型に変換する
	backendURL, err := url.Parse(fmt.Sprintf("https://localhost:%d/", TestServerPort))
	if err != nil {
		panic(err)
	}

	opts := cssp.Options{
		EntityID:    backendURL.String(),
		URL:         *backendURL, // acsやsloのURLを作成する際のベースとなるURL
		IDPMetadata: idpMetadata,
		SignRequest: false,
	}

	sp := cssp.DefaultServiceProvider(opts)
	sp.AuthnNameIDFormat = cs.UnspecifiedNameIDFormat
	self := &TestSpService{
		sp: sp,
	}
	self.initKeycloak()
	return self
}

func (t *TestSpService) initKeycloak() {
	param := keycloak.ClientParam{
		keycloak.KeyUser:    common.GetConfig().Saml.Keycloak.AdminUser.Username,
		keycloak.KeyPass:    common.GetConfig().Saml.Keycloak.AdminUser.Password,
		keycloak.KeyBaseURL: strings.TrimRight(common.GetConfig().Saml.Keycloak.BaseURL, "/"),
	}
	c := keycloak.NewClient(param)
	if err := c.Login(); err != nil {
		panic(fmt.Sprintf("ログインに失敗しました: %v", err))
	}

	c.DeleteClient(common.GetConfig().Saml.Keycloak.RealmName, "test-sp")

	var req keycloak.ClientRepresentation
	req.Id = "test-sp"
	req.ClientId = t.sp.EntityID
	req.Protocol = keycloak.ProtocolSAML
	req.RedirectUris = []string{t.sp.EntityID + "*"}
	req.Attributes = map[string]string{
		keycloak.AttributeSamlClientSignature:         "false",
		keycloak.AttributeValidPostLogoutRedirectURIs: t.sp.EntityID + "*",
		keycloak.AttributeLogoutServicePostBindingURL: t.sp.SloURL.String(),
		keycloak.AttributeNameIdFormat:                "email",
	}
	if err := c.CreateClient(common.GetConfig().Saml.Keycloak.RealmName, req); err != nil {
		panic(fmt.Sprintf("クライアントの作成に失敗しました: %v", err))
	}
}

func (b *TestSpService) GetHome(ctx echo.Context) error {
	f, err := os.Open("./cmd/testsp/home.html")
	if err != nil {
		panic(err)
	}
	bytes, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	return ctx.HTML(http.StatusOK, string(bytes))
}

// ユーザをシングルサインオンさせるログインリクエストを作成し、FormのPOSTによってidPに送信するスクリプトタグを含んだHTMLを返却するエンドポイント。
//
// (GET /saml/login)
func (t *TestSpService) GetSamlLogin(ctx echo.Context) error {
	// idpURLを取得
	idpURL := t.sp.GetSSOBindingLocation(cs.HTTPPostBinding)
	// AuthnRequestの作成
	authnRequest, err := t.sp.MakeAuthenticationRequest(idpURL, cs.HTTPPostBinding, cs.HTTPPostBinding)
	if err != nil {
		panic(err)
	}

	// ヘッダの設定
	ctx.Response().Header().Add(echo.HeaderContentSecurityPolicy, "default-src; "+
		"script-src 'sha256-AjPdJSbZmeWHnEc5ykvJFay8FTWeTeRbs9dutfZ0HqE='; "+
		"reflected-xss block; referrer no-referrer;")
	ctx.Response().Header().Add(echo.HeaderContentType, echo.MIMETextHTML)

	const emptyRelayState = ""

	// Bodyの内容組み立て
	var buf bytes.Buffer
	buf.WriteString(`<!DOCTYPE html><html><body>`)
	buf.Write(authnRequest.Post(emptyRelayState))
	buf.WriteString(`</body></html>`)

	return ctx.HTML(http.StatusOK, buf.String())
}

// IdPから受け取る認証レスポンス（SAMLアサーション）を処理するエンドポイント。
//
// (POST /saml/acs)
func (t *TestSpService) PostSamlAcs(ctx echo.Context) error {
	return ctx.Redirect(http.StatusFound, t.sp.EntityID)
}

// ユーザをログアウトさせるログアウトリクエストを作成し、FormのPOSTによってidPに送信するスクリプトタグを含んだHTMLを返却するエンドポイント。
//
// (GET /saml/logout)
func (t *TestSpService) GetSamlLogout(ctx echo.Context) error {
	// idpURLを取得
	idpURL := t.sp.GetSLOBindingLocation(cs.HTTPPostBinding)
	// LogoutRequestの作成
	logoutRequest, err := t.sp.MakeLogoutRequest(idpURL, common.GetConfig().Saml.Keycloak.NewUsers[0].Email)
	if err != nil {
		panic(err)
	}

	// ヘッダの設定
	ctx.Response().Header().Add(echo.HeaderContentSecurityPolicy, "default-src; "+
		"script-src 'sha256-AjPdJSbZmeWHnEc5ykvJFay8FTWeTeRbs9dutfZ0HqE='; "+
		"reflected-xss block; referrer no-referrer;")
	ctx.Response().Header().Add(echo.HeaderContentType, echo.MIMETextHTML)

	const emptyRelayState = ""

	// Bodyの内容組み立て
	var buf bytes.Buffer
	buf.WriteString(`<!DOCTYPE html><html><body>`)
	buf.Write(logoutRequest.Post(emptyRelayState))
	buf.WriteString(`</body></html>`)

	return ctx.HTML(http.StatusOK, buf.String())
}

// IdPから受け取るログアウトリクエストを処理し、ユーザーをログアウトさせるエンドポイント。
//
// (POST /saml/slo)
func (b *TestSpService) PostSamlSlo(ctx echo.Context) error {
	return ctx.Redirect(http.StatusFound, b.sp.EntityID)
}
