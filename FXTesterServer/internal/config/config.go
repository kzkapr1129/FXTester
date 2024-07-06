package config

import (
	"errors"
	"fmt"
	"fxtester/internal"
	"sync"
)

type Config struct {
	Port                     uint16
	DSN                      string
	DatabaseName             string
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

	port, err := internal.GetEnvAs[uint16]("PORT", false, 8080)
	if err != nil {
		errs = append(errs, err)
	}

	databaseName, err := internal.GetEnvAs("DATABASE_NAME", true, "")
	if err != nil {
		errs = append(errs, err)
	}

	dsn, err := internal.GetEnvAs("DSN", true, "")
	if err != nil {
		errs = append(errs, err)
	}

	maxIdleConnections, err := internal.GetEnvAs("MAX_IDLE_CONNECTIONS", false, 10)
	if err != nil {
		errs = append(errs, err)
	}

	maxOpenConnections, err := internal.GetEnvAs("MAX_OPEN_CONNECTIONS", false, 10)
	if err != nil {
		errs = append(errs, err)
	}

	connectionMaxLifeTimeSec, err := internal.GetEnvAs("CONNECTION_MAX_LIFE_TIME_SEC", false, 10)
	if err != nil {
		errs = append(errs, err)
	}

	idpMetadataUrl, err := internal.GetEnvAs("IDP_METADATA_URL", true, "")
	fmt.Println("idpMetadataUrl: ", idpMetadataUrl)
	if err != nil {
		errs = append(errs, err)
	}

	sslCertPath, err := internal.GetEnvAs("SSL_CERT_PATH", true, "")
	if err != nil {
		errs = append(errs, err)
	}

	sslKeyPath, err := internal.GetEnvAs("SSL_KEY_PATH", true, "")
	if err != nil {
		errs = append(errs, err)
	}

	samlCertPath, err := internal.GetEnvAs("SAML_CERT_PATH", true, "")
	if err != nil {
		errs = append(errs, err)
	}

	samlKeyPath, err := internal.GetEnvAs("SAML_KEY_PATH", true, "")
	if err != nil {
		errs = append(errs, err)
	}

	rootUrl, err := internal.GetEnvAs("ROOT_URL", true, "")
	if err != nil {
		errs = append(errs, err)
	}

	entityID, err := internal.GetEnvAs("ENTITY_ID", true, "")
	if err != nil {
		errs = append(errs, err)
	}

	redirectUrlAfterLogin, err := internal.GetEnvAs("REDIRECT_URL_AFTER_LOGIN", true, "")
	if err != nil {
		errs = append(errs, err)
	}

	redirectUrlAfterLogout, err := internal.GetEnvAs("REDIRECT_URL_AFTER_LOGOUT", true, "")
	if err != nil {
		errs = append(errs, err)
	}

	accessTokenKey, err := internal.GetEnvAs("ACCESS_TOKEN_KEY", true, "")
	if err != nil {
		errs = append(errs, err)
	}

	refreshTokenKey, err := internal.GetEnvAs("REFRESH_TOKEN_KEY", true, "")
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
	}, nil
}
