package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"fxtester/internal/common"
	"fxtester/internal/db"
	"fxtester/internal/gen"
	"fxtester/internal/lang"
	"fxtester/internal/reader"
	"fxtester/internal/saml"
	"fxtester/internal/validator"
	"fxtester/internal/websock"
	"mime/multipart"
	"net/http"

	"github.com/labstack/echo/v4"
)

type BarService struct {
	samlClient saml.ISamlClient
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

// GetSamlLogin ユーザをシングルサインオンさせるログインリクエストを作成し、FormのPOSTによってidPに送信するスクリプトタグを含んだHTMLを返却するエンドポイント。
//
// (GET /saml/login)
func (b *BarService) GetSamlLogin(ctx echo.Context, params gen.GetSamlLoginParams) error {
	return b.samlClient.ExecuteSamlLogin(ctx, params)
}

// PostSamlAcs IdPから受け取る認証レスポンス（SAMLアサーション）を処理するエンドポイント。
//
// (POST /saml/acs)
func (b *BarService) PostSamlAcs(ctx echo.Context) error {
	return b.samlClient.ExecuteSamlAcs(ctx)
}

// GetSamlError SAMLログインのエラー詳細を返却するエンドポイント
//
// (GET /saml/error)
func (b *BarService) GetSamlError(ctx echo.Context) error {
	return b.samlClient.ExecuteSamlError(ctx)
}

// GetSamlLogout ユーザをログアウトさせるログアウトリクエストを作成し、FormのPOSTによってidPに送信するスクリプトタグを含んだHTMLを返却するエンドポイント。
//
// (GET /saml/logout)
func (b *BarService) GetSamlLogout(ctx echo.Context, params gen.GetSamlLogoutParams) error {
	return b.samlClient.ExecuteSamlLogout(ctx, params)
}

// PostSamlSlo IdPから受け取るログアウトリクエストを処理し、ユーザーをログアウトさせるエンドポイント。
//
// (POST /saml/slo)
func (b *BarService) PostSamlSlo(ctx echo.Context) error {
	return b.samlClient.ExecuteSamlSlo(ctx)
}

// GetWsUuid Websocketと接続します。
//
// (GET /ws/:uuid)
func (b *BarService) GetWsUuid(ctx echo.Context) error {
	return b.websockClient.CommunicateViaWs(ctx)
}

// PostZigzag CSVまたはローソク足のデータをアップロードし、Zigzagのデータを作成します。
//
// (POST /zigzag)
func (b *BarService) PostZigzag(ctx echo.Context) error {
	err := ctx.Request().ParseMultipartForm(1 * 1024 * 1024)
	if err != nil {
		if errors.Is(err, multipart.ErrMessageTooLarge) {
			return lang.NewFxtError(lang.ErrTooLargeMessageError)
		} else {
			return lang.NewFxtError(lang.ErrInvalidRequestProtocol).SetCause(err)
		}
	}

	// リクエストパラメータのバリデーション
	if err := validator.ValidatePostZigzag(ctx); err != nil {
		return err
	}

	form := ctx.Request().MultipartForm
	types := common.NextValue(form.Value["type"])
	csvInfos := common.NextValue(form.Value["csvInfo"])
	candless := common.NextValue(form.Value["candles"])
	csvs := common.NextValue(form.File["csv"])

	for {
		t, err := types()
		if err != nil {
			break
		}

		candles, err := func() ([]gen.Candle, error) {
			var res []gen.Candle
			switch t {
			case string(gen.PostZigzagRequestTypeCsv):
				v, err := csvInfos()
				if err != nil {
					// バリデーション済みのため発生しない想定のエラー
					panic("invalid csvInfo")
				}
				var csvInfo gen.CsvInfo
				if err := json.Unmarshal([]byte(v), &csvInfo); err != nil {
					// バリデーション済みのため発生しない想定のエラー
					panic("invalid csvInfo")
				}

				csvh, err := csvs()
				if err != nil {
					// バリデーション済みのため発生しない想定のエラー
					panic("invalid csvInfo")
				}

				csvf, err := csvh.Open()
				if err != nil {
					return nil, lang.NewFxtError(lang.ErrInvalidParameterError, "csv")
				}
				defer csvf.Close()

				candles, err := reader.ReadCandleCsv(csvInfo, csvf)
				if err != nil {
					return nil, lang.NewFxtError(lang.ErrInvalidParameterError, "csv").SetCause(err)
				}

				fmt.Println(candles)

			case string(gen.PostZigzagRequestTypeCandles):
				v, err := candless()
				if err != nil {
					// バリデーション済みのため発生しない想定のエラー
					panic("invalid candles")
				}
				if err := json.Unmarshal([]byte(v), &res); err != nil {
					// バリデーション済みのため発生しない想定のエラー
					panic("invalid candles")
				}
			default:
				// バリデーション済みのため発生しない想定のエラー
				panic("invalid type " + t)
			}
			return res, nil
		}()
		if err != nil {
			return err
		}

		fmt.Println(candles)
	}

	return ctx.JSON(http.StatusAccepted, gen.PostZigzagResult{})
}
