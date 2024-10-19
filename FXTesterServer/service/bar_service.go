package service

import (
	"encoding/json"
	"errors"
	"fxtester/internal/algo"
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
	"sort"
	"time"

	"github.com/labstack/echo/v4"
)

type BarService struct {
	samlClient saml.ISamlClient
	idb        db.IDB

	websockClient *websock.WebsockClient
}

func NewBarService() *BarService {
	db := &db.DB{}
	samlClient := saml.NewSamlClient(&saml.Delegator{}, db)
	websockClient := websock.NewWebsockClient()

	return &BarService{
		samlClient:    samlClient,
		idb:           db,
		websockClient: websockClient,
	}
}

func (b *BarService) Init() error {
	// SAMLクライアントの初期化
	if err := b.samlClient.Init(); err != nil {
		return err
	}
	// DBハンドルプロバイダーの初期化
	if err := b.idb.Init(); err != nil {
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
	types := form.Value["type"]
	csvInfos := form.Value["csvInfo"]
	candless := form.Value["candles"]
	csvs := form.File["csv"]

	paramCandles, err := func() ([]common.Candle, error) {
		t := types[0]
		var res []common.Candle
		switch t {
		case string(gen.PostZigzagRequestTypeCsv):
			v := csvInfos[0]
			var csvInfo gen.CsvInfo
			if err := json.Unmarshal([]byte(v), &csvInfo); err != nil {
				// バリデーション済みのため発生しない想定のエラー
				panic("invalid csvInfo")
			}

			csvf, err := csvs[0].Open()
			if err != nil {
				return nil, lang.NewFxtError(lang.ErrInvalidParameterError, "csv")
			}
			defer csvf.Close()

			res, err = reader.ReadCandleCsv(csvInfo, csvf)
			if err != nil {
				return nil, lang.NewFxtError(lang.ErrInvalidParameterError, "csv").SetCause(err)
			}

		case string(gen.PostZigzagRequestTypeCandles):
			candles := []gen.Candle{}
			if err := json.Unmarshal([]byte(candless[0]), &candles); err != nil {
				// バリデーション済みのため発生しない想定のエラー
				panic("invalid candles")
			}

			// gen.Candle -> common.Candle に変換
			for _, c := range candles {
				t, err := common.ToTime(c.Time)
				if err != nil {
					// バリデーション済みのため発生しない想定のエラー
					panic("invalid candles")
				}
				res = append(res, common.Candle{
					Time:  *t,
					High:  float64(c.High),
					Open:  float64(c.Open),
					Close: float64(c.Close),
					Low:   float64(c.Low),
				})
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

	// ジグザグの計算
	pbs := algo.FindZigzagPeakToBottom(paramCandles)
	bps := algo.FindZigzagBottomToPeak(paramCandles)

	// 時刻順に並び替える
	zigzags := append(pbs, bps...)
	sort.Slice(zigzags, func(i, j int) bool {
		return zigzags[i].StartTime.Unix() < zigzags[j].StartTime.Unix()
	})

	items := []gen.Zigzag{}
	for _, z := range zigzags {
		items = append(items, gen.Zigzag{
			BottomIndex: z.BottomIndex,
			Delta:       float32(z.Delta),
			Kind:        z.Kind,
			PeakIndex:   z.PeakIndex,
			StartTime:   z.StartTime.Format(time.RFC3339),
			Velocity:    float32(z.Velocity),
		})
	}

	return ctx.JSON(http.StatusCreated, gen.PostZigzagResult{
		Count: len(items),
		Items: items,
	})
}
