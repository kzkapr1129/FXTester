package keycloak

type PolicyEnforcementMode string
type DecisionStrategy string
type Logic string

type CreateRealmRequest struct {
	RealmName string `json:"realm"`
	Enabled   bool   `json:"enabled"`
}

type ClientRepresentation struct {
	Id                                    string                         `json:"id,omitempty"`
	ClientId                              string                         `json:"clientId,omitempty"`
	Name                                  string                         `json:"name,omitempty"`
	Description                           string                         `json:"description,omitempty"`
	RootURL                               string                         `json:"rootUrl,omitempty"`
	AdminUrl                              string                         `json:"adminUrl,omitempty"`
	BaseUrl                               string                         `json:"baseUrl,omitempty"`
	SurrogateAuthRequired                 bool                           `json:"surrogateAuthRequired,omitempty"`
	Enabled                               bool                           `json:"enabled,omitempty"`
	AlwaysDisplayInConsole                bool                           `json:"alwaysDisplayInConsole,omitempty"`
	ClientAuthenticatorType               string                         `json:"clientAuthenticatorType,omitempty"`
	Secret                                string                         `json:"secret,omitempty"`
	RegistrationAccessToken               string                         `json:"registrationAccessToken,omitempty"`
	DefaultRoles                          []string                       `json:"defaultRoles,omitempty"`
	RedirectUris                          []string                       `json:"redirectUris,omitempty"`
	WebOrigins                            []string                       `json:"webOrigins,omitempty"`
	NotBefore                             int64                          `json:"notBefore,omitempty"`
	BearerOnly                            bool                           `json:"bearerOnly,omitempty"`
	ConsentRequired                       bool                           `json:"consentRequired,omitempty"`
	StandardFlowEnabled                   bool                           `json:"standardFlowEnabled,omitempty"`
	ImplicitFlowEnabled                   bool                           `json:"implicitFlowEnabled,omitempty"`
	DirectAccessGrantsEnabled             bool                           `json:"directAccessGrantsEnabled,omitempty"`
	ServiceAccountsEnabled                bool                           `json:"serviceAccountsEnabled,omitempty"`
	Oauth2DeviceAuthorizationGrantEnabled bool                           `json:"oauth2DeviceAuthorizationGrantEnabled,omitempty"`
	AuthorizationServicesEnabled          bool                           `json:"authorizationServicesEnabled,omitempty"`
	DirectGrantsOnly                      bool                           `json:"directGrantsOnly,omitempty"`
	PublicClient                          bool                           `json:"publicClient,omitempty"`
	FrontchannelLogout                    bool                           `json:"frontchannelLogout,omitempty"`
	Protocol                              string                         `json:"protocol,omitempty"`
	Attributes                            map[string]string              `json:"attributes,omitempty"`
	AuthenticationFlowBindingOverrides    map[string]string              `json:"authenticationFlowBindingOverrides,omitempty"`
	FullScopeAllowed                      bool                           `json:"fullScopeAllowed,omitempty"`
	NodeReRegistrationTimeout             int64                          `json:"nodeReRegistrationTimeout,omitempty"`
	RegisteredNodes                       map[string]int64               `json:"registeredNodes,omitempty"`
	ProtocolMappers                       []ProtocolMapperRepresentation `json:"protocolMappers,omitempty"`
	ClientTemplate                        string                         `json:"clientTemplate,omitempty"`
	UseTemplateConfig                     bool                           `json:"useTemplateConfig,omitempty"`
	UseTemplateScope                      bool                           `json:"useTemplateScope,omitempty"`
	UseTemplateMappers                    bool                           `json:"useTemplateMappers,omitempty"`
	DefaultClientScopes                   []string                       `json:"defaultClientScopes,omitempty"`
	OptionalClientScopes                  []string                       `json:"optionalClientScopes,omitempty"`
}

type ProtocolMapperRepresentation struct {
	Id              string                 `json:"id,omitempty"`
	Name            string                 `json:"name,omitempty"`
	Protocol        string                 `json:"protocol,omitempty"`
	ProtocolMapper  string                 `json:"protocolMapper,omitempty"`
	ConsentRequired bool                   `json:"consentRequired,omitempty"`
	ConsentText     string                 `json:"consentText,omitempty"`
	Config          map[string]interface{} `json:"config,omitempty"`
}

