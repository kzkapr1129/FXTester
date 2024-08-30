package saml

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/xml"
	"fxtester/internal/common"
	"fxtester/internal/db"
	"fxtester/internal/gen"
	"fxtester/internal/lang"
	"fxtester/internal/net"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	cs "github.com/crewjam/saml"
	cssp "github.com/crewjam/saml/samlsp"
	"github.com/labstack/echo/v4"
)

const SchemeFile = "file://"
const SchemeHttps = "https://"
const SchemeHttp = "http://"

const SAMLRequest = "SAMLRequest"
const SAMLResponse = "SAMLResponse"

const URLParamSamlError = "saml_error"

type ISamlClientReader interface {
	OpenFile(path string) (io.ReadCloser, error)
	FetchMetadata(ctx context.Context, url url.URL, timeout time.Duration) (*cs.EntityDescriptor, error)
}

type SamlClientReader struct {
}

func (c *SamlClientReader) OpenFile(path string) (io.ReadCloser, error) {
	return os.Open(path)
}
func (c *SamlClientReader) FetchMetadata(ctx context.Context, url url.URL, timeout time.Duration) (*cs.EntityDescriptor, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return cssp.FetchMetadata(ctx, http.DefaultClient, url)
}

type SamlClient struct {
	reader ISamlClientReader
	sp     cs.ServiceProvider
	dao    db.IUserEntityDao
}

// NewSamlClient SAMLクライアントを生成します
func NewSamlClient(reader ISamlClientReader, dbProvider db.IDbWrapper) *SamlClient {
	return &SamlClient{
		reader: reader,
		dao:    db.NewUserEntityDao(dbProvider),
	}
}

// Init SAMLクライアントを初期化します
func (c *SamlClient) Init() error {
	// バックエンドのURLをurl.URL型に変換する
	backendURL, err := url.Parse(common.GetConfig().Saml.BackendURL)
	if err != nil {
		return lang.NewFxtError(lang.ErrCodeConfig).SetCause(err)
	}

	// idPのmetadata.xmlをフェッチする
	idpMetadata, err := c.FetchIdpMetadata()
	if err != nil {
		return err
	}

	opts := cssp.Options{
		EntityID:    common.GetConfig().Saml.EntityId,
		URL:         *backendURL, // acsやsloのURLを作成する際のベースとなるURL
		IDPMetadata: idpMetadata,
		SignRequest: false,
	}
	sp := cssp.DefaultServiceProvider(opts)
	sp.AuthnNameIDFormat = cs.UnspecifiedNameIDFormat
	c.sp = sp
	return nil
}

