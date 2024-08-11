package keycloak

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type ClientParamKey string
type ClientParam map[ClientParamKey]string

const (
	KeyBaseURL ClientParamKey = "base_url"
	KeyUser    ClientParamKey = "user"
	KeyPass    ClientParamKey = "pass"
)

const (
	ProtocolSAML = "saml"
)

const (
	AttributeSamlClientSignature         = "saml.client.signature"
	AttributeValidPostLogoutRedirectURIs = "post.logout.redirect.uris"
	AttributeLogoutServicePostBindingURL = "saml_single_logout_service_url_post"
	AttributeNameIdFormat                = "saml_name_id_format"
)

type client struct {
	param ClientParam

	data struct {
		accessToken string
	}
}

func NewClient(param ClientParam) *client {
	return &client{
		param: param,
	}
}

func (c *client) Login() error {
	data := url.Values{}
	data.Set("username", c.param[KeyUser]) // admin
	data.Set("password", c.param[KeyPass]) // admin
	data.Set("grant_type", "password")
	data.Set("client_id", "admin-cli")

	// リクエストURLの作成
	postRes, err := c.doHttpRequest(struct {
		requestURL      string
		method          string
		body            string
		header          map[string]string
		withAccessToken bool
	}{
		requestURL: fmt.Sprintf("%s/realms/master/protocol/openid-connect/token", c.param[KeyBaseURL]),
		method:     "POST",
		body:       data.Encode(),
		header: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		withAccessToken: false,
	})
	if err != nil {
		return err
	}

	var res map[string]interface{}
	if err := json.Unmarshal(postRes.resBody, &res); err != nil {
		return err
	}

	// 受信したレスポンスを保存
	if v, ok := res["access_token"]; !ok {
		return fmt.Errorf("failed to get access_token")
	} else if accessToken, ok := v.(string); !ok {
		return fmt.Errorf("invalid type of access_token")
	} else {
		c.data.accessToken = accessToken
	}
	return nil
}

func (c *client) CreateRealm(realmName string) error {
	// リクエストボディの作成
	reqBody, err := func() (string, error) {
		var req CreateRealmRequest
		req.RealmName = realmName
		req.Enabled = true
		bytes, err := json.Marshal(req)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	}()
	if err != nil {
		return err
	}

	// リクエストURLの作成
	res, err := c.doHttpRequest(struct {
		requestURL      string
		method          string
		body            string
		header          map[string]string
		withAccessToken bool
	}{
		requestURL: fmt.Sprintf("%s/admin/realms/", c.param[KeyBaseURL]),
		method:     "POST",
		body:       reqBody,
		header: map[string]string{
			"Content-Type": "application/json",
		},
		withAccessToken: true,
	})
	if err != nil {
		return err
	}

	if res.status != 201 {
		return fmt.Errorf("invalid response: %d", res.status)
	}

	return nil
}

func (c *client) DeleteRealm(realmName string) error {
	// リクエストURLの作成
	deleteRes, err := c.doHttpRequest(struct {
		requestURL      string
		method          string
		body            string
		header          map[string]string
		withAccessToken bool
	}{
		requestURL:      fmt.Sprintf("%s/admin/realms/%s", c.param[KeyBaseURL], realmName),
		method:          "DELETE",
		withAccessToken: true,
	})
	if err != nil {
		return err
	}
	if deleteRes.status != 204 {
		return fmt.Errorf("invalid response: %d", deleteRes.status)
	}
	return nil
}

func (c *client) CreateClient(realmName string, body ClientRepresentation) error {
	// リクエストボディの作成
	reqBody, err := func() (string, error) {
		bytes, err := json.Marshal(body)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	}()
	if err != nil {
		return err
	}

	// リクエストURLの作成
	postRes, err := c.doHttpRequest(struct {
		requestURL      string
		method          string
		body            string
		header          map[string]string
		withAccessToken bool
	}{
		requestURL: fmt.Sprintf("%s/admin/realms/%s/clients", c.param[KeyBaseURL], realmName),
		method:     "POST",
		body:       reqBody,
		header: map[string]string{
			"Content-Type": "application/json",
		},
		withAccessToken: true,
	})
	if err != nil {
		return err
	}

	if postRes.status != 201 {
		return fmt.Errorf("invalid response: %d", postRes.status)
	}

	return nil
}