type ResourceServerRepresentation struct {
	Id                            string                   `json:"id,omitempty"`
	ClientId                      string                   `json:"clientId,omitempty"`
	Name                          string                   `json:"name,omitempty"`
	AllowRemoteResourceManagement bool                     `json:"allowRemoteResourceManagement,omitempty"`
	PolicyEnforcementMode         PolicyEnforcementMode    `json:"policyEnforcementMode,omitempty"`
	Resources                     []ResourceRepresentation `json:"resources,omitempty"`
	Policies                      []PolicyRepresentation   `json:"policies,omitempty"`
	Scopes                        []ScopeRepresentation    `json:"scopes,omitempty"`
	DecisionStrategy              DecisionStrategy         `json:"decisionStrategy,omitempty"`
}

type ResourceRepresentation struct {
	Id                 string                      `json:"id,omitempty"`
	Name               string                      `json:"name,omitempty"`
	Uris               []string                    `json:"uris,omitempty"`
	Type               string                      `json:"type,omitempty"`
	Scopes             []ScopeRepresentation       `json:"scopes,omitempty"`
	IconUri            string                      `json:"icon_uri,omitempty"`
	Owner              ResourceRepresentationOwner `json:"owner,omitempty"`
	OwnerManagedAccess bool                        `json:"ownerManagedAccess,omitempty"`
	DisplayName        string                      `json:"displayName,omitempty"`
	Attributes         map[string][]string         `json:"attributes,omitempty"`
	Uri                string                      `json:"uri,omitempty"`
	ScopesUma          []ScopeRepresentation       `json:"scopesUma,omitempty"`
}

type ResourceRepresentationOwner struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type PolicyRepresentation struct {
	Id               string                   `json:"id,omitempty"`
	Name             string                   `json:"name,omitempty"`
	Description      string                   `json:"description,omitempty"`
	Type             string                   `json:"type,omitempty"`
	Policies         []string                 `json:"policies,omitempty"`
	Resources        []string                 `json:"resources,omitempty"`
	Scopes           []string                 `json:"scopes,omitempty"`
	Logic            Logic                    `json:"logic,omitempty"`
	DecisionStrategy DecisionStrategy         `json:"decisionStrategy,omitempty"`
	Owner            string                   `json:"owner,omitempty"`
	ResourcesData    []ResourceRepresentation `json:"resourcesData,omitempty"`
	ScopesData       []ScopeRepresentation    `json:"scopesData,omitempty"`
	Config           map[string]string        `json:"config,omitempty"`
}

type ScopeRepresentation struct {
	Id          string                   `json:"id,omitempty"`
	Name        string                   `json:"name,omitempty"`
	IconUri     string                   `json:"iconUri,omitempty"`
	Policies    []PolicyRepresentation   `json:"policies,omitempty"`
	Resources   []ResourceRepresentation `json:"resources,omitempty"`
	DisplayName string                   `json:"displayName,omitempty"`
}

type ClientScopeRepresentation struct {
	Id              string                         `json:"id,omitempty"`
	Name            string                         `json:"name,omitempty"`
	Description     string                         `json:"description,omitempty"`
	Protocol        string                         `json:"protocol,omitempty"`
	Attributes      map[string]interface{}         `json:"attributes,omitempty"`
	ProtocolMappers []ProtocolMapperRepresentation `json:"protocolMappers,omitempty"`
}