// ExecuteSamlLogin SAMLのSSOを開始します
func (c *SamlClient) ExecuteSamlLogin(ctx echo.Context, params gen.GetSamlLoginParams) error {
	// idpURLを取得
	idpURL := c.sp.GetSSOBindingLocation(cs.HTTPPostBinding)
	// AuthnRequestの作成
	authnRequest, err := c.sp.MakeAuthenticationRequest(idpURL, cs.HTTPPostBinding, cs.HTTPPostBinding)
	if err != nil {
		return lang.NewFxtError(lang.ErrSSOAuthnRequest)
	}

	// SSOのセッションを作成
	// (クッキーの設定はヘッダ書き込みとBody書き込みより先に行う実装制約があるため)
	if err := net.CreateSSOSession(ctx.Response().Writer, authnRequest.ID, params.XRedirectURL, params.XRedirectURLOnError); err != nil {
		return err
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

	// text/htmlとしてレスポンスを返却する
	if err = ctx.HTML(http.StatusOK, buf.String()); err != nil {
		return lang.NewFxtError(lang.ErrSSOHtmlWriting).SetCause(err)
	}

	return nil
}

func (s *SamlClient) ExecuteSamlAcs(ctx echo.Context) (lastError error) {
	// リクエストのFormを解析する (のちの処理で'ctx.Request().PostForm'を参照できるようにするため)
	err := ctx.Request().ParseForm()
	if err != nil {
		return lang.NewFxtError(lang.ErrRequestParse).SetCause(err)
	}

	// SSOのセッション情報を取得
	session, err := net.GetSSOSession(ctx.Request())
	if err != nil {
		return err
	}

	// エラーの有無によってリダイレクト先やパラメータを変更するdefer
	defer func() {
		// SSOのセッションを破棄する
		net.DeleteSSOSession(ctx.Response().Writer)

		if lastError != nil {
			ctx.Logger().Error(lastError)

			// エラーレスポンスを作成
			_, res := lang.ConvertToGenError(ctx, lastError)
			// エラーの詳細をクッキーに保存
			net.CreateSamlErrorSession(ctx.Response().Writer, *res)
			// リダイレクト先にURLパラメータでエラー内容を通知
			params := url.Values{}
			params.Add(URLParamSamlError, "1")
			// 古いエラーは握り潰しリダイレクトに失敗した場合のみエラー処理を行う
			lastError = ctx.Redirect(http.StatusFound, session.RedirectURLOnError+"?"+params.Encode())
			return
		} else {
			// エラーを空にする
			net.CreateSamlErrorSession(ctx.Response().Writer, gen.Error{})
			// エラーなしの場合
			lastError = ctx.Redirect(http.StatusFound, session.RedirectURL)
			return
		}
	}()

	// SAMLResponseを解析する
	possibleRequestIds := []string{session.AuthnRequestId}
	assertion, err := s.sp.ParseResponse(ctx.Request(), possibleRequestIds)
	if err != nil {
		// SAMLResponseの解析に失敗した場合 (metadata.xmlとの鍵の不一致等)
		return lang.NewFxtError(lang.ErrCodeSSOParseResponse).SetCause(err)
	}

	// SAMLアサーションが取得できているかチェック
	if assertion == nil || assertion.Subject == nil || assertion.Subject.NameID == nil || assertion.Subject.NameID.Value == "" {
		// SAMLアサーションが取得できなかった場合
		return lang.NewFxtError(lang.ErrUnexpectedAssertion).SetCause(err)
	}
	// アサーションのNameIdとemailとする (keycloakの設定が正しければemailになっている)
	email := assertion.Subject.NameID.Value

	return func() (lastError error) {
		// トランザクション開始
		if err := s.dao.Begin(); err != nil {
			return err
		}

		defer func() {
			// エラーの有無に応じてRollbackまたはCommitを実行する
			if lastError != nil {
				err := s.dao.Rollback()
				if err != nil {
					ctx.Logger().Error(err)
				}
			} else {
				err := s.dao.Commit()
				if err != nil {
					lastError = err
				}
			}
		}()

		// ユーザが存在するか確認する
		entity, err := s.dao.SelectWithEmail(email)
		if err != nil {
			// ユーザが存在しない場合はユーザを作成する
			entity, err = s.dao.CreateUser(email)
			if err != nil {
				return err
			}
		}

		// 認証セッションを作成する
		return net.CreateAuthSession(ctx.Response().Writer, entity.UserId, entity.Email, func(accessToken, refreshToken string) error {
			// トークンをDBに保存
			return s.dao.UpdateToken(entity.UserId, accessToken, refreshToken)
		})
	}()
}

func (c *SamlClient) ExecuteSamlLogout(ctx echo.Context, params gen.GetSamlLogoutParams) error {
	// アクセストークンの検証
	session, err := net.GetAuthSessionAccessToken(ctx.Request())
	if err != nil {
		return err
	}

	// idpURLを取得
	idpURL := c.sp.GetSLOBindingLocation(cs.HTTPPostBinding)
	// LogoutRequestの作成
	logoutRequest, err := c.sp.MakeLogoutRequest(idpURL, session.Email)
	if err != nil {
		return lang.NewFxtError(lang.ErrSLOAuthnRequest)
	}

	// SLOセッションの作成
	err = net.CreateSLOSession(ctx.Response().Writer, session.UserId, logoutRequest.ID, params.XRedirectURL, params.XRedirectURLOnError)
	if err != nil {
		return err
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

	// text/htmlとしてレスポンスを返却する
	if err = ctx.HTML(http.StatusOK, buf.String()); err != nil {
		return lang.NewFxtError(lang.ErrSSOHtmlWriting).SetCause(err)
	}

	return nil
}

func (c *SamlClient) ExecuteSamlSlo(ctx echo.Context) error {
	// リクエストのFormを解析する (のちの処理で'ctx.Request().PostForm'を参照できるようにするため)
	err := ctx.Request().ParseForm()
	if err != nil {
		return lang.NewFxtError(lang.ErrRequestParse).SetCause(err)
	}

	form := ctx.Request().Form

	if form.Has(SAMLRequest) {
		// 他SP起点のシングルログアウトの場合
		return c.executeSamlSloByOther(ctx)
	} else if form.Has(SAMLResponse) {
		// 本SP起点のシングルログアウトの場合
		return c.executeSamlSloByMySp(ctx)
	}

	// 許可されていない操作エラー
	return lang.NewFxtError(lang.ErrOperationNotAllow)
}

// executeSamlSloByOther 他SP起点のシングルサインアウトを処理する
func (c *SamlClient) executeSamlSloByOther(ctx echo.Context) (lastError error) {
	if err := c.dao.Begin(); err != nil {
		return err
	}
	defer func() {
		if lastError != nil {
			if err := c.dao.Rollback(); err != nil {
				ctx.Logger().Warnf("Error in rollback: %v", err)
			}
		} else {
			if err := c.dao.Commit(); err != nil {
				ctx.Logger().Warnf("Error in commit: %v", err)
			}
		}
	}()

	// リクエストボディからSAMLRequestを取り出す
	samlRequestEncoded := ctx.Request().Form.Get(SAMLRequest)
	// base64でエンコードされているSAMLRequestをデコードする
	samlRequestXML, err := base64.StdEncoding.DecodeString(samlRequestEncoded)
	if err != nil {
		return lang.NewFxtError(lang.ErrBase64SamlRequest).SetCause(err)
	}

	var logoutRequest cs.LogoutRequest
	// xml形式のSAMLRequestをLogoutRequestオブジェクトにUnmarshalする
	if err := xml.Unmarshal(samlRequestXML, &logoutRequest); err != nil {
		return lang.NewFxtError(lang.ErrUnmarshalSamlRequest).SetCause(err)
	}

	// LogoutRequestからnameIDを取り出す
	nameId, err := func() (string, error) {
		if logoutRequest.NameID == nil || logoutRequest.NameID.Value == "" {
			// nameIDが格納されていない場合
			return "", lang.NewFxtError(lang.ErrEmptyNameId)
		}
		return logoutRequest.NameID.Value, nil
	}()
	if err != nil {
		return err
	}

	// アクセストークンの検証
	session, err := net.GetAuthSessionAccessToken(ctx.Request())
	if err == nil {
		// サインアウトするユーザ情報が一致しているか確認
		if session.Email != nameId {
			ctx.Logger().Warnf("cannot delete the auth session: mismatch nameId %s vs %s", session.Email, nameId)
		} else {
			// Authセッションの削除
			net.DeleteAuthSession(ctx.Response().Writer)
			// DBのクリア
			if user, err := c.dao.SelectWithEmail(session.Email); err != nil {
				ctx.Logger().Warnf("cannot select user: email=%s", session.Email)
			} else if err = c.dao.UpdateToken(user.UserId, "", ""); err != nil {
				ctx.Logger().Warnf("cannot update token: userId=%d", user.UserId)
			}
		}
	} else {
		ctx.Logger().Warn("cannot delete the auth session: session nothing")
	}

	// idPのURLを取得
	idpURL := c.sp.GetSLOBindingLocation(cs.HTTPPostBinding)
	// LogoutResponseを作成する
	logoutResponse, err := c.sp.MakeLogoutResponse(idpURL, logoutRequest.ID)
	if err != nil {
		return lang.NewFxtError(lang.ErrSamlLogoutResponseCreation).SetCause(err)
	}

	// ヘッダの設定
	ctx.Response().Header().Add(echo.HeaderContentSecurityPolicy, "default-src; "+
		"script-src 'sha256-ae3F9sw3MnGNUqmT+7gdyojm/I6ukOUOr9mHRkJJvCU='; "+
		"reflected-xss block; referrer no-referrer;")
	ctx.Response().Header().Add(echo.HeaderContentType, echo.MIMETextHTML)

	const emptyRelayState = ""

	// html作成
	var buf bytes.Buffer
	buf.WriteString(`<!DOCTYPE html><html><body>`)
	buf.Write(logoutResponse.Post(emptyRelayState))
	buf.WriteString(`</body></html>`)

	return ctx.HTML(http.StatusOK, buf.String())
}

// executeSamlSloByMySp 自SP起点のシングルサインアウトを処理する
func (c *SamlClient) executeSamlSloByMySp(ctx echo.Context) (lastError error) {
	// SLOセッションの取得
	session, err := net.GetSLOSession(ctx.Request())
	if err != nil {
		return err
	}

	if err := c.dao.Begin(); err != nil {
		return err
	}

	defer func() {
		if lastError != nil {
			if err := c.dao.Rollback(); err != nil {
				ctx.Logger().Warn("error in rollback: %v", err)
			}

			ctx.Logger().Error(lastError)

			// executeSamlSloByMySp()がエラーを返却した場合

			// エラーレスポンスを作成
			_, res := lang.ConvertToGenError(ctx, lastError)
			// エラーの詳細をクッキーに保存
			net.CreateSamlErrorSession(ctx.Response().Writer, *res)
			// リダイレクト先にURLパラメータでエラー内容を通知
			params := url.Values{}
			params.Add(URLParamSamlError, "1")
			// リダイレクト要求を発効
			lastError = ctx.Redirect(http.StatusFound, session.RedirectURLOnError+"?"+params.Encode())
		} else {
			if err := c.dao.Commit(); err != nil {
				ctx.Logger().Warn("error in commit: %v", err)
			}

			// executeSamlSloByMySp()がエラーを返却しなかった場合
			// エラーを空にする
			net.CreateSamlErrorSession(ctx.Response().Writer, gen.Error{})
			// リダイレクト要求を発効
			lastError = ctx.Redirect(http.StatusFound, session.RedirectURL)
		}
	}()

	// リクエストボディからSAMLResponseを取り出す
	samlResponseEncoded := ctx.Request().Form.Get(SAMLResponse)
	// base64でエンコードされているSAMLResponseをデコードする
	samlResponseXML, err := base64.StdEncoding.DecodeString(samlResponseEncoded)
	if err != nil {
		return lang.NewFxtError(lang.ErrBase64SamlResponse).SetCause(err)
	}

	var logoutResponse cs.LogoutResponse
	// xml形式のSAMLRequestをLogoutRequestオブジェクトにUnmarshalする
	if err := xml.Unmarshal(samlResponseXML, &logoutResponse); err != nil {
		return lang.NewFxtError(lang.ErrUnmarshalSamlResponse).SetCause(err)
	}

	// LogoutResponseからInResponseTo(AuthnRequest.ID)を取り出す
	logoutRequestId, err := func() (string, error) {
		if logoutResponse.InResponseTo == "" {
			// AuthnRequestIDが格納されていない場合
			return "", lang.NewFxtError(lang.ErrEmptyLogoutRequestId)
		}
		return logoutResponse.InResponseTo, nil
	}()
	if err != nil {
		return err
	}

	// IDが一致しているか確認
	if session.AuthnRequestId != logoutRequestId {
		// IDが一致していない場合
		return lang.NewFxtError(lang.ErrInvalidNameId)
	}

	// LogoutResponseの検証
	if err := c.sp.ValidateLogoutResponseRequest(ctx.Request()); err != nil {
		return lang.NewFxtError(lang.ErrSLOValidation).SetCause(err)
	}

	// Authセッションの削除
	net.DeleteAuthSession(ctx.Response().Writer)
	// SLOセッションの削除
	net.DeleteSLOSession(ctx.Response().Writer)
	// DBのクリア
	if err := c.dao.UpdateToken(session.UserId, "", ""); err != nil {
		return err
	}

	return nil
}

func (c *SamlClient) ExecuteSamlError(ctx echo.Context) error {
	session, err := net.GetSamlErrorSession(ctx.Request())
	if err != nil {
		return err
	}
	// データを一度取得したらセッションを削除
	net.DeleteSamlErrorSession(ctx.Response())
	return ctx.JSON(http.StatusOK, session)
}

func (c *SamlClient) FetchIdpMetadata() (*cs.EntityDescriptor, error) {
	idpMetadataUrl := common.GetConfig().Saml.IdpMetadataUrl
	if strings.HasPrefix(idpMetadataUrl, SchemeFile) {
		// ファイル読み込みの場合
		return c.fetchIdpMetadataFromFile(idpMetadataUrl)
	} else if strings.HasPrefix(idpMetadataUrl, SchemeHttps) || strings.HasPrefix(idpMetadataUrl, SchemeHttp) {
		// ネットワークダウンロードの場合 (https or http)
		return c.fetchIdpMetadataFromNetwork(idpMetadataUrl)
	}
	return nil, lang.NewFxtError(lang.ErrCodeConfig)
}

func (c *SamlClient) fetchIdpMetadataFromFile(idpMetadataUrl string) (*cs.EntityDescriptor, error) {
	// スキーム(file://)を削除してファイルパスを抽出する
	path := strings.TrimPrefix(idpMetadataUrl, SchemeFile)
	f, err := c.reader.OpenFile(path)
	if err != nil {
		return nil, lang.NewFxtError(lang.ErrCodeDisk).SetCause(err)
	}
	defer f.Close()

	// ファイルを読み込む
	bytes, err := io.ReadAll(f)
	if err != nil {
		return nil, lang.NewFxtError(lang.ErrCodeDisk).SetCause(err)
	}

	// メタデータを解析する
	descriptor, err := cssp.ParseMetadata(bytes)
	if err != nil {
		return nil, lang.NewFxtError(lang.ErrInvalidIdpMetadata).SetCause(err)
	}

	return descriptor, nil
}

func (c *SamlClient) fetchIdpMetadataFromNetwork(idpMetadataUrl string) (*cs.EntityDescriptor, error) {
	u, err := url.Parse(idpMetadataUrl)
	if err != nil {
		return nil, lang.NewFxtError(lang.ErrCodeConfig).SetCause(err)
	}

	var descriptor *cs.EntityDescriptor

	baseCtx := context.Background()
	for retry := 1; retry <= 2; retry++ {
		timeout := time.Duration(5*retry) * time.Second
		d, err := c.reader.FetchMetadata(baseCtx, *u, timeout)
		if err == nil {
			descriptor = d
			break
		}
		time.Sleep(timeout)
	}
	if descriptor == nil {
		return nil, lang.NewFxtError(lang.ErrDownloadIdpMetadata)
	}
	return descriptor, nil
}