func (c *client) GetClient(realmName string) error {
	postRes, err := c.doHttpRequest(struct {
		requestURL      string
		method          string
		body            string
		header          map[string]string
		withAccessToken bool
	}{
		requestURL: fmt.Sprintf("%s/admin/realms/%s/clients", c.param[KeyBaseURL], realmName),
		method:     "GET",
		header: map[string]string{
			"Content-Type": "application/json",
		},
		withAccessToken: true,
	})
	if err != nil {
		return err
	}

	fmt.Println(string(postRes.resBody))

	return nil
}

func (c *client) DeleteClient(realmName string, id string) error {
	// リクエストURLの作成
	deleteRes, err := c.doHttpRequest(struct {
		requestURL      string
		method          string
		body            string
		header          map[string]string
		withAccessToken bool
	}{
		requestURL:      fmt.Sprintf("%s/admin/realms/%s/clients/%s", c.param[KeyBaseURL], realmName, id),
		method:          "DELETE",
		withAccessToken: true,
	})
	if err != nil {
		return err
	}

	if deleteRes.status != 204 {
		return fmt.Errorf("invalid response: %d", deleteRes.status)
	}
	return nil
}

func (c *client) GetClientScope(realm string, scopeName string) (*ClientScopeRepresentation, error) {
	res, err := c.doHttpRequest(struct {
		requestURL      string
		method          string
		body            string
		header          map[string]string
		withAccessToken bool
	}{
		requestURL:      fmt.Sprintf("%s/admin/realms/%s/client-scopes", c.param[KeyBaseURL], realm),
		method:          "GET",
		withAccessToken: true,
	})
	if err != nil {
		return nil, err
	}

	if res.status != 200 {
		return nil, fmt.Errorf("invalid response: %d", res.status)
	}

	var m []ClientScopeRepresentation
	json.Unmarshal(res.resBody, &m)
	for _, v := range m {
		if v.Name == scopeName {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("no found: %s:%s", realm, scopeName)
}

func (c *client) UpdateProtocolMapper(realm string, clientScopeId string, protocolMapper ProtocolMapperRepresentation) error {
	// リクエストボディの作成
	reqBody, err := func() (string, error) {
		bytes, err := json.Marshal(protocolMapper)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	}()
	if err != nil {
		return err
	}

	res, err := c.doHttpRequest(struct {
		requestURL      string
		method          string
		body            string
		header          map[string]string
		withAccessToken bool
	}{
		requestURL:      fmt.Sprintf("%s/admin/realms/%s/client-scopes/%s/protocol-mappers/models/%s", c.param[KeyBaseURL], realm, clientScopeId, protocolMapper.Id),
		method:          "PUT",
		body:            reqBody,
		withAccessToken: true,
	})
	if err != nil {
		return err
	}

	if res.status != 204 {
		return fmt.Errorf("invalid response: %d", res.status)
	}

	return nil
}

func (c *client) CreateUser(realm string, username string, email string, password string) error {
	user := UserRepresentation{
		Email:         email,
		Username:      username,
		Enabled:       true,
		EmailVerified: false,
		Credentials: []CredentialRepresentation{
			{
				UserLabel: "MyPassword",
				Type:      "password",
				Value:     password,
				Temporary: false,
			},
		},
	}

	// リクエストボディの作成
	reqBody, err := func() (string, error) {
		bytes, err := json.Marshal(user)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	}()
	if err != nil {
		return err
	}

	res, err := c.doHttpRequest(struct {
		requestURL      string
		method          string
		body            string
		header          map[string]string
		withAccessToken bool
	}{
		requestURL:      fmt.Sprintf("%s/admin/realms/%s/users", c.param[KeyBaseURL], realm),
		method:          "POST",
		body:            reqBody,
		withAccessToken: true,
	})
	if err != nil {
		return err
	}

	if res.status != 201 {
		return fmt.Errorf("invalid response: %d", res.status)
	}

	return nil
}

func (c *client) doHttpRequest(param struct {
	requestURL      string
	method          string
	body            string
	header          map[string]string
	withAccessToken bool
}) (*struct {
	resBody []byte
	status  int
}, error) {
	req, err := http.NewRequest(strings.ToUpper(param.method), param.requestURL, strings.NewReader(param.body))
	if err != nil {
		return nil, err
	}

	// HTTPヘッダの設定
	for key, value := range param.header {
		req.Header.Set(key, value)
	}

	if param.withAccessToken {
		req.Header.Set("Authorization", "Bearer "+c.data.accessToken)
	}

	client := &http.Client{}

	// httpリクエストの送信/レスポンスの受信
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// レスポンスボディを読み取る
	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &struct {
		resBody []byte
		status  int
	}{
		resBody: res,
		status:  resp.StatusCode,
	}, nil
}
