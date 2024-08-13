package main

import (
	"errors"
	"fmt"
	"fxtester/internal/common"
	"fxtester/internal/keycloak"
	"strings"
)

func main() {
	param := keycloak.ClientParam{
		keycloak.KeyUser:    common.GetConfig().Saml.Keycloak.AdminUser.Username,
		keycloak.KeyPass:    common.GetConfig().Saml.Keycloak.AdminUser.Password,
		keycloak.KeyBaseURL: strings.TrimRight(common.GetConfig().Saml.Keycloak.BaseURL, "/"),
	}
	c := keycloak.NewClient(param)
	if err := c.Login(); err != nil {
		panic(fmt.Sprintf("ログインに失敗しました: %v", err))
	}

	realmName := common.GetConfig().Saml.Keycloak.RealmName

	// Realm作成前に古いRealmを削除する
	c.DeleteRealm(realmName)

	// Realmの作成
	if err := c.CreateRealm(realmName); err != nil {
		panic(fmt.Sprintf("Realmの作成に失敗しました: %v", err))
	}

	// クライアントの作成
	var req keycloak.ClientRepresentation
	req.Id = common.GetConfig().Saml.Keycloak.NewClientId
	req.ClientId = common.GetConfig().Saml.EntityId
	req.Protocol = keycloak.ProtocolSAML
	req.RedirectUris = []string{common.GetConfig().Saml.ValidRedirectURI}
	req.Attributes = map[string]string{
		keycloak.AttributeSamlClientSignature:         "false",
		keycloak.AttributeValidPostLogoutRedirectURIs: common.GetConfig().Saml.ValidPostLogoutRedirectURI,
		keycloak.AttributeLogoutServicePostBindingURL: common.GetConfig().Saml.LogoutServicePostBindingURL,
		keycloak.AttributeNameIdFormat:                "email",
	}
	if err := c.CreateClient(realmName, req); err != nil {
		panic(fmt.Sprintf("クライアントの作成に失敗しました: %v", err))
	}

	protocolMapper, scope, err := func() (*keycloak.ProtocolMapperRepresentation, *keycloak.ClientScopeRepresentation, error) {
		if scope, err := c.GetClientScope(realmName, "role_list"); err != nil {
			return nil, nil, err
		} else {
			for _, mp := range scope.ProtocolMappers {
				if mp.Name == "role list" {
					return &mp, scope, nil
				}
			}
		}
		return nil, nil, errors.New("no found role_list")
	}()
	if err != nil {
		panic(fmt.Sprintf("role_listのProtocolMapperの取得に失敗しました: %v", err))
	}

	// Single Role Attributeを有効にする
	protocolMapper.Config["single"] = true

	if err := c.UpdateProtocolMapper(realmName, scope.Id, *protocolMapper); err != nil {
		panic(fmt.Sprintf("Single Role Attributeの変更に失敗しました: %v", err))
	}

	for _, user := range common.GetConfig().Saml.Keycloak.NewUsers {
		if err := c.CreateUser(realmName, user.Username, user.Email, user.Password); err != nil {
			panic(fmt.Sprintf("%sユーザの作成に失敗しました: %v", user.Username, err))
		}
	}

}
