package saml

import (
	"context"
	"database/sql"
	"errors"
	"fxtester/internal/common"
	"fxtester/internal/lang"
	"io"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	cs "github.com/crewjam/saml"
	cssp "github.com/crewjam/saml/samlsp"
)

type MockSamlClientReader struct {
	delegateOpenFile      func(path string) (io.ReadCloser, error)
	delegateFetchMetadata func(ctx context.Context, url url.URL, timeout time.Duration) (*cs.EntityDescriptor, error)
}

func (m *MockSamlClientReader) OpenFile(path string) (io.ReadCloser, error) {
	return m.delegateOpenFile(path)
}
func (m *MockSamlClientReader) FetchMetadata(ctx context.Context, url url.URL, timeout time.Duration) (*cs.EntityDescriptor, error) {
	return m.delegateFetchMetadata(ctx, url, timeout)
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
					r := &MockSamlClientReader{
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
					r := &MockSamlClientReader{
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
					r := &MockSamlClientReader{
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
					r := &MockSamlClientReader{
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
					r := &MockSamlClientReader{
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
					r := &MockSamlClientReader{
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
					r := &MockSamlClientReader{
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
			name: "test8_download_error",
			args: args{
				samlClient: func() ISamlClient {
					r := &MockSamlClientReader{
						delegateOpenFile: func(path string) (io.ReadCloser, error) {
							return nil, nil
						},
						delegateFetchMetadata: func(ctx context.Context, url url.URL, timeout time.Duration) (*cs.EntityDescriptor, error) {
							return nil, errors.New("test error") // fetchに失敗
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
			name: "test9_error",
			args: args{
				samlClient: func() ISamlClient {
					r := &MockSamlClientReader{
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
