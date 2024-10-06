// Package common 共通機能をまとめたパッケージ
package common

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type Config struct {
	// 一般設定
	Server struct {
		// ポート番号
		Port uint16 `yaml:"port"`
		// CORSで接続を許可するオリジン一覧
		AllowOrigins []string `yaml:"allowOrigins"`
		// SSL設定
		Ssl struct {
			// SSLの有効化
			IsEnabled bool `yaml:"isEnabled"`
			// 証明書ファイルのパス
			CertPath string `yaml:"certPath"`
			// キーファイルのパス
			KeyPath string `yaml:"keyPath"`
		} `yaml:"ssl"`
	} `yaml:"server"`

	// DB設定
	Db struct {
		// データベース名前(e.g. postgres)
		Name string `yaml:"name"`
		// Data Source Name
		Dsn string `yaml:"dsn"`
		// 最大アイドル時間
		MaxIdleConnections int `yaml:"maxIdleConnections"`
		// 最大オープン接続数
		MaxOpenConnections int `yaml:"maxOpenConnections"`
		// 最大ライフタイム(秒単位)
		MaxLifeTimeBySec int `yaml:"maxLifeTimeBySec"`
	} `yaml:"db"`

	// SAML設定
	Saml struct {
		Keycloak struct {
			// keycloakのログインユーザの情報 (e.g. admin)
			AdminUser struct {
				Username string `yaml:"username"`
				Password string `yaml:"password"`
			} `yaml:"adminUser"`
			// keycloakのURL (e.g. http://localhost:28080)
			BaseURL string `yaml:"baseURL"`
			// realmの名前
			RealmName string `yaml:"realmName"`
			NewUsers  []struct {
				Username string `yaml:"username"`
				Password string `yaml:"password"`
				Email    string `yaml:"email"`
			} `yaml:"newUsers"`
			NewClientId string `yaml:"newClientId"`
		} `yaml:"keycloak"`

		// idpのmetadata.xmlを返却するURLもしくはファイルパス
		IdpMetadataUrl string `yaml:"idpMetadataUrl"`
		// ルートURL (リダイレクト先のベースURL)
		RootURL string `yaml:"rootURL"`
		// SAMLクライアントのEntityId
		EntityId string `yaml:"entityId"`
		// バックエンドのURL
		BackendURL string `yaml:"backendURL"`
		// Valid redirect URI
		ValidRedirectURI string `yaml:"validRedirectURI"`
		// Valid post logout redirect URI
		ValidPostLogoutRedirectURI string `yaml:"validPostLogoutRedirectURI"`
		// Logout Service POST Binding URL
		LogoutServicePostBindingURL string `yaml:"logoutServicePostBindingURL"`
	} `yaml:"saml"`

	// 辞書設定
	Dict struct {
		// 辞書ファイルのパス
		Path string `yaml:"path"`
	} `yaml:"dict"`

	// Websocket
	Websocket struct {
		// 最大接続数
		MaxConnections int `yaml:"maxConnections"`
	} `yaml:"websocket"`
}

var once sync.Once
var config *Config = &Config{}

func GetConfig() *Config {
	f := func() {
		config = loadConfig()
	}
	once.Do(f)
	return config
}

func loadConfig() *Config {
	// 環境変数からプロジェクトパスを取得
	projectPath, err := GetEnvAs("PROJECT_PATH", false, "")
	if err != nil {
		panic(fmt.Errorf("environment variable PROJECT_PATH couldn't be obtained: %v", err))
	}
	// プロジェクトパスをベースに設定ファイルへのパスを取得
	configPath := fmt.Sprintf("%s/settings/config.yaml", projectPath)
	// 設定ファイルの読み込み
	configFileBytes, err := os.ReadFile(configPath)
	if err != nil {
		panic(fmt.Sprintf("failed to read %s: %v", configPath, err))
	}

	// テンプレートの作成
	tmp, err := template.New("template").Parse(string(configFileBytes))
	if err != nil {
		panic(fmt.Errorf("failed to parse template: %v", err))
	}

	// テンプレートの適用
	a := bytes.NewBufferString("")
	if err := tmp.Execute(a, map[string]string{
		"pwd": projectPath,
	}); err != nil {
		panic(fmt.Errorf("failed to execute template: %v", err))
	}

	// ConfigのUnmarshal
	config := &Config{}
	if err = yaml.Unmarshal(a.Bytes(), config); err != nil {
		panic(fmt.Errorf("failed to unmarshal: %v", err))
	}

	return config
}
