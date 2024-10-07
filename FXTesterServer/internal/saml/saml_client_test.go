package saml

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fxtester/internal/common"
	"fxtester/internal/db"
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
	delegateOpenFile                      func(path string) (io.ReadCloser, error)
	delegateFetchMetadata                 func(ctx context.Context, url url.URL, timeout time.Duration) (*cs.EntityDescriptor, error)
	delegateParseAuthResponse             func(sp cs.ServiceProvider, request *http.Request, possibleRequestIds []string) (*cs.Assertion, error)
	delegateValidateLogoutResponseRequest func(sp cs.ServiceProvider, request *http.Request) error
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
func (m *MockSamlClientDelegator) ValidateLogoutResponseRequest(sp cs.ServiceProvider, request *http.Request) error {
	return m.delegateValidateLogoutResponseRequest(sp, request)
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

type MockUserDao struct {
	db.IUserEntityDao
}

func (*MockUserDao) UpdateToken(userId int64, accessToken, refreshToken string) error {
	// accessTokenとrefreshTokenの値が動的となり、sqlmockでは対処が難しいためメソッドをオーバーライドして対処する
	return nil
}

func NewCookieContext[T any](values []struct {
	name    string
	secret  []byte
	payload T
}, w http.ResponseWriter, t *testing.T) echo.Context {
	req := httptest.NewRequest(echo.POST, "https://localhost", nil)

	expires := time.Now().Add(60 * time.Minute)

	for _, v := range values {
		token, err := net.GenerateToken(v.payload, expires, v.secret)
		if err != nil {
			t.Errorf("failed net.GenerateToken: %v", err)
		}

		cookie := http.Cookie{
			Name:  v.name,
			Value: token,
		}

		req.Header.Add("Cookie", cookie.String())
	}

	return echo.New().NewContext(req, w)
}

func toFormUrlencoded(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func uint32ptr(v int) *uint32 {
	tmp := uint32(v)
	return &tmp
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
					}
					idb := &MockDB{}
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
					}
					idb := &MockDB{}
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
					}
					idb := &MockDB{}
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
					}
					idb := &MockDB{}
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
						delegateFetchMetadata: func(ctx context.Context, url url.URL, timeout time.Duration) (*cs.EntityDescriptor, error) {
							descriptor, err := cssp.ParseMetadata([]byte(TestDataIdpMetadata))
							if err != nil {
								t.Errorf("failed ParseMetadata: %v", err)
							}
							return descriptor, nil
						},
					}
					idb := &MockDB{}
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
						delegateFetchMetadata: func(ctx context.Context, url url.URL, timeout time.Duration) (*cs.EntityDescriptor, error) {
							descriptor, err := cssp.ParseMetadata([]byte(TestDataIdpMetadata))
							if err != nil {
								t.Errorf("failed ParseMetadata: %v", err)
							}
							return descriptor, nil
						},
					}
					idb := &MockDB{}
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
					idb := &MockDB{}
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
						delegateFetchMetadata: func(ctx context.Context, url url.URL, timeout time.Duration) (*cs.EntityDescriptor, error) {
							return nil, errors.New("test-error") // 必ずエラー
						},
					}
					idb := &MockDB{}
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
					}
					idb := &MockDB{}
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
	const expectUserId = 1000
	const expectEmail = "test@test.co.jp"

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
		wantSamlErr     *uint32
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
										Value: expectEmail,
									},
								},
							}, nil
						},
					}
					mockDB, mock, err := sqlmock.New()
					if err != nil {
						t.Errorf("failed sqlmock.New(): %v", err)
					}
					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta(`select id, email, access_token, refresh_token from fxtester_schema.select_user_with_email($1)`)).WithArgs(expectEmail).WillReturnRows(sqlmock.NewRows([]string{"id", "email", "access_token", "refresh_token"}).AddRow(expectUserId, expectEmail, "access", "refresh"))
					mock.ExpectCommit()
					idb := &MockDB{
						db: mockDB,
					}
					return &SamlClient{
						delegate: r,
						dao: &MockUserDao{
							IUserEntityDao: db.NewUserEntityDao(idb),
						},
					}
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
				ctx: func(w http.ResponseWriter) echo.Context {
					return NewCookieContext([]struct {
						name    string
						secret  []byte
						payload net.SSOSessionPayload
					}{
						{
							name:   net.NameSSOToken,
							secret: net.SSOSessionSecret,
							payload: net.SSOSessionPayload{
								AuthnRequestId:     "test-authn-request-id",
								RedirectURL:        "http://localhost/test-redirect",
								RedirectURLOnError: "http://localhost/test-redirect-test",
							},
						},
					}, w, t)
				},
			},
			wantRedirectURL: "http://localhost/test-redirect",
			wantSamlErr:     uint32ptr(0),
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
										Value: expectEmail,
									},
								},
							}, nil
						},
					}
					mockDB, mock, err := sqlmock.New()
					if err != nil {
						t.Errorf("failed sqlmock.New(): %v", err)
					}
					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta(`select id, email, access_token, refresh_token from fxtester_schema.select_user_with_email($1)`)).WithArgs(expectEmail).WillReturnRows(sqlmock.NewRows([]string{"id", "email", "access_token", "refresh_token"}))
					mock.ExpectQuery(regexp.QuoteMeta(`select fxtester_schema.create_user($1)`)).WithArgs(expectEmail).WillReturnRows(sqlmock.NewRows([]string{"res"}).AddRow(expectUserId))
					mock.ExpectCommit()
					idb := &MockDB{
						db: mockDB,
					}
					return &SamlClient{
						delegate: r,
						dao: &MockUserDao{
							IUserEntityDao: db.NewUserEntityDao(idb),
						},
					}
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
				ctx: func(w http.ResponseWriter) echo.Context {
					return NewCookieContext([]struct {
						name    string
						secret  []byte
						payload net.SSOSessionPayload
					}{
						{
							name:   net.NameSSOToken,
							secret: net.SSOSessionSecret,
							payload: net.SSOSessionPayload{
								AuthnRequestId:     "test-authn-request-id",
								RedirectURL:        "http://localhost/test-redirect",
								RedirectURLOnError: "http://localhost/test-redirect-test",
							},
						},
					}, w, t)
				},
			},
			wantRedirectURL: "http://localhost/test-redirect",
			wantSamlErr:     uint32ptr(0),
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
					mock.ExpectQuery(regexp.QuoteMeta(`select id, email, access_token, refresh_token from fxtester_schema.select_user_with_email($1)`)).WithArgs(expectEmail).WillReturnRows(sqlmock.NewRows([]string{"id", "email", "access_token", "refresh_token"}).AddRow(expectUserId, expectEmail, "access", "refresh"))
					mock.ExpectCommit()
					idb := &MockDB{
						db: db,
					}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
				ctx: func(w http.ResponseWriter) echo.Context {
					return NewCookieContext([]struct {
						name    string
						secret  []byte
						payload net.SSOSessionPayload
					}{
						{
							name:   net.NameSSOToken,
							secret: net.SSOSessionSecret,
							payload: net.SSOSessionPayload{
								AuthnRequestId:     "test-authn-request-id",
								RedirectURL:        "http://localhost/test-redirect",
								RedirectURLOnError: "http://localhost/test-redirect-test",
							},
						},
					}, w, t)
				},
			},
			wantRedirectURL: "http://localhost/test-redirect-test?saml_error=1",
			wantSamlErr:     uint32ptr(0x80000010),
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
					mock.ExpectQuery(regexp.QuoteMeta(`select id, email, access_token, refresh_token from fxtester_schema.select_user_with_email($1)`)).WithArgs(expectEmail).WillReturnRows(sqlmock.NewRows([]string{"id", "email", "access_token", "refresh_token"}).AddRow(expectUserId, expectEmail, "access", "refresh"))
					mock.ExpectCommit()
					idb := &MockDB{
						db: db,
					}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
				ctx: func(w http.ResponseWriter) echo.Context {
					return NewCookieContext([]struct {
						name    string
						secret  []byte
						payload net.SSOSessionPayload
					}{
						{
							name:   net.NameSSOToken,
							secret: net.SSOSessionSecret,
							payload: net.SSOSessionPayload{
								AuthnRequestId:     "test-authn-request-id",
								RedirectURL:        "http://localhost/test-redirect",
								RedirectURLOnError: "http://localhost/test-redirect-test",
							},
						},
					}, w, t)
				},
			},
			wantRedirectURL: "http://localhost/test-redirect-test?saml_error=1",
			wantSamlErr:     uint32ptr(0x80000012),
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
					mock.ExpectQuery(regexp.QuoteMeta(`select id, email, access_token, refresh_token from fxtester_schema.select_user_with_email($1)`)).WithArgs(expectEmail).WillReturnRows(sqlmock.NewRows([]string{"id", "email", "access_token", "refresh_token"}).AddRow(expectUserId, expectEmail, "access", "refresh"))
					mock.ExpectCommit()
					idb := &MockDB{
						db: db,
					}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
				ctx: func(w http.ResponseWriter) echo.Context {
					return NewCookieContext([]struct {
						name    string
						secret  []byte
						payload net.SSOSessionPayload
					}{
						{
							name:   net.NameSSOToken,
							secret: net.SSOSessionSecret,
							payload: net.SSOSessionPayload{
								AuthnRequestId:     "test-authn-request-id",
								RedirectURL:        "http://localhost/test-redirect",
								RedirectURLOnError: "http://localhost/test-redirect-test",
							},
						},
					}, w, t)
				},
			},
			wantRedirectURL: "http://localhost/test-redirect-test?saml_error=1",
			wantSamlErr:     uint32ptr(0x80000012),
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
					mock.ExpectQuery(regexp.QuoteMeta(`select id, email, access_token, refresh_token from fxtester_schema.select_user_with_email($1)`)).WithArgs(expectEmail).WillReturnRows(sqlmock.NewRows([]string{"id", "email", "access_token", "refresh_token"}).AddRow(expectUserId, expectEmail, "access", "refresh"))
					mock.ExpectCommit()
					idb := &MockDB{
						db: db,
					}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
				ctx: func(w http.ResponseWriter) echo.Context {
					return NewCookieContext([]struct {
						name    string
						secret  []byte
						payload net.SSOSessionPayload
					}{
						{
							name:   net.NameSSOToken,
							secret: net.SSOSessionSecret,
							payload: net.SSOSessionPayload{
								AuthnRequestId:     "test-authn-request-id",
								RedirectURL:        "http://localhost/test-redirect",
								RedirectURLOnError: "http://localhost/test-redirect-test",
							},
						},
					}, w, t)
				},
			},
			wantRedirectURL: "http://localhost/test-redirect-test?saml_error=1",
			wantSamlErr:     uint32ptr(0x80000012),
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
					mock.ExpectQuery(regexp.QuoteMeta(`select id, email, access_token, refresh_token from fxtester_schema.select_user_with_email($1)`)).WithArgs(expectEmail).WillReturnRows(sqlmock.NewRows([]string{"id", "email", "access_token", "refresh_token"}).AddRow(expectUserId, expectEmail, "access", "refresh"))
					mock.ExpectCommit()
					idb := &MockDB{
						db: db,
					}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
				ctx: func(w http.ResponseWriter) echo.Context {
					return NewCookieContext([]struct {
						name    string
						secret  []byte
						payload net.SSOSessionPayload
					}{
						{
							name:   net.NameSSOToken,
							secret: net.SSOSessionSecret,
							payload: net.SSOSessionPayload{
								AuthnRequestId:     "test-authn-request-id",
								RedirectURL:        "http://localhost/test-redirect",
								RedirectURLOnError: "http://localhost/test-redirect-test",
							},
						},
					}, w, t)
				},
			},
			wantRedirectURL: "http://localhost/test-redirect-test?saml_error=1",
			wantSamlErr:     uint32ptr(0x80000012),
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
										Value: expectEmail,
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
					return NewCookieContext([]struct {
						name    string
						secret  []byte
						payload net.SSOSessionPayload
					}{
						{
							name:   net.NameSSOToken,
							secret: net.SSOSessionSecret,
							payload: net.SSOSessionPayload{
								AuthnRequestId:     "test-authn-request-id",
								RedirectURL:        "http://localhost/test-redirect",
								RedirectURLOnError: "http://localhost/test-redirect-test",
							},
						},
					}, w, t)
				},
			},
			wantRedirectURL: "http://localhost/test-redirect-test?saml_error=1",
			wantSamlErr:     uint32ptr(0x80000014),
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
										Value: expectEmail,
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
					mock.ExpectQuery(regexp.QuoteMeta(`select id, email, access_token, refresh_token from fxtester_schema.select_user_with_email($1)`)).WithArgs(expectEmail).WillReturnRows(sqlmock.NewRows([]string{"id", "email", "access_token", "refresh_token"}))
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
					return NewCookieContext([]struct {
						name    string
						secret  []byte
						payload net.SSOSessionPayload
					}{
						{
							name:   net.NameSSOToken,
							secret: net.SSOSessionSecret,
							payload: net.SSOSessionPayload{
								AuthnRequestId:     "test-authn-request-id",
								RedirectURL:        "http://localhost/test-redirect",
								RedirectURLOnError: "http://localhost/test-redirect-test",
							},
						},
					}, w, t)
				},
			},
			wantRedirectURL: "http://localhost/test-redirect-test?saml_error=1",
			wantSamlErr:     uint32ptr(0x80000017),
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
										Value: expectEmail,
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
					mock.ExpectQuery(regexp.QuoteMeta(`select id, email, access_token, refresh_token from fxtester_schema.select_user_with_email($1)`)).WillReturnError(errors.New("test-error"))
					mock.ExpectRollback()
					idb := &MockDB{
						db: db,
					}
					return NewSamlClient(r, idb)
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
				ctx: func(w http.ResponseWriter) echo.Context {
					return NewCookieContext([]struct {
						name    string
						secret  []byte
						payload net.SSOSessionPayload
					}{
						{
							name:   net.NameSSOToken,
							secret: net.SSOSessionSecret,
							payload: net.SSOSessionPayload{
								AuthnRequestId:     "test-authn-request-id",
								RedirectURL:        "http://localhost/test-redirect",
								RedirectURLOnError: "http://localhost/test-redirect-test",
							},
						},
					}, w, t)
				},
			},
			wantRedirectURL: "http://localhost/test-redirect-test?saml_error=1",
			wantSamlErr:     uint32ptr(0x80000017),
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
										Value: expectEmail,
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
					mock.ExpectQuery(regexp.QuoteMeta(`select id, email, access_token, refresh_token from fxtester_schema.select_user_with_email($1)`)).WithArgs(expectEmail).WillReturnRows(sqlmock.NewRows([]string{"id", "email", "access_token", "refresh_token"}))
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
					return NewCookieContext([]struct {
						name    string
						secret  []byte
						payload net.SSOSessionPayload
					}{
						{
							name:   net.NameSSOToken,
							secret: net.SSOSessionSecret,
							payload: net.SSOSessionPayload{
								AuthnRequestId:     "test-authn-request-id",
								RedirectURL:        "http://localhost/test-redirect",
								RedirectURLOnError: "http://localhost/test-redirect-test",
							},
						},
					}, w, t)
				},
			},
			wantRedirectURL: "http://localhost/test-redirect-test?saml_error=1",
			wantSamlErr:     uint32ptr(0x80000017), // ロールバックエラー時はlastErrorを書き換えないためErrDBQueryとなる
		},
		{
			name: "test14_error",
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
										Value: expectEmail,
									},
								},
							}, nil
						},
					}
					mockDB, mock, err := sqlmock.New()
					if err != nil {
						t.Errorf("failed sqlmock.New(): %v", err)
					}
					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta(`select id, email, access_token, refresh_token from fxtester_schema.select_user_with_email($1)`)).WithArgs(expectEmail).WillReturnRows(sqlmock.NewRows([]string{"id", "email", "access_token", "refresh_token"}).AddRow(expectUserId, expectEmail, "access", "refresh"))
					mock.ExpectCommit().WillReturnError(errors.New("test-error"))
					idb := &MockDB{
						db: mockDB,
					}
					return &SamlClient{
						delegate: r,
						dao: &MockUserDao{
							IUserEntityDao: db.NewUserEntityDao(idb),
						},
					}
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
				ctx: func(w http.ResponseWriter) echo.Context {
					return NewCookieContext([]struct {
						name    string
						secret  []byte
						payload net.SSOSessionPayload
					}{
						{
							name:   net.NameSSOToken,
							secret: net.SSOSessionSecret,
							payload: net.SSOSessionPayload{
								AuthnRequestId:     "test-authn-request-id",
								RedirectURL:        "http://localhost/test-redirect",
								RedirectURLOnError: "http://localhost/test-redirect-test",
							},
						},
					}, w, t)
				},
			},
			wantRedirectURL: "http://localhost/test-redirect-test?saml_error=1",
			wantSamlErr:     uint32ptr(0x80000016),
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

					parser := &http.Request{Header: http.Header{"Cookie": rec.Header()["Set-Cookie"]}}

					if redirectURL == "http://localhost/test-redirect" {
						// 成功時

						// クッキーのチェック ここから ==>
						// アクセストークン
						c, err := parser.Cookie(net.NameAccessToken)
						if err != nil {
							t.Errorf("invalid cookie: %v", err)
						}
						claims, err := net.VerifyToken[net.AuthSessionPayload](c.Value, net.AccessTokenSecret)
						if err != nil {
							t.Errorf("invalid cookie: %v", err)
						}
						if claims.Value.UserId != expectUserId {
							t.Errorf("Invalid wantUserId: %v", claims.Value.UserId)
						}
						if claims.Value.Email != expectEmail {
							t.Error("Empty Email")
						}

						// リフレッシュトークン
						c, err = parser.Cookie(net.NameRefreshToken)
						if err != nil {
							t.Errorf("invalid cookie: %v", err)
						}
						claims, err = net.VerifyToken[net.AuthSessionPayload](c.Value, net.RefreshTokenSecret)
						if err != nil {
							t.Errorf("invalid cookie: %v", err)
						}
						if claims.Value.UserId != expectUserId {
							t.Errorf("Invalid wantUserId: %v", claims.Value.UserId)
						}
						if claims.Value.Email != expectEmail {
							t.Error("Empty Email")
						}
					}

					// エラートークン
					c, err := parser.Cookie(net.NameSAMLErrorToken)
					if (tt.wantSamlErr != nil) != (err == nil) {
						t.Errorf("ExecuteSamlAcs()=%v, wantSamlErr=%v", err, tt.wantSamlErr)
					} else if err == nil {
						errClaims, err := net.VerifyToken[gen.ErrorWithTime](c.Value, net.SAMLErrorSessionSecret)
						if err != nil {
							t.Errorf("invalid cookie: %v", err)
						}
						if errClaims.Value.Err.Code != *tt.wantSamlErr {
							t.Errorf("ExecuteSamlAcs()=%v, wantSamlErr=%v", errClaims.Value.Err.Code, tt.wantSamlErr)
						}
					}
					// <== ここまで クッキーのチェック
				}
			}
		})

		common.GetConfig().Saml.IdpMetadataUrl = saveIdpMetadataUrl
		common.GetConfig().Saml.BackendURL = saveBackendURL
	}
}

