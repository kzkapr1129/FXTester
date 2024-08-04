package internal

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

type Config struct {
	Port                     uint16
	DSN                      string
	DatabaseName             string
	AllowOrigins             []string
	MaxIdleConnections       int
	MaxOpenConnections       int
	ConnectionMaxLifeTimeSec int
	IdpMetadataURL           string
	SslCertPath              string
	SslKeyPath               string
	SamlCertPath             string
	SamlKeyPath              string
	RootUrl                  string
	EntityID                 string
	RedirectUrlAfterLogin    string
	RedirectUrlAfterLogout   string
	BlacklistRedirectUrls    map[string]struct{}
	AccessTokenKey           string
	RefreshTokenKey          string
	DictFilePath             string
}

var once sync.Once
var config *Config = &Config{}

func GetConfig() *Config {
	var err error
	f := func() {
		config, err = loadConfig()
		if err != nil {
			panic(fmt.Errorf("failed to loadConfig: %w", err))
		}
	}
	once.Do(f)
	return config
}

func loadConfig() (*Config, error) {
	errs := []error{}

	port, err := GetEnvAs[uint16]("PORT", false, 8080)
	if err != nil {
		errs = append(errs, err)
	}

	databaseName, err := GetEnvAs("DATABASE_NAME", true, "")
	if err != nil {
		errs = append(errs, err)
	}

	var allowOrigins []string
	if v, err := GetEnvAs("ALLOW_ORIGINS", false, "https://127.0.0.1:3000,https://localhost:3000"); err != nil {
		errs = append(errs, err)
	} else {
		allowOrigins = strings.Split(v, ",")
		if len(allowOrigins) <= 0 {
			allowOrigins = []string{"*"}
		}
	}

	dsn, err := GetEnvAs("DSN", true, "")
	if err != nil {
		errs = append(errs, err)
	}

	maxIdleConnections, err := GetEnvAs("MAX_IDLE_CONNECTIONS", false, 10)
	if err != nil {
		errs = append(errs, err)
	}

	maxOpenConnections, err := GetEnvAs("MAX_OPEN_CONNECTIONS", false, 10)
	if err != nil {
		errs = append(errs, err)
	}

	connectionMaxLifeTimeSec, err := GetEnvAs("CONNECTION_MAX_LIFE_TIME_SEC", false, 10)
	if err != nil {
		errs = append(errs, err)
	}

	idpMetadataUrl, err := GetEnvAs("IDP_METADATA_URL", true, "")
	fmt.Println("idpMetadataUrl: ", idpMetadataUrl)
	if err != nil {
		errs = append(errs, err)
	}

	sslCertPath, err := GetEnvAs("SSL_CERT_PATH", true, "")
	if err != nil {
		errs = append(errs, err)
	}

	sslKeyPath, err := GetEnvAs("SSL_KEY_PATH", true, "")
	if err != nil {
		errs = append(errs, err)
	}

	samlCertPath, err := GetEnvAs("SAML_CERT_PATH", true, "")
	if err != nil {
		errs = append(errs, err)
	}

	samlKeyPath, err := GetEnvAs("SAML_KEY_PATH", true, "")
	if err != nil {
		errs = append(errs, err)
	}

	rootUrl, err := GetEnvAs("ROOT_URL", true, "")
	if err != nil {
		errs = append(errs, err)
	}

	entityID, err := GetEnvAs("ENTITY_ID", true, "")
	if err != nil {
		errs = append(errs, err)
	}

	redirectUrlAfterLogin, err := GetEnvAs("REDIRECT_URL_AFTER_LOGIN", true, "")
	if err != nil {
		errs = append(errs, err)
	}

	redirectUrlAfterLogout, err := GetEnvAs("REDIRECT_URL_AFTER_LOGOUT", true, "")
	if err != nil {
		errs = append(errs, err)
	}

	accessTokenKey, err := GetEnvAs("ACCESS_TOKEN_KEY", true, "")
	if err != nil {
		errs = append(errs, err)
	}

	refreshTokenKey, err := GetEnvAs("REFRESH_TOKEN_KEY", true, "")
	if err != nil {
		errs = append(errs, err)
	}

	dictFilePath, err := GetEnvAs("DICT_FILE_PATH", true, "")
	if err != nil {
		errs = append(errs, err)
	}

	if 0 < len(errs) {
		return nil, errors.Join(errs...)
	}

	return &Config{
		Port:                     port,
		DSN:                      dsn,
		DatabaseName:             databaseName,
		AllowOrigins:             allowOrigins,
		MaxIdleConnections:       maxIdleConnections,
		MaxOpenConnections:       maxOpenConnections,
		ConnectionMaxLifeTimeSec: connectionMaxLifeTimeSec,
		IdpMetadataURL:           idpMetadataUrl,
		SslCertPath:              sslCertPath,
		SslKeyPath:               sslKeyPath,
		SamlCertPath:             samlCertPath,
		SamlKeyPath:              samlKeyPath,
		RootUrl:                  rootUrl,
		EntityID:                 entityID,
		RedirectUrlAfterLogin:    redirectUrlAfterLogin,
		RedirectUrlAfterLogout:   redirectUrlAfterLogout,
		AccessTokenKey:           accessTokenKey,
		RefreshTokenKey:          refreshTokenKey,
		DictFilePath:             dictFilePath,
	}, nil
}
