package validator

import (
	"encoding/json"
	"fxtester/internal/gen"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func Test_ValidatePostZigzag(t *testing.T) {
	type args struct {
		ctx echo.Context
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "正常ケース1",
			args: args{
				ctx: func() echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhost:8100", nil)
					w := httptest.NewRecorder()
					ctx := echo.New().NewContext(req, w)

					ctx.Request().MultipartForm = &multipart.Form{
						Value: map[string][]string{
							"type": {
								string(gen.PostZigzagRequestTypeCsv),
							},
							"csvInfo": {
								func() string {
									var csvInfo gen.CsvInfo
									csvInfo.DelimiterChar = ","
									csvInfo.CloseColumnIndex = 0
									csvInfo.HighColumnIndex = 1
									csvInfo.LowColumnIndex = 2
									csvInfo.OpenColumnIndex = 3
									csvInfo.TimeColumnIndex = 4
									bytes, err := json.Marshal(csvInfo)
									if err != nil {
										t.Errorf("failed to create gen.CsvInfo: %v", err)
									}
									return string(bytes)
								}(),
							},
						},
						File: map[string][]*multipart.FileHeader{
							"csv": {
								{},
							},
						},
					}
					return ctx
				}(),
			},
		},
		{
			name: "正常ケース2(区切り文字がスペース)",
			args: args{
				ctx: func() echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhost:8100", nil)
					w := httptest.NewRecorder()
					ctx := echo.New().NewContext(req, w)

					ctx.Request().MultipartForm = &multipart.Form{
						Value: map[string][]string{
							"type": {
								string(gen.PostZigzagRequestTypeCsv),
							},
							"csvInfo": {
								func() string {
									var csvInfo gen.CsvInfo
									csvInfo.DelimiterChar = " "
									csvInfo.CloseColumnIndex = 0
									csvInfo.HighColumnIndex = 1
									csvInfo.LowColumnIndex = 2
									csvInfo.OpenColumnIndex = 3
									csvInfo.TimeColumnIndex = 4
									bytes, err := json.Marshal(csvInfo)
									if err != nil {
										t.Errorf("failed to create gen.CsvInfo: %v", err)
									}
									return string(bytes)
								}(),
							},
						},
						File: map[string][]*multipart.FileHeader{
							"csv": {
								{},
							},
						},
					}
					return ctx
				}(),
			},
		},
		{
			name: "正常ケース3(区切り文字がタブ)",
			args: args{
				ctx: func() echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhost:8100", nil)
					w := httptest.NewRecorder()
					ctx := echo.New().NewContext(req, w)

					ctx.Request().MultipartForm = &multipart.Form{
						Value: map[string][]string{
							"type": {
								string(gen.PostZigzagRequestTypeCsv),
							},
							"csvInfo": {
								func() string {
									var csvInfo gen.CsvInfo
									csvInfo.DelimiterChar = "\t"
									csvInfo.CloseColumnIndex = 0
									csvInfo.HighColumnIndex = 1
									csvInfo.LowColumnIndex = 2
									csvInfo.OpenColumnIndex = 3
									csvInfo.TimeColumnIndex = 4
									bytes, err := json.Marshal(csvInfo)
									if err != nil {
										t.Errorf("failed to create gen.CsvInfo: %v", err)
									}
									return string(bytes)
								}(),
							},
						},
						File: map[string][]*multipart.FileHeader{
							"csv": {
								{},
							},
						},
					}
					return ctx
				}(),
			},
		},
		{
			name: "正常ケース4",
			args: args{
				ctx: func() echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhost:8100", nil)
					w := httptest.NewRecorder()
					ctx := echo.New().NewContext(req, w)

					ctx.Request().MultipartForm = &multipart.Form{
						Value: map[string][]string{
							"type": {
								string(gen.PostZigzagRequestTypeCandles),
							},
							"candles": {
								func() string {
									candles := []gen.Candle{
										{
											Time:  "2024-01-01T00:00:00Z",
											High:  4,
											Open:  3,
											Close: 2,
											Low:   1,
										},
										{
											Time:  "2024-01-02T00:00:00Z",
											High:  5,
											Open:  3,
											Close: 4,
											Low:   2,
										},
									}
									bytes, err := json.Marshal(candles)
									if err != nil {
										t.Errorf("failed to create []gen.Candle: %v", err)
									}
									return string(bytes)
								}(),
							},
						},
						File: map[string][]*multipart.FileHeader{},
					}
					return ctx
				}(),
			},
		},
		{
			name: "必須パラメータ(type)未指定",
			args: args{
				ctx: func() echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhost:8100", nil)
					w := httptest.NewRecorder()
					ctx := echo.New().NewContext(req, w)

					ctx.Request().MultipartForm = &multipart.Form{
						Value: map[string][]string{
							"type":    {},
							"csvInfo": {},
						},
						File: map[string][]*multipart.FileHeader{
							"csv": {},
						},
					}
					return ctx
				}(),
			},
			wantErr: true,
		},
		{
			name: "typeにcsvまたはcandles以外を指定",
			args: args{
				ctx: func() echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhost:8100", nil)
					w := httptest.NewRecorder()
					ctx := echo.New().NewContext(req, w)

					ctx.Request().MultipartForm = &multipart.Form{
						Value: map[string][]string{
							"type": {
								"abc",
							},
							"candles": {
								func() string {
									candles := []gen.Candle{
										{
											Time:  "2024-01-01T00:00:00Z",
											High:  4,
											Open:  3,
											Close: 2,
											Low:   1,
										},
										{
											Time:  "2024-01-02T00:00:00Z",
											High:  5,
											Open:  3,
											Close: 4,
											Low:   2,
										},
									}
									bytes, err := json.Marshal(candles)
									if err != nil {
										t.Errorf("failed to create []gen.Candle: %v", err)
									}
									return string(bytes)
								}(),
							},
						},
						File: map[string][]*multipart.FileHeader{},
					}
					return ctx
				}(),
			},
			wantErr: true,
		},
		{
			name: "csvInfoに空文字",
			args: args{
				ctx: func() echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhost:8100", nil)
					w := httptest.NewRecorder()
					ctx := echo.New().NewContext(req, w)

					ctx.Request().MultipartForm = &multipart.Form{
						Value: map[string][]string{
							"type": {
								string(gen.PostZigzagRequestTypeCsv),
							},
							"csvInfo": {
								"",
							},
						},
						File: map[string][]*multipart.FileHeader{
							"csv": {},
						},
					}
					return ctx
				}(),
			},
			wantErr: true,
		},
		{
			name: "csvパラメータ未指定",
			args: args{
				ctx: func() echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhost:8100", nil)
					w := httptest.NewRecorder()
					ctx := echo.New().NewContext(req, w)

					ctx.Request().MultipartForm = &multipart.Form{
						Value: map[string][]string{
							"type": {
								string(gen.PostZigzagRequestTypeCsv),
							},
							"csvInfo": {
								func() string {
									var csvInfo gen.CsvInfo
									csvInfo.DelimiterChar = ","
									csvInfo.CloseColumnIndex = 0
									csvInfo.HighColumnIndex = 1
									csvInfo.LowColumnIndex = 2
									csvInfo.OpenColumnIndex = 3
									csvInfo.TimeColumnIndex = 4
									bytes, err := json.Marshal(csvInfo)
									if err != nil {
										t.Errorf("failed to create gen.CsvInfo: %v", err)
									}
									return string(bytes)
								}(),
							},
						},
						File: map[string][]*multipart.FileHeader{
							"csv": {},
						},
					}
					return ctx
				}(),
			},
			wantErr: true,
		},
		{
			name: "candlesが空文字",
			args: args{
				ctx: func() echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhost:8100", nil)
					w := httptest.NewRecorder()
					ctx := echo.New().NewContext(req, w)

					ctx.Request().MultipartForm = &multipart.Form{
						Value: map[string][]string{
							"type": {
								string(gen.PostZigzagRequestTypeCandles),
							},
							"candles": {
								"",
							},
						},
						File: map[string][]*multipart.FileHeader{},
					}
					return ctx
				}(),
			},
			wantErr: true,
		},
		{
			name: "csvInfoに不正文字",
			args: args{
				ctx: func() echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhost:8100", nil)
					w := httptest.NewRecorder()
					ctx := echo.New().NewContext(req, w)

					ctx.Request().MultipartForm = &multipart.Form{
						Value: map[string][]string{
							"type": {
								string(gen.PostZigzagRequestTypeCsv),
							},
							"csvInfo": {
								"abc",
							},
						},
						File: map[string][]*multipart.FileHeader{
							"csv": {
								{},
							},
						},
					}
					return ctx
				}(),
			},
			wantErr: true,
		},
		{
			name: "インデックスが重複1(csvInfo)",
			args: args{
				ctx: func() echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhost:8100", nil)
					w := httptest.NewRecorder()
					ctx := echo.New().NewContext(req, w)

					ctx.Request().MultipartForm = &multipart.Form{
						Value: map[string][]string{
							"type": {
								string(gen.PostZigzagRequestTypeCsv),
							},
							"csvInfo": {
								func() string {
									var csvInfo gen.CsvInfo
									csvInfo.DelimiterChar = ","
									csvInfo.CloseColumnIndex = 0
									csvInfo.HighColumnIndex = 0
									csvInfo.LowColumnIndex = 2
									csvInfo.OpenColumnIndex = 3
									csvInfo.TimeColumnIndex = 4
									bytes, err := json.Marshal(csvInfo)
									if err != nil {
										t.Errorf("failed to create gen.CsvInfo: %v", err)
									}
									return string(bytes)
								}(),
							},
						},
						File: map[string][]*multipart.FileHeader{
							"csv": {
								{},
							},
						},
					}
					return ctx
				}(),
			},
			wantErr: true,
		},
		{
			name: "インデックスが重複2(csvInfo)",
			args: args{
				ctx: func() echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhost:8100", nil)
					w := httptest.NewRecorder()
					ctx := echo.New().NewContext(req, w)

					ctx.Request().MultipartForm = &multipart.Form{
						Value: map[string][]string{
							"type": {
								string(gen.PostZigzagRequestTypeCsv),
							},
							"csvInfo": {
								func() string {
									var csvInfo gen.CsvInfo
									csvInfo.DelimiterChar = ","
									csvInfo.CloseColumnIndex = 0
									csvInfo.HighColumnIndex = 1
									csvInfo.LowColumnIndex = 1
									csvInfo.OpenColumnIndex = 3
									csvInfo.TimeColumnIndex = 4
									bytes, err := json.Marshal(csvInfo)
									if err != nil {
										t.Errorf("failed to create gen.CsvInfo: %v", err)
									}
									return string(bytes)
								}(),
							},
						},
						File: map[string][]*multipart.FileHeader{
							"csv": {
								{},
							},
						},
					}
					return ctx
				}(),
			},
			wantErr: true,
		},
		{
			name: "インデックスが重複3(csvInfo)",
			args: args{
				ctx: func() echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhost:8100", nil)
					w := httptest.NewRecorder()
					ctx := echo.New().NewContext(req, w)

					ctx.Request().MultipartForm = &multipart.Form{
						Value: map[string][]string{
							"type": {
								string(gen.PostZigzagRequestTypeCsv),
							},
							"csvInfo": {
								func() string {
									var csvInfo gen.CsvInfo
									csvInfo.DelimiterChar = ","
									csvInfo.CloseColumnIndex = 0
									csvInfo.HighColumnIndex = 1
									csvInfo.LowColumnIndex = 2
									csvInfo.OpenColumnIndex = 2
									csvInfo.TimeColumnIndex = 4
									bytes, err := json.Marshal(csvInfo)
									if err != nil {
										t.Errorf("failed to create gen.CsvInfo: %v", err)
									}
									return string(bytes)
								}(),
							},
						},
						File: map[string][]*multipart.FileHeader{
							"csv": {
								{},
							},
						},
					}
					return ctx
				}(),
			},
			wantErr: true,
		},
		{
			name: "インデックスが重複4(csvInfo)",
			args: args{
				ctx: func() echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhost:8100", nil)
					w := httptest.NewRecorder()
					ctx := echo.New().NewContext(req, w)

					ctx.Request().MultipartForm = &multipart.Form{
						Value: map[string][]string{
							"type": {
								string(gen.PostZigzagRequestTypeCsv),
							},
							"csvInfo": {
								func() string {
									var csvInfo gen.CsvInfo
									csvInfo.DelimiterChar = ","
									csvInfo.CloseColumnIndex = 0
									csvInfo.HighColumnIndex = 1
									csvInfo.LowColumnIndex = 2
									csvInfo.OpenColumnIndex = 3
									csvInfo.TimeColumnIndex = 3
									bytes, err := json.Marshal(csvInfo)
									if err != nil {
										t.Errorf("failed to create gen.CsvInfo: %v", err)
									}
									return string(bytes)
								}(),
							},
						},
						File: map[string][]*multipart.FileHeader{
							"csv": {
								{},
							},
						},
					}
					return ctx
				}(),
			},
			wantErr: true,
		},
		{
			name: "区切り文字が２文字",
			args: args{
				ctx: func() echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhost:8100", nil)
					w := httptest.NewRecorder()
					ctx := echo.New().NewContext(req, w)

					ctx.Request().MultipartForm = &multipart.Form{
						Value: map[string][]string{
							"type": {
								string(gen.PostZigzagRequestTypeCsv),
							},
							"csvInfo": {
								func() string {
									var csvInfo gen.CsvInfo
									csvInfo.DelimiterChar = ",,"
									csvInfo.CloseColumnIndex = 0
									csvInfo.HighColumnIndex = 1
									csvInfo.LowColumnIndex = 2
									csvInfo.OpenColumnIndex = 3
									csvInfo.TimeColumnIndex = 4
									bytes, err := json.Marshal(csvInfo)
									if err != nil {
										t.Errorf("failed to create gen.CsvInfo: %v", err)
									}
									return string(bytes)
								}(),
							},
						},
						File: map[string][]*multipart.FileHeader{
							"csv": {
								{},
							},
						},
					}
					return ctx
				}(),
			},
			wantErr: true,
		},
		{
			name: "区切り文字が[,\\t\\s]以外",
			args: args{
				ctx: func() echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhost:8100", nil)
					w := httptest.NewRecorder()
					ctx := echo.New().NewContext(req, w)

					ctx.Request().MultipartForm = &multipart.Form{
						Value: map[string][]string{
							"type": {
								string(gen.PostZigzagRequestTypeCsv),
							},
							"csvInfo": {
								func() string {
									var csvInfo gen.CsvInfo
									csvInfo.DelimiterChar = "a"
									csvInfo.CloseColumnIndex = 0
									csvInfo.HighColumnIndex = 1
									csvInfo.LowColumnIndex = 2
									csvInfo.OpenColumnIndex = 3
									csvInfo.TimeColumnIndex = 4
									bytes, err := json.Marshal(csvInfo)
									if err != nil {
										t.Errorf("failed to create gen.CsvInfo: %v", err)
									}
									return string(bytes)
								}(),
							},
						},
						File: map[string][]*multipart.FileHeader{
							"csv": {
								{},
							},
						},
					}
					return ctx
				}(),
			},
			wantErr: true,
		},
		{
			name: "candlesに不正なjson",
			args: args{
				ctx: func() echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhost:8100", nil)
					w := httptest.NewRecorder()
					ctx := echo.New().NewContext(req, w)

					ctx.Request().MultipartForm = &multipart.Form{
						Value: map[string][]string{
							"type": {
								string(gen.PostZigzagRequestTypeCandles),
							},
							"candles": {
								"abc",
							},
						},
						File: map[string][]*multipart.FileHeader{},
					}
					return ctx
				}(),
			},
			wantErr: true,
		},
		{
			name: "gen.Candleのバリデーション失敗",
			args: args{
				ctx: func() echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhost:8100", nil)
					w := httptest.NewRecorder()
					ctx := echo.New().NewContext(req, w)

					ctx.Request().MultipartForm = &multipart.Form{
						Value: map[string][]string{
							"type": {
								string(gen.PostZigzagRequestTypeCandles),
							},
							"candles": {
								func() string {
									candles := []gen.Candle{
										{},
									}
									bytes, err := json.Marshal(candles)
									if err != nil {
										t.Errorf("failed to create []gen.Candle: %v", err)
									}
									return string(bytes)
								}(),
							},
						},
						File: map[string][]*multipart.FileHeader{},
					}
					return ctx
				}(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := ValidatePostZigzag(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("ValidatePostZigzag()=%v wantErr=%v", err, tt.wantErr)
			}
		})
	}
}