func Test_SamlClient_ExecuteSamlLogout(t *testing.T) {
	type args struct {
		samlClient     ISamlClient
		idpMetadataUrl string
		backendURL     string
		ctx            func(w http.ResponseWriter) echo.Context
		params         gen.GetSamlLogoutParams
	}

	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantUserId int64
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
					return NewCookieContext([]struct {
						name    string
						secret  []byte
						payload any
					}{
						{
							name:   net.NameAccessToken,
							secret: net.AccessTokenSecret,
							payload: net.AuthSessionPayload{
								UserId: 100,
								Email:  "test-mail@test.co.jp",
							},
						},
					}, w, t)
				},
				params: gen.GetSamlLogoutParams{
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

				err := tt.args.samlClient.ExecuteSamlLogout(tt.args.ctx(rec), tt.args.params)
				if (err != nil) != tt.wantErr {
					t.Errorf("ExecuteSamlLogout()=%v want=%v", err, tt.wantErr)
				} else if err != nil {
					if _, ok := err.(*lang.FxtError); !ok {
						t.Errorf("invalid error type: %v", err)
					}
				} else {
					// クッキーのチェック ここから ==>
					parser := &http.Request{Header: http.Header{"Cookie": rec.Header()["Set-Cookie"]}}
					c, err := parser.Cookie(net.NameSLOToken)
					if err != nil {
						t.Errorf("invalid cookie: %v", err)
					}
					claims, err := net.VerifyToken[net.SLOSessionPayload](c.Value, net.SLOSessionSecret)
					if err != nil {
						t.Errorf("invalid cookie: %v", err)
					}
					if claims.Value.UserId == tt.wantUserId {
						t.Errorf("Invalid wantUserId: %v", claims.Value.UserId)
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

func Test_SamlClient_ExecuteSamlSlo(t *testing.T) {
	const expectUserId = 100
	type args struct {
		samlClient     ISamlClient
		idpMetadataUrl string
		backendURL     string
		ctx            func(w http.ResponseWriter) echo.Context
	}

	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantUserId int64
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
						delegateValidateLogoutResponseRequest: func(sp cs.ServiceProvider, request *http.Request) error {
							return nil
						},
					}
					mockDB, mock, err := sqlmock.New()
					if err != nil {
						t.Errorf("failed sqlmock.New(): %v", err)
					}
					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta(`call fxtester_schema.update_token($1, $2, $3)`)).WillReturnError(errors.New("test-error"))
					mock.ExpectCommit()
					idb := &MockDB{
						db: mockDB,
					}
					return &SamlClient{
						delegate: r,
						dao: &MockUserDao{
							IUserEntityDao: db.NewUserEntityDao(idb),
						},
					}
				}(),
				idpMetadataUrl: "file://test",
				backendURL:     common.GetConfig().Saml.BackendURL,
				ctx: func(w http.ResponseWriter) echo.Context {
					ctx := NewCookieContext([]struct {
						name    string
						secret  []byte
						payload any
					}{
						{
							name:   net.NameAccessToken,
							secret: net.AccessTokenSecret,
							payload: net.AuthSessionPayload{
								UserId: expectUserId,
								Email:  "test-mail@test.co.jp",
							},
						},
						{
							name:   net.NameSLOToken,
							secret: net.SLOSessionSecret,
							payload: net.SLOSessionPayload{
								UserId:             expectUserId,
								AuthnRequestId:     "test-authn-request-id",
								RedirectURL:        "http://localhost/test-redirect",
								RedirectURLOnError: "http://localhost/test-redirect-test",
							},
						},
					}, w, t)

					ctx.Request().Form = url.Values{}
					ctx.Request().Form.Add("SAMLResponse", toFormUrlencoded(TestDataLogoutResponse))

					return ctx
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

				err := tt.args.samlClient.ExecuteSamlSlo(tt.args.ctx(rec))
				if (err != nil) != tt.wantErr {
					t.Errorf("ExecuteSamlLogout()=%v want=%v", err, tt.wantErr)
				} else if err != nil {
					if _, ok := err.(*lang.FxtError); !ok {
						t.Errorf("invalid error type: %v", err)
					}
				}
			}
		})

		common.GetConfig().Saml.IdpMetadataUrl = saveIdpMetadataUrl
		common.GetConfig().Saml.BackendURL = saveBackendURL
	}
}
