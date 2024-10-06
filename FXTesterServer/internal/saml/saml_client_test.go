package saml

import (
	"context"
	"database/sql"
	"errors"
	"fxtester/internal/common"
	"fxtester/internal/gen"
	"fxtester/internal/lang"
	"fxtester/internal/net"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	cs "github.com/crewjam/saml"
	cssp "github.com/crewjam/saml/samlsp"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/html"
)

type MockSamlClientDelegator struct {
	delegateOpenFile          func(path string) (io.ReadCloser, error)
	delegateFetchMetadata     func(ctx context.Context, url url.URL, timeout time.Duration) (*cs.EntityDescriptor, error)
	delegateParseAuthResponse func(sp cs.ServiceProvider, request *http.Request, possibleRequestIds []string) (*cs.Assertion, error)
}

func (m *MockSamlClientDelegator) OpenFile(path string) (io.ReadCloser, error) {
	return m.delegateOpenFile(path)
}
func (m *MockSamlClientDelegator) FetchMetadata(ctx context.Context, url url.URL, timeout time.Duration) (*cs.EntityDescriptor, error) {
	return m.delegateFetchMetadata(ctx, url, timeout)
}
func (m *MockSamlClientDelegator) ParseAuthResponse(sp cs.ServiceProvider, request *http.Request, possibleRequestIds []string) (*cs.Assertion, error) {
	return m.delegateParseAuthResponse(sp, request, possibleRequestIds)
}

type MockDB struct {
	db *sql.DB
}

func (m *MockDB) Init() error {
	return nil
}

func (m *MockDB) GetDB() *sql.DB {
	return m.db
}

type MockReaderCloser struct {
	deleteRead func(p []byte) (n int, err error)
}

func (m *MockReaderCloser) Close() error {
	return nil
}

func (m *MockReaderCloser) Read(p []byte) (n int, err error) {
	return m.deleteRead(p)
}