type UserRepresentation struct {
	Id                         string                            `json:"id,omitempty"`
	Username                   string                            `json:"username,omitempty"`
	FirstName                  string                            `json:"firstName,omitempty"`
	LastName                   string                            `json:"lastName,omitempty"`
	Email                      string                            `json:"email,omitempty"`
	EmailVerified              bool                              `json:"emailVerified,omitempty"`
	Attributes                 map[string][]interface{}          `json:"attributes,omitempty"`
	UserProfileMetadata        UserProfileMetadata               `json:"userProfileMetadata,omitempty"`
	Self                       string                            `json:"self,omitempty"`
	Origin                     string                            `json:"origin,omitempty"`
	CreatedTimestamp           int64                             `json:"createdTimestamp,omitempty"`
	Enabled                    bool                              `json:"enabled,omitempty"`
	Totp                       bool                              `json:"totp,omitempty"`
	FederationLink             string                            `json:"federationLink,omitempty"`
	ServiceAccountClientId     string                            `json:"serviceAccountClientId,omitempty"`
	Credentials                []CredentialRepresentation        `json:"credentials,omitempty"`
	DisableableCredentialTypes []string                          `json:"disableableCredentialTypes,omitempty"`
	RequiredActions            []string                          `json:"requiredActions,omitempty"`
	FederatedIdentities        []FederatedIdentityRepresentation `json:"federatedIdentities,omitempty"`
	RealmRoles                 []string                          `json:"realmRoles,omitempty"`
	ClientRoles                map[string]interface{}            `json:"clientRoles,omitempty"`
	ClientConsents             []UserConsentRepresentation       `json:"clientConsents,omitempty"`
	NotBefore                  int32                             `json:"notBefore,omitempty"`
	ApplicationRoles           map[string][]interface{}          `json:"applicationRoles,omitempty"`
	SocialLinks                []SocialLinkRepresentation        `json:"socialLinks,omitempty"`
	Groups                     []string                          `json:"groups,omitempty"`
	Access                     map[string]bool                   `json:"access,omitempty"`
}

type UserProfileMetadata struct {
	Attributes []UserProfileAttributeMetadata      `json:"attributes,omitempty"`
	Groups     []UserProfileAttributeGroupMetadata `json:"groups,omitempty"`
}

type UserProfileAttributeGroupMetadata struct {
	Name               string                 `json:"name,omitempty"`
	DisplayHeader      string                 `json:"displayHeader,omitempty"`
	DisplayDescription string                 `json:"displayDescription,omitempty"`
	Annotations        map[string]interface{} `json:"annotations,omitempty"`
}

type UserProfileAttributeMetadata struct {
	Name        string                            `json:"name,omitempty"`
	DisplayName string                            `json:"displayName,omitempty"`
	Required    bool                              `json:"required,omitempty"`
	ReadOnly    bool                              `json:"readOnly,omitempty"`
	Annotations map[string]interface{}            `json:"annotations,omitempty"`
	Validators  map[string]map[string]interface{} `json:"validators,omitempty"`
	Group       string                            `json:"group,omitempty"`
	Multivalued bool                              `json:"multivalued,omitempty"`
}

type CredentialRepresentation struct {
	Id                string                 `json:"id,omitempty"`
	Type              string                 `json:"type,omitempty"`
	UserLabel         string                 `json:"userLabel,omitempty"`
	CreatedDate       int64                  `json:"createdDate,omitempty"`
	SecretData        string                 `json:"secretData,omitempty"`
	CredentialData    string                 `json:"credentialData,omitempty"`
	Priority          int32                  `json:"priority,omitempty"`
	Value             string                 `json:"value,omitempty"`
	Temporary         bool                   `json:"temporary,omitempty"`
	Device            string                 `json:"device,omitempty"`
	HashedSaltedValue string                 `json:"hashedSaltedValue,omitempty"`
	Salt              string                 `json:"salt,omitempty"`
	HashIterations    int32                  `json:"hashIterations,omitempty"`
	Counter           int32                  `json:"counter,omitempty"`
	Algorithm         string                 `json:"algorithm,omitempty"`
	Digits            int32                  `json:"digits,omitempty"`
	Period            int32                  `json:"period,omitempty"`
	Config            map[string]interface{} `json:"config,omitempty"`
}

type FederatedIdentityRepresentation struct {
	IdentityProvider string `json:"identityProvider,omitempty"`
	UserId           string `json:"userId,omitempty"`
	UserName         string `json:"userName,omitempty"`
}

type UserConsentRepresentation struct {
	ClientId            string   `json:"userName,omitempty"`
	GrantedClientScopes []string `json:"grantedClientScopes,omitempty"`
	CreatedDate         int64    `json:"createdDate,omitempty"`
	LastUpdatedDate     int64    `json:"lastUpdatedDate,omitempty"`
	GrantedRealmRoles   []string `json:"grantedRealmRoles,omitempty"`
}

type SocialLinkRepresentation struct {
	SocialProvider string `json:"socialProvider,omitempty"`
	SocialUserId   string `json:"socialUserId,omitempty"`
	SocialUsername string `json:"socialUsername,omitempty"`
}
