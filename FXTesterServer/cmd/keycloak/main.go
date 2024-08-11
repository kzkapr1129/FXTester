package main

import (
	"fmt"
	"fxtester/internal"
	"fxtester/internal/keycloak"
)

func main() {
	param := keycloak.ClientParam{
		keycloak.KeyUser:    "admin",
		keycloak.KeyPass:    "admin",
		keycloak.KeyBaseURL: "http://localhost:28080",
	}
	c := keycloak.NewClient(param)
	if err := c.Login(); err != nil {
		panic("ログインに失敗しました")
	}

	c.DeleteRealm("fxtester")
	if err := c.CreateRealm("fxtester"); err != nil {
		panic("Realmの作成に失敗しました")
	}

	var req keycloak.ClientRepresentation
	req.Id = "test-client"
	req.ClientId = internal.GetConfig().Saml.EntityId
	req.Protocol = keycloak.ProtocolSAML
	req.RedirectUris = []string{"http://localhost:1234/test2"}
	req.Attributes = map[string]string{
		keycloak.AttributeSamlClientSignature:         "false",
		keycloak.AttributeValidPostLogoutRedirectURIs: "http://localhost:1234/test",
		keycloak.AttributeLogoutServicePostBindingURL: "https://example.com/logout",
		keycloak.AttributeNameIdFormat:                "email",
	}
	fmt.Println("client", c.CreateClient("fxtester", req))

	if scope, err := c.GetClientScope("fxtester", "role_list"); err != nil {
		panic("failed to get role_list")
	} else {
		for _, mp := range scope.ProtocolMappers {
			if mp.Name == "role list" {
				mp.Config["single"] = true
				fmt.Println("UpdateProtocolMapper: ", c.UpdateProtocolMapper("fxtester", scope.Id, mp))
				break
			}
		}
	}

	c.GetClient("fxtester")

	c.CreateUser("master", "test", "test")
}
