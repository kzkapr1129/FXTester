package service

import (
	"fxtester/internal/db"
	"fxtester/internal/gen"
	"fxtester/internal/saml"
	"fxtester/internal/websock"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type BarService struct {
	samlClient *saml.SamlClient
	dbWrapper  db.IDbWrapper

	websockClient *websock.WebsockClient
}

func NewBarService() *BarService {
	dbWrapper := &db.DbWrapper{}
	samlClient := saml.NewSamlClient(&saml.SamlClientReader{}, dbWrapper)
	websockClient := websock.NewWebsockClient()

	return &BarService{
		samlClient:    samlClient,
		dbWrapper:     dbWrapper,
		websockClient: websockClient,
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

// Websocketと接続します。
//
// (GET /ws/:uuid)
func (b *BarService) GetWsUuid(ctx echo.Context) error {
	return b.websockClient.CommunicateViaWs(ctx)
}

// ローソク足のデータが格納されたファイル(MT4が出力したCSVなど)をアップロードし、特徴点の抽出を開始します。
// (POST /feature_points)
func (b *BarService) PostFeaturePoints(ctx echo.Context) error {
	writer, closer, uuid, err := b.websockClient.NewWs()
	if err != nil {
		return err
	}

	go func() {
		defer closer()

		const max = 100
		for i := 0; i < max; i++ {
			writer("progress", float32(i)/max)
			time.Sleep(100 * time.Millisecond)
		}

		// 処理結果をDB(またはfireStore)に格納して、wsの切断等に対処する
	}()

	return ctx.JSON(http.StatusAccepted, gen.CreateFeaturePointsResult{
		Uuid: uuid,
	})
}
