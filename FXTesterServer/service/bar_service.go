package service

import (
	"fxtester/internal/db"
	"fxtester/internal/gen"
	"fxtester/internal/saml"

	"github.com/labstack/echo/v4"
)

type BarService struct {
	samlClient *saml.SamlClient
	dbWrapper  db.IDbWrapper
}

func NewBarService() *BarService {
	dbWrapper := &db.DbWrapper{}
	client := saml.NewSamlClient(&saml.SamlClientReader{}, dbWrapper)
	return &BarService{
		samlClient: client,
		dbWrapper:  dbWrapper,
	}
}

func (b *BarService) Init() error {
	// SAMLクライアントの初期化
	if err := b.samlClient.Init(); err != nil {
		return err
	}
	// DBハンドルプロバイダーの初期化
	if err := b.dbWrapper.Init(); err != nil {
		return err
	}
	return nil
}

// ユーザをシングルサインオンさせるログインリクエストを作成し、FormのPOSTによってidPに送信するスクリプトタグを含んだHTMLを返却するエンドポイント。
//
// (GET /saml/login)
func (b *BarService) GetSamlLogin(ctx echo.Context, params gen.GetSamlLoginParams) error {
	return b.samlClient.ExecuteSamlLogin(ctx, params)
}

// IdPから受け取る認証レスポンス（SAMLアサーション）を処理するエンドポイント。
//
// (POST /saml/acs)
func (b *BarService) PostSamlAcs(ctx echo.Context) error {
	return b.samlClient.ExecuteSamlAcs(ctx)
}

// SAMLログインのエラー詳細を返却するエンドポイント
//
// (GET /saml/error)
func (b *BarService) GetSamlError(ctx echo.Context) error {
	return b.samlClient.ExecuteSamlError(ctx)
}

// ユーザをログアウトさせるログアウトリクエストを作成し、FormのPOSTによってidPに送信するスクリプトタグを含んだHTMLを返却するエンドポイント。
//
// (GET /saml/logout)
func (b *BarService) GetSamlLogout(ctx echo.Context, params gen.GetSamlLogoutParams) error {
	return b.samlClient.ExecuteSamlLogout(ctx, params)
}

// IdPから受け取るログアウトリクエストを処理し、ユーザーをログアウトさせるエンドポイント。
//
// (POST /saml/slo)
func (b *BarService) PostSamlSlo(ctx echo.Context) error {
	return b.samlClient.ExecuteSamlSlo(ctx)
}