func Test_SamlClient_Init(t *testing.T) {
	type args struct {
		samlClient     ISamlClient
		idpMetadataUrl string
		backendURL     string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test1_file_normal",
			args: args{
				samlClient: func() ISamlClient {
					r := &MockSamlClientDelegator{
						delegateOpenFile: func(path string) (io.ReadCloser, error) {
							r := strings.NewReader(TestDataIdpMetadata)
							return &MockReaderCloser{
								deleteRead: func(p []byte) (n int, err error) {
									return r.Read(p)
								},
							}, nil
						},
						delegateFetchMetadata: func(ctx context.Context, url url.URL, timeout time.Duration) (*cs.EntityDescriptor, error) {
							return nil, nil
						},
					}
					db, _, err := sqlmock.New()
					if err != nil {
						t.Errorf("failed sqlmock.New(): %v", err)
					}
					idb := &MockDB{
						db: db,
					}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
			},
		},
		{
			name: "test2_file_error",
			args: args{
				samlClient: func() ISamlClient {
					r := &MockSamlClientDelegator{
						delegateOpenFile: func(path string) (io.ReadCloser, error) {
							r := strings.NewReader(TestDataIdpMetadata)
							return &MockReaderCloser{
								deleteRead: func(p []byte) (n int, err error) {
									return r.Read(p)
								},
							}, nil
						},
						delegateFetchMetadata: func(ctx context.Context, url url.URL, timeout time.Duration) (*cs.EntityDescriptor, error) {
							return nil, nil
						},
					}
					db, _, err := sqlmock.New()
					if err != nil {
						t.Errorf("failed sqlmock.New(): %v", err)
					}
					idb := &MockDB{
						db: db,
					}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     string([]byte{0x7f}), // 不正なURL
			},
			wantErr: true,
		},
		{
			name: "test3_file_error",
			args: args{
				samlClient: func() ISamlClient {
					r := &MockSamlClientDelegator{
						delegateOpenFile: func(path string) (io.ReadCloser, error) {
							r := strings.NewReader("abc") // 不正なIDP Metadata
							return &MockReaderCloser{
								deleteRead: func(p []byte) (n int, err error) {
									return r.Read(p)
								}}, nil
						},
						delegateFetchMetadata: func(ctx context.Context, url url.URL, timeout time.Duration) (*cs.EntityDescriptor, error) {
							return nil, nil
						},
					}
					db, _, err := sqlmock.New()
					if err != nil {
						t.Errorf("failed sqlmock.New(): %v", err)
					}
					idb := &MockDB{
						db: db,
					}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
			},
			wantErr: true,
		},
		{
			name: "test4_file_error",
			args: args{
				samlClient: func() ISamlClient {
					r := &MockSamlClientDelegator{
						delegateOpenFile: func(path string) (io.ReadCloser, error) {
							return nil, errors.New("test error") // ファイルOpenのエラー
						},
						delegateFetchMetadata: func(ctx context.Context, url url.URL, timeout time.Duration) (*cs.EntityDescriptor, error) {
							return nil, nil
						},
					}
					db, _, err := sqlmock.New()
					if err != nil {
						t.Errorf("failed sqlmock.New(): %v", err)
					}
					idb := &MockDB{
						db: db,
					}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
			},
			wantErr: true,
		},
		{
			name: "test5_file_error",
			args: args{
				samlClient: func() ISamlClient {
					r := &MockSamlClientDelegator{
						delegateOpenFile: func(path string) (io.ReadCloser, error) {
							return &MockReaderCloser{
								deleteRead: func(p []byte) (n int, err error) {
									return 0, errors.New("test error") // Readのエラー
								},
							}, nil
						},
						delegateFetchMetadata: func(ctx context.Context, url url.URL, timeout time.Duration) (*cs.EntityDescriptor, error) {
							return nil, nil
						},
					}
					db, _, err := sqlmock.New()
					if err != nil {
						t.Errorf("failed sqlmock.New(): %v", err)
					}
					idb := &MockDB{
						db: db,
					}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
			},
			wantErr: true,
		},
		{
			name: "test6_download_normal",
			args: args{
				samlClient: func() ISamlClient {
					r := &MockSamlClientDelegator{
						delegateOpenFile: func(path string) (io.ReadCloser, error) {
							return nil, nil
						},
						delegateFetchMetadata: func(ctx context.Context, url url.URL, timeout time.Duration) (*cs.EntityDescriptor, error) {
							descriptor, err := cssp.ParseMetadata([]byte(TestDataIdpMetadata))
							if err != nil {
								t.Errorf("failed ParseMetadata: %v", err)
							}
							return descriptor, nil
						},
					}
					db, _, err := sqlmock.New()
					if err != nil {
						t.Errorf("failed sqlmock.New(): %v", err)
					}
					idb := &MockDB{
						db: db,
					}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "https://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
			},
		},
		{
			name: "test7_download_error",
			args: args{
				samlClient: func() ISamlClient {
					r := &MockSamlClientDelegator{
						delegateOpenFile: func(path string) (io.ReadCloser, error) {
							return nil, nil
						},
						delegateFetchMetadata: func(ctx context.Context, url url.URL, timeout time.Duration) (*cs.EntityDescriptor, error) {
							descriptor, err := cssp.ParseMetadata([]byte(TestDataIdpMetadata))
							if err != nil {
								t.Errorf("failed ParseMetadata: %v", err)
							}
							return descriptor, nil
						},
					}
					db, _, err := sqlmock.New()
					if err != nil {
						t.Errorf("failed sqlmock.New(): %v", err)
					}
					idb := &MockDB{
						db: db,
					}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "https://" + string([]byte{0x7f}), // 不正なURL
				backendURL:     common.GetConfig().Saml.BackendURL,
			},
			wantErr: true,
		},
		{
			name: "test8_download_normal",
			args: args{
				samlClient: func() ISamlClient {
					r := &MockSamlClientDelegator{
						delegateOpenFile: func(path string) (io.ReadCloser, error) {
							return nil, nil
						},
						delegateFetchMetadata: func() func(ctx context.Context, url url.URL, timeout time.Duration) (*cs.EntityDescriptor, error) {
							count := 0

							return func(ctx context.Context, url url.URL, timeout time.Duration) (*cs.EntityDescriptor, error) {
								defer func() {
									count++
								}()
								if count == 0 {
									// 初回は必ず失敗する
									return nil, errors.New("test-error")
								}
								// 二回目以降
								descriptor, err := cssp.ParseMetadata([]byte(TestDataIdpMetadata))
								if err != nil {
									t.Errorf("failed ParseMetadata: %v", err)
								}
								return descriptor, nil
							}
						}(),
					}
					db, _, err := sqlmock.New()
					if err != nil {
						t.Errorf("failed sqlmock.New(): %v", err)
					}
					idb := &MockDB{
						db: db,
					}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "https://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
			},
		},
		{
			name: "test9_download_error",
			args: args{
				samlClient: func() ISamlClient {
					r := &MockSamlClientDelegator{
						delegateOpenFile: func(path string) (io.ReadCloser, error) {
							return nil, nil
						},
						delegateFetchMetadata: func(ctx context.Context, url url.URL, timeout time.Duration) (*cs.EntityDescriptor, error) {
							return nil, errors.New("test-error") // 必ずエラー
						},
					}
					db, _, err := sqlmock.New()
					if err != nil {
						t.Errorf("failed sqlmock.New(): %v", err)
					}
					idb := &MockDB{
						db: db,
					}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "https://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
			},
			wantErr: true,
		},
		{
			name: "test10_error",
			args: args{
				samlClient: func() ISamlClient {
					r := &MockSamlClientDelegator{
						delegateOpenFile: func(path string) (io.ReadCloser, error) {
							return nil, nil
						},
						delegateFetchMetadata: func(ctx context.Context, url url.URL, timeout time.Duration) (*cs.EntityDescriptor, error) {
							return nil, nil
						},
					}
					idb := &MockDB{}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "xxx://test", // 不正なスキーム
				backendURL:     common.GetConfig().Saml.BackendURL,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		saveBackendURL := common.GetConfig().Saml.BackendURL
		saveIdpMetadataUrl := common.GetConfig().Saml.IdpMetadataUrl

		t.Run(tt.name, func(t *testing.T) {

			common.GetConfig().Saml.BackendURL = tt.args.backendURL
			common.GetConfig().Saml.IdpMetadataUrl = tt.args.idpMetadataUrl

			if err := tt.args.samlClient.Init(); (err != nil) != tt.wantErr {
				t.Errorf("Init()=%v want=%v", err, tt.wantErr)
			} else if err != nil {
				if _, ok := err.(*lang.FxtError); !ok {
					t.Errorf("invalid error type: %v", err)
				}
			}
		})

		common.GetConfig().Saml.IdpMetadataUrl = saveIdpMetadataUrl
		common.GetConfig().Saml.BackendURL = saveBackendURL
	}
}

func Test_SamlClient_ExecuteSamlLogin(t *testing.T) {
	type args struct {
		samlClient     ISamlClient
		idpMetadataUrl string
		backendURL     string
		ctx            func(w http.ResponseWriter) echo.Context
		params         gen.GetSamlLoginParams
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test1_normal",
			args: args{
				samlClient: func() ISamlClient {
					r := &MockSamlClientDelegator{
						delegateOpenFile: func(path string) (io.ReadCloser, error) {
							r := strings.NewReader(TestDataIdpMetadata)
							return &MockReaderCloser{
								deleteRead: func(p []byte) (n int, err error) {
									return r.Read(p)
								},
							}, nil
						},
						delegateFetchMetadata: func(ctx context.Context, url url.URL, timeout time.Duration) (*cs.EntityDescriptor, error) {
							return nil, nil
						},
					}
					db, _, err := sqlmock.New()
					if err != nil {
						t.Errorf("failed sqlmock.New(): %v", err)
					}
					idb := &MockDB{
						db: db,
					}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
				ctx: func(w http.ResponseWriter) echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhsot", nil)
					return echo.New().NewContext(req, w)
				},
				params: gen.GetSamlLoginParams{
					XRedirectURL:        "https://localhost/test-redirect",
					XRedirectURLOnError: "https://localhost/test-redirect-error",
				},
			},
		},
	}

	for _, tt := range tests {
		saveBackendURL := common.GetConfig().Saml.BackendURL
		saveIdpMetadataUrl := common.GetConfig().Saml.IdpMetadataUrl

		t.Run(tt.name, func(t *testing.T) {

			common.GetConfig().Saml.BackendURL = tt.args.backendURL
			common.GetConfig().Saml.IdpMetadataUrl = tt.args.idpMetadataUrl

			if err := tt.args.samlClient.Init(); err != nil {
				t.Errorf("Init()=%v want=%v", err, tt.wantErr)
			} else {
				rec := httptest.NewRecorder()

				err := tt.args.samlClient.ExecuteSamlLogin(tt.args.ctx(rec), tt.args.params)
				if (err != nil) != tt.wantErr {
					t.Errorf("ExecuteSamlLogin()=%v want=%v", err, tt.wantErr)
				} else if err != nil {
					if _, ok := err.(*lang.FxtError); !ok {
						t.Errorf("invalid error type: %v", err)
					}
				} else {
					// クッキーのチェック ここから ==>
					parser := &http.Request{Header: http.Header{"Cookie": rec.Header()["Set-Cookie"]}}
					c, err := parser.Cookie(net.NameSSOToken)
					if err != nil {
						t.Errorf("invalid cookie: %v", err)
					}
					claims, err := net.VerifyToken[net.SSOSessionPayload](c.Value, net.SSOSessionSecret)
					if err != nil {
						t.Errorf("invalid cookie: %v", err)
					}
					if claims.Value.AuthnRequestId == "" {
						t.Error("Empty AuthnRequestId")
					}
					if claims.Value.RedirectURL != "https://localhost/test-redirect" {
						t.Errorf("invalid RedirectURL: %v", claims.Value.RedirectURL)
					}
					if claims.Value.RedirectURLOnError != "https://localhost/test-redirect-error" {
						t.Errorf("invalid RedirectURLOnError: %v", claims.Value.RedirectURL)
					}
					// <== ここまで クッキーのチェック

					// HTMLのフォーマットチェック
					if _, err := html.Parse(rec.Body); err != nil {
						t.Errorf("invalid body: %v", err)
					}

					// ContentTypeのチェック
					if contentType := rec.Header().Get(echo.HeaderContentType); echo.MIMETextHTML != contentType {
						t.Errorf("invalid ContentType: %v", contentType)
					}
				}
			}
		})

		common.GetConfig().Saml.IdpMetadataUrl = saveIdpMetadataUrl
		common.GetConfig().Saml.BackendURL = saveBackendURL
	}
}

func Test_SamlClient_ExecuteSamlAcs(t *testing.T) {
	type args struct {
		samlClient     ISamlClient
		idpMetadataUrl string
		backendURL     string
		ctx            func(w http.ResponseWriter) echo.Context
	}

	tests := []struct {
		name            string
		args            args
		wantErr         bool
		wantRedirectURL string
	}{
		{
			name: "test1_normal",
			args: args{
				samlClient: func() ISamlClient {
					r := &MockSamlClientDelegator{
						delegateOpenFile: func(path string) (io.ReadCloser, error) {
							r := strings.NewReader(TestDataIdpMetadata)
							return &MockReaderCloser{
								deleteRead: func(p []byte) (n int, err error) {
									return r.Read(p)
								},
							}, nil
						},
						delegateParseAuthResponse: func(sp cs.ServiceProvider, request *http.Request, possibleRequestIds []string) (*cs.Assertion, error) {
							return &cs.Assertion{
								Subject: &cs.Subject{
									NameID: &cs.NameID{
										Value: "test-name-id",
									},
								},
							}, nil
						},
					}
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Errorf("failed sqlmock.New(): %v", err)
					}
					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta(`select id, email, access_token, refresh_token from fxtester_schema.select_user_with_email($1)`)).WithArgs("test-name-id").WillReturnRows(sqlmock.NewRows([]string{"id", "email", "access_token", "refresh_token"}).AddRow(0, "test-name-id", "access", "refresh"))
					mock.ExpectCommit()
					idb := &MockDB{
						db: db,
					}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
				ctx: func(w http.ResponseWriter) echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhsot", nil)

					now := time.Now()
					expires := now.Add(60 * time.Minute)

					payload := net.SSOSessionPayload{
						AuthnRequestId:     "test-authn-request-id",
						RedirectURL:        "http://localhost/test-redirect",
						RedirectURLOnError: "http://localhost/test-redirect-test",
					}

					token, err := net.GenerateToken(payload, expires, net.SSOSessionSecret)
					if err != nil {
						t.Errorf("failed net.GenerateToken: %v", err)
					}

					cookie := http.Cookie{
						Name:  net.NameSSOToken,
						Value: token,
					}

					req.Header.Set("Cookie", cookie.String())
					return echo.New().NewContext(req, w)
				},
			},
			wantRedirectURL: "http://localhost/test-redirect",
		},
		{
			name: "test2_normal",
			args: args{
				samlClient: func() ISamlClient {
					r := &MockSamlClientDelegator{
						delegateOpenFile: func(path string) (io.ReadCloser, error) {
							r := strings.NewReader(TestDataIdpMetadata)
							return &MockReaderCloser{
								deleteRead: func(p []byte) (n int, err error) {
									return r.Read(p)
								},
							}, nil
						},
						delegateParseAuthResponse: func(sp cs.ServiceProvider, request *http.Request, possibleRequestIds []string) (*cs.Assertion, error) {
							return &cs.Assertion{
								Subject: &cs.Subject{
									NameID: &cs.NameID{
										Value: "test-name-id",
									},
								},
							}, nil
						},
					}
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Errorf("failed sqlmock.New(): %v", err)
					}
					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta(`select id, email, access_token, refresh_token from fxtester_schema.select_user_with_email($1)`)).WithArgs("test-name-id").WillReturnRows(sqlmock.NewRows([]string{"id", "email", "access_token", "refresh_token"}))
					mock.ExpectQuery(regexp.QuoteMeta(`select fxtester_schema.create_user($1)`)).WithArgs("test-name-id").WillReturnRows(sqlmock.NewRows([]string{"res"}).AddRow(0))
					mock.ExpectCommit()
					idb := &MockDB{
						db: db,
					}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
				ctx: func(w http.ResponseWriter) echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhsot", nil)

					now := time.Now()
					expires := now.Add(60 * time.Minute)

					payload := net.SSOSessionPayload{
						AuthnRequestId:     "test-authn-request-id",
						RedirectURL:        "http://localhost/test-redirect",
						RedirectURLOnError: "http://localhost/test-redirect-test",
					}

					token, err := net.GenerateToken(payload, expires, net.SSOSessionSecret)
					if err != nil {
						t.Errorf("failed net.GenerateToken: %v", err)
					}

					cookie := http.Cookie{
						Name:  net.NameSSOToken,
						Value: token,
					}

					req.Header.Set("Cookie", cookie.String())
					return echo.New().NewContext(req, w)
				},
			},
			wantRedirectURL: "http://localhost/test-redirect",
		},
		{
			name: "test3_error",
			args: args{
				samlClient: func() ISamlClient {
					r := &MockSamlClientDelegator{
						delegateOpenFile: func(path string) (io.ReadCloser, error) {
							r := strings.NewReader(TestDataIdpMetadata)
							return &MockReaderCloser{
								deleteRead: func(p []byte) (n int, err error) {
									return r.Read(p)
								},
							}, nil
						},
					}
					idb := &MockDB{}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
				ctx: func(w http.ResponseWriter) echo.Context {
					// PostForm()でエラーを発生させる
					mockReader := &MockReaderCloser{
						deleteRead: func(p []byte) (n int, err error) {
							return 0, errors.New("test-error")
						},
					}
					req := httptest.NewRequest(echo.POST, "https://localhost", mockReader)
					req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
					return echo.New().NewContext(req, w)
				},
			},
			wantErr: true,
		},

		{
			name: "test4_error",
			args: args{
				samlClient: func() ISamlClient {
					r := &MockSamlClientDelegator{
						delegateOpenFile: func(path string) (io.ReadCloser, error) {
							r := strings.NewReader(TestDataIdpMetadata)
							return &MockReaderCloser{
								deleteRead: func(p []byte) (n int, err error) {
									return r.Read(p)
								},
							}, nil
						},
					}
					idb := &MockDB{}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
				ctx: func(w http.ResponseWriter) echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhost", nil)
					// SSOセッションのクッキー未設定
					return echo.New().NewContext(req, w)
				},
			},
			wantErr: true,
		},

		{
			name: "test5_error",
			args: args{
				samlClient: func() ISamlClient {
					r := &MockSamlClientDelegator{
						delegateOpenFile: func(path string) (io.ReadCloser, error) {
							r := strings.NewReader(TestDataIdpMetadata)
							return &MockReaderCloser{
								deleteRead: func(p []byte) (n int, err error) {
									return r.Read(p)
								},
							}, nil
						},
						delegateParseAuthResponse: func(sp cs.ServiceProvider, request *http.Request, possibleRequestIds []string) (*cs.Assertion, error) {
							return nil, errors.New("test-error") // parseResponseのエラー
						},
					}
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Errorf("failed sqlmock.New(): %v", err)
					}
					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta(`select id, email, access_token, refresh_token from fxtester_schema.select_user_with_email($1)`)).WithArgs("test-name-id").WillReturnRows(sqlmock.NewRows([]string{"id", "email", "access_token", "refresh_token"}).AddRow(0, "test-name-id", "access", "refresh"))
					mock.ExpectCommit()
					idb := &MockDB{
						db: db,
					}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
				ctx: func(w http.ResponseWriter) echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhsot", nil)

					now := time.Now()
					expires := now.Add(60 * time.Minute)

					payload := net.SSOSessionPayload{
						AuthnRequestId:     "test-authn-request-id",
						RedirectURL:        "http://localhost/test-redirect",
						RedirectURLOnError: "http://localhost/test-redirect-test",
					}

					token, err := net.GenerateToken(payload, expires, net.SSOSessionSecret)
					if err != nil {
						t.Errorf("failed net.GenerateToken: %v", err)
					}

					cookie := http.Cookie{
						Name:  net.NameSSOToken,
						Value: token,
					}

					req.Header.Set("Cookie", cookie.String())
					return echo.New().NewContext(req, w)
				},
			},
			wantRedirectURL: "http://localhost/test-redirect-test?saml_error=1",
		},

		{
			name: "test6_error",
			args: args{
				samlClient: func() ISamlClient {
					r := &MockSamlClientDelegator{
						delegateOpenFile: func(path string) (io.ReadCloser, error) {
							r := strings.NewReader(TestDataIdpMetadata)
							return &MockReaderCloser{
								deleteRead: func(p []byte) (n int, err error) {
									return r.Read(p)
								},
							}, nil
						},
						delegateParseAuthResponse: func(sp cs.ServiceProvider, request *http.Request, possibleRequestIds []string) (*cs.Assertion, error) {
							return nil, nil // assertionがnull
						},
					}
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Errorf("failed sqlmock.New(): %v", err)
					}
					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta(`select id, email, access_token, refresh_token from fxtester_schema.select_user_with_email($1)`)).WithArgs("test-name-id").WillReturnRows(sqlmock.NewRows([]string{"id", "email", "access_token", "refresh_token"}).AddRow(0, "test-name-id", "access", "refresh"))
					mock.ExpectCommit()
					idb := &MockDB{
						db: db,
					}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
				ctx: func(w http.ResponseWriter) echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhsot", nil)

					now := time.Now()
					expires := now.Add(60 * time.Minute)

					payload := net.SSOSessionPayload{
						AuthnRequestId:     "test-authn-request-id",
						RedirectURL:        "http://localhost/test-redirect",
						RedirectURLOnError: "http://localhost/test-redirect-test",
					}

					token, err := net.GenerateToken(payload, expires, net.SSOSessionSecret)
					if err != nil {
						t.Errorf("failed net.GenerateToken: %v", err)
					}

					cookie := http.Cookie{
						Name:  net.NameSSOToken,
						Value: token,
					}

					req.Header.Set("Cookie", cookie.String())
					return echo.New().NewContext(req, w)
				},
			},
			wantRedirectURL: "http://localhost/test-redirect-test?saml_error=1",
		},

		{
			name: "test7_error",
			args: args{
				samlClient: func() ISamlClient {
					r := &MockSamlClientDelegator{
						delegateOpenFile: func(path string) (io.ReadCloser, error) {
							r := strings.NewReader(TestDataIdpMetadata)
							return &MockReaderCloser{
								deleteRead: func(p []byte) (n int, err error) {
									return r.Read(p)
								},
							}, nil
						},
						delegateParseAuthResponse: func(sp cs.ServiceProvider, request *http.Request, possibleRequestIds []string) (*cs.Assertion, error) {
							return &cs.Assertion{}, nil // assertion.Subjectがnull
						},
					}
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Errorf("failed sqlmock.New(): %v", err)
					}
					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta(`select id, email, access_token, refresh_token from fxtester_schema.select_user_with_email($1)`)).WithArgs("test-name-id").WillReturnRows(sqlmock.NewRows([]string{"id", "email", "access_token", "refresh_token"}).AddRow(0, "test-name-id", "access", "refresh"))
					mock.ExpectCommit()
					idb := &MockDB{
						db: db,
					}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
				ctx: func(w http.ResponseWriter) echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhsot", nil)

					now := time.Now()
					expires := now.Add(60 * time.Minute)

					payload := net.SSOSessionPayload{
						AuthnRequestId:     "test-authn-request-id",
						RedirectURL:        "http://localhost/test-redirect",
						RedirectURLOnError: "http://localhost/test-redirect-test",
					}

					token, err := net.GenerateToken(payload, expires, net.SSOSessionSecret)
					if err != nil {
						t.Errorf("failed net.GenerateToken: %v", err)
					}

					cookie := http.Cookie{
						Name:  net.NameSSOToken,
						Value: token,
					}

					req.Header.Set("Cookie", cookie.String())
					return echo.New().NewContext(req, w)
				},
			},
			wantRedirectURL: "http://localhost/test-redirect-test?saml_error=1",
		},

		{
			name: "test8_error",
			args: args{
				samlClient: func() ISamlClient {
					r := &MockSamlClientDelegator{
						delegateOpenFile: func(path string) (io.ReadCloser, error) {
							r := strings.NewReader(TestDataIdpMetadata)
							return &MockReaderCloser{
								deleteRead: func(p []byte) (n int, err error) {
									return r.Read(p)
								},
							}, nil
						},
						delegateParseAuthResponse: func(sp cs.ServiceProvider, request *http.Request, possibleRequestIds []string) (*cs.Assertion, error) {
							return &cs.Assertion{
								Subject: &cs.Subject{},
							}, nil // assertion.Subject.NameIdがnull
						},
					}
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Errorf("failed sqlmock.New(): %v", err)
					}
					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta(`select id, email, access_token, refresh_token from fxtester_schema.select_user_with_email($1)`)).WithArgs("test-name-id").WillReturnRows(sqlmock.NewRows([]string{"id", "email", "access_token", "refresh_token"}).AddRow(0, "test-name-id", "access", "refresh"))
					mock.ExpectCommit()
					idb := &MockDB{
						db: db,
					}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
				ctx: func(w http.ResponseWriter) echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhsot", nil)

					now := time.Now()
					expires := now.Add(60 * time.Minute)

					payload := net.SSOSessionPayload{
						AuthnRequestId:     "test-authn-request-id",
						RedirectURL:        "http://localhost/test-redirect",
						RedirectURLOnError: "http://localhost/test-redirect-test",
					}

					token, err := net.GenerateToken(payload, expires, net.SSOSessionSecret)
					if err != nil {
						t.Errorf("failed net.GenerateToken: %v", err)
					}

					cookie := http.Cookie{
						Name:  net.NameSSOToken,
						Value: token,
					}

					req.Header.Set("Cookie", cookie.String())
					return echo.New().NewContext(req, w)
				},
			},
			wantRedirectURL: "http://localhost/test-redirect-test?saml_error=1",
		},

		{
			name: "test9_error",
			args: args{
				samlClient: func() ISamlClient {
					r := &MockSamlClientDelegator{
						delegateOpenFile: func(path string) (io.ReadCloser, error) {
							r := strings.NewReader(TestDataIdpMetadata)
							return &MockReaderCloser{
								deleteRead: func(p []byte) (n int, err error) {
									return r.Read(p)
								},
							}, nil
						},
						delegateParseAuthResponse: func(sp cs.ServiceProvider, request *http.Request, possibleRequestIds []string) (*cs.Assertion, error) {
							return &cs.Assertion{
								Subject: &cs.Subject{
									NameID: &cs.NameID{},
								},
							}, nil // assertion.Subject.NameId.Valueが空文字
						},
					}
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Errorf("failed sqlmock.New(): %v", err)
					}
					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta(`select id, email, access_token, refresh_token from fxtester_schema.select_user_with_email($1)`)).WithArgs("test-name-id").WillReturnRows(sqlmock.NewRows([]string{"id", "email", "access_token", "refresh_token"}).AddRow(0, "test-name-id", "access", "refresh"))
					mock.ExpectCommit()
					idb := &MockDB{
						db: db,
					}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
				ctx: func(w http.ResponseWriter) echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhsot", nil)

					now := time.Now()
					expires := now.Add(60 * time.Minute)

					payload := net.SSOSessionPayload{
						AuthnRequestId:     "test-authn-request-id",
						RedirectURL:        "http://localhost/test-redirect",
						RedirectURLOnError: "http://localhost/test-redirect-test",
					}

					token, err := net.GenerateToken(payload, expires, net.SSOSessionSecret)
					if err != nil {
						t.Errorf("failed net.GenerateToken: %v", err)
					}

					cookie := http.Cookie{
						Name:  net.NameSSOToken,
						Value: token,
					}

					req.Header.Set("Cookie", cookie.String())
					return echo.New().NewContext(req, w)
				},
			},
			wantRedirectURL: "http://localhost/test-redirect-test?saml_error=1",
		},
		{
			name: "test10_error",
			args: args{
				samlClient: func() ISamlClient {
					r := &MockSamlClientDelegator{
						delegateOpenFile: func(path string) (io.ReadCloser, error) {
							r := strings.NewReader(TestDataIdpMetadata)
							return &MockReaderCloser{
								deleteRead: func(p []byte) (n int, err error) {
									return r.Read(p)
								},
							}, nil
						},
						delegateParseAuthResponse: func(sp cs.ServiceProvider, request *http.Request, possibleRequestIds []string) (*cs.Assertion, error) {
							return &cs.Assertion{
								Subject: &cs.Subject{
									NameID: &cs.NameID{
										Value: "test-name-id",
									},
								},
							}, nil
						},
					}
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Errorf("failed sqlmock.New(): %v", err)
					}
					mock.ExpectBegin().WillReturnError(errors.New("test-error")) // トランザクション開始エラー
					idb := &MockDB{
						db: db,
					}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
				ctx: func(w http.ResponseWriter) echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhsot", nil)

					now := time.Now()
					expires := now.Add(60 * time.Minute)

					payload := net.SSOSessionPayload{
						AuthnRequestId:     "test-authn-request-id",
						RedirectURL:        "http://localhost/test-redirect",
						RedirectURLOnError: "http://localhost/test-redirect-test",
					}

					token, err := net.GenerateToken(payload, expires, net.SSOSessionSecret)
					if err != nil {
						t.Errorf("failed net.GenerateToken: %v", err)
					}

					cookie := http.Cookie{
						Name:  net.NameSSOToken,
						Value: token,
					}

					req.Header.Set("Cookie", cookie.String())
					return echo.New().NewContext(req, w)
				},
			},
			wantRedirectURL: "http://localhost/test-redirect-test?saml_error=1",
		},
		{
			name: "test11_error",
			args: args{
				samlClient: func() ISamlClient {
					r := &MockSamlClientDelegator{
						delegateOpenFile: func(path string) (io.ReadCloser, error) {
							r := strings.NewReader(TestDataIdpMetadata)
							return &MockReaderCloser{
								deleteRead: func(p []byte) (n int, err error) {
									return r.Read(p)
								},
							}, nil
						},
						delegateParseAuthResponse: func(sp cs.ServiceProvider, request *http.Request, possibleRequestIds []string) (*cs.Assertion, error) {
							return &cs.Assertion{
								Subject: &cs.Subject{
									NameID: &cs.NameID{
										Value: "test-name-id",
									},
								},
							}, nil
						},
					}
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Errorf("failed sqlmock.New(): %v", err)
					}
					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta(`select id, email, access_token, refresh_token from fxtester_schema.select_user_with_email($1)`)).WithArgs("test-name-id").WillReturnRows(sqlmock.NewRows([]string{"id", "email", "access_token", "refresh_token"}))
					mock.ExpectQuery(regexp.QuoteMeta(`select fxtester_schema.create_user($1)`)).WillReturnError(errors.New("test-error"))
					mock.ExpectRollback()
					idb := &MockDB{
						db: db,
					}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
				ctx: func(w http.ResponseWriter) echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhsot", nil)

					now := time.Now()
					expires := now.Add(60 * time.Minute)

					payload := net.SSOSessionPayload{
						AuthnRequestId:     "test-authn-request-id",
						RedirectURL:        "http://localhost/test-redirect",
						RedirectURLOnError: "http://localhost/test-redirect-test",
					}

					token, err := net.GenerateToken(payload, expires, net.SSOSessionSecret)
					if err != nil {
						t.Errorf("failed net.GenerateToken: %v", err)
					}

					cookie := http.Cookie{
						Name:  net.NameSSOToken,
						Value: token,
					}

					req.Header.Set("Cookie", cookie.String())
					return echo.New().NewContext(req, w)
				},
			},
			wantRedirectURL: "http://localhost/test-redirect-test?saml_error=1",
		},
		{
			name: "test12_error",
			args: args{
				samlClient: func() ISamlClient {
					r := &MockSamlClientDelegator{
						delegateOpenFile: func(path string) (io.ReadCloser, error) {
							r := strings.NewReader(TestDataIdpMetadata)
							return &MockReaderCloser{
								deleteRead: func(p []byte) (n int, err error) {
									return r.Read(p)
								},
							}, nil
						},
						delegateParseAuthResponse: func(sp cs.ServiceProvider, request *http.Request, possibleRequestIds []string) (*cs.Assertion, error) {
							return &cs.Assertion{
								Subject: &cs.Subject{
									NameID: &cs.NameID{
										Value: "test-name-id",
									},
								},
							}, nil
						},
					}
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Errorf("failed sqlmock.New(): %v", err)
					}
					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta(`select id, email, access_token, refresh_token from fxtester_schema.select_user_with_email($1)`)).WithArgs("test-name-id").WillReturnRows(sqlmock.NewRows([]string{"id", "email", "access_token", "refresh_token"}))
					mock.ExpectQuery(regexp.QuoteMeta(`select fxtester_schema.create_user($1)`)).WillReturnError(errors.New("test-error"))
					mock.ExpectRollback().WillReturnError(errors.New("test-error"))
					idb := &MockDB{
						db: db,
					}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
				ctx: func(w http.ResponseWriter) echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhsot", nil)

					now := time.Now()
					expires := now.Add(60 * time.Minute)

					payload := net.SSOSessionPayload{
						AuthnRequestId:     "test-authn-request-id",
						RedirectURL:        "http://localhost/test-redirect",
						RedirectURLOnError: "http://localhost/test-redirect-test",
					}

					token, err := net.GenerateToken(payload, expires, net.SSOSessionSecret)
					if err != nil {
						t.Errorf("failed net.GenerateToken: %v", err)
					}

					cookie := http.Cookie{
						Name:  net.NameSSOToken,
						Value: token,
					}

					req.Header.Set("Cookie", cookie.String())
					return echo.New().NewContext(req, w)
				},
			},
			wantRedirectURL: "http://localhost/test-redirect-test?saml_error=1",
		},
		{
			name: "test13_error",
			args: args{
				samlClient: func() ISamlClient {
					r := &MockSamlClientDelegator{
						delegateOpenFile: func(path string) (io.ReadCloser, error) {
							r := strings.NewReader(TestDataIdpMetadata)
							return &MockReaderCloser{
								deleteRead: func(p []byte) (n int, err error) {
									return r.Read(p)
								},
							}, nil
						},
						delegateParseAuthResponse: func(sp cs.ServiceProvider, request *http.Request, possibleRequestIds []string) (*cs.Assertion, error) {
							return &cs.Assertion{
								Subject: &cs.Subject{
									NameID: &cs.NameID{
										Value: "test-name-id",
									},
								},
							}, nil
						},
					}
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Errorf("failed sqlmock.New(): %v", err)
					}
					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta(`select id, email, access_token, refresh_token from fxtester_schema.select_user_with_email($1)`)).WithArgs("test-name-id").WillReturnRows(sqlmock.NewRows([]string{"id", "email", "access_token", "refresh_token"}).AddRow(0, "test-name-id", "access", "refresh"))
					mock.ExpectCommit().WillReturnError(errors.New("test-error"))
					idb := &MockDB{
						db: db,
					}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
				ctx: func(w http.ResponseWriter) echo.Context {
					req := httptest.NewRequest(echo.POST, "https://localhsot", nil)

					now := time.Now()
					expires := now.Add(60 * time.Minute)

					payload := net.SSOSessionPayload{
						AuthnRequestId:     "test-authn-request-id",
						RedirectURL:        "http://localhost/test-redirect",
						RedirectURLOnError: "http://localhost/test-redirect-test",
					}

					token, err := net.GenerateToken(payload, expires, net.SSOSessionSecret)
					if err != nil {
						t.Errorf("failed net.GenerateToken: %v", err)
					}

					cookie := http.Cookie{
						Name:  net.NameSSOToken,
						Value: token,
					}

					req.Header.Set("Cookie", cookie.String())
					return echo.New().NewContext(req, w)
				},
			},
			wantRedirectURL: "http://localhost/test-redirect-test?saml_error=1",
		},
	}

	for _, tt := range tests {
		saveBackendURL := common.GetConfig().Saml.BackendURL
		saveIdpMetadataUrl := common.GetConfig().Saml.IdpMetadataUrl

		t.Run(tt.name, func(t *testing.T) {

			common.GetConfig().Saml.BackendURL = tt.args.backendURL
			common.GetConfig().Saml.IdpMetadataUrl = tt.args.idpMetadataUrl

			if err := tt.args.samlClient.Init(); err != nil {
				t.Errorf("Init() err=%v", err)
			} else {
				rec := httptest.NewRecorder()

				err := tt.args.samlClient.ExecuteSamlAcs(tt.args.ctx(rec))
				if (err != nil) != tt.wantErr {
					// ExecuteSamlAcsはエラーがあった場合でもnilを返却する
					t.Errorf("ExecuteSamlAcs()=%v wantErr=%v", err, tt.wantErr)
				} else {
					if err != nil {
						// エラーの型チェック
						if _, ok := err.(*lang.FxtError); !ok {
							t.Errorf("invalid error type: %v", err)
						}
					}

					redirectURL := rec.Header().Get(echo.HeaderLocation)
					if redirectURL != tt.wantRedirectURL {
						// リダイレクト先URLのチェック
						t.Errorf("ExecuteSamlAcs()=%v wantRedirectURL=%v", redirectURL, tt.wantRedirectURL)
					}
				}

			}
		})

		common.GetConfig().Saml.IdpMetadataUrl = saveIdpMetadataUrl
		common.GetConfig().Saml.BackendURL = saveBackendURL
	}
}
