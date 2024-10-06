package saml

var TestDataIdpMetadata = `
<md:EntityDescriptor
	xmlns="urn:oasis:names:tc:SAML:2.0:metadata"
	xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata"
	xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion"
	xmlns:ds="http://www.w3.org/2000/09/xmldsig#" entityID="http://keycloak:8080/realms/my-realm">
	<script/>
	<md:IDPSSODescriptor WantAuthnRequestsSigned="true" protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
		<md:KeyDescriptor use="signing">
			<ds:KeyInfo>
				<ds:KeyName>4HSi_Vq42yK3KAVnIB8_dGfgVgMovHXgogVxZIw9eQc</ds:KeyName>
				<ds:X509Data>
					<ds:X509Certificate>MIICnzCCAYcCBgGSX4nv7zANBgkqhkiG9w0BAQsFADATMREwDwYDVQQDDAhteS1yZWFsbTAeFw0yNDEwMDYwMTUzNDhaFw0zNDEwMDYwMTU1MjhaMBMxETAPBgNVBAMMCG15LXJlYWxtMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzhUQJALhalNVybHN5oanq/2hWLFhbqnU2ESlM3GbuDbGdNbcuTx8aRFMALtdxFywttpNvZNj+aibMRJsfJeNZ5yOylMAkAMssTX831W1jpI6IZe11PcpLpxHgroDTjMn+hv83atnxHZQTaFIAzRmfP7vKMQBbEvej7AcsdNjuvJF0/DLwRUt5JPAQbCalisSR3a1g6THjZyEdym/HUaLTNsjcC3yvIBr1A54AElyDVinCiqOKfy8CaU5EEwn2qCjyjbVN7f3Zz05Orqsv479hKqEP10nXVjhkfOOX7nK2gOBQzkUPt1X7lgEaYJQuzeLMrSviA839OPzYno+JXDYdQIDAQABMA0GCSqGSIb3DQEBCwUAA4IBAQBhlsRPpLKuhu1ZcKpOJjV6oC+A3bdTzanemJ1cd8NkObLTgxX0k6VwO5sdigEngKS5Jvr5A+Q2o7CxRXmMI56IwRL5r+Q1/YBlwBWt+AjDoR5RXdRWwKF5+RUMizhyeQrZT2KNaWL4gj+unMbQOfft95v2rTDlhugNMpZ93HfO3nxk24qRkGpteZmXpTfriuJSTt47wP2ZzZD7bV3q50BwfE//Jjo++nX5Qpw+ci7u5UVZU6wFXi3K9zVD/yfT6y7bU6BpD8OXhv1c8Ss62l5cKUfGo3tZioOJ7hblNdE4dFP6W0V8FFn5sQZzxZC/EGT1eUI2HluRcvhGaKSSfkRy</ds:X509Certificate>
				</ds:X509Data>
			</ds:KeyInfo>
		</md:KeyDescriptor>
		<md:ArtifactResolutionService Binding="urn:oasis:names:tc:SAML:2.0:bindings:SOAP" Location="http://keycloak:8080/realms/my-realm/protocol/saml/resolve" index="0"/>
		<md:SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="http://keycloak:8080/realms/my-realm/protocol/saml"/>
		<md:SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="http://keycloak:8080/realms/my-realm/protocol/saml"/>
		<md:SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Artifact" Location="http://keycloak:8080/realms/my-realm/protocol/saml"/>
		<md:SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:SOAP" Location="http://keycloak:8080/realms/my-realm/protocol/saml"/>
		<md:NameIDFormat>urn:oasis:names:tc:SAML:2.0:nameid-format:persistent</md:NameIDFormat>
		<md:NameIDFormat>urn:oasis:names:tc:SAML:2.0:nameid-format:transient</md:NameIDFormat>
		<md:NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified</md:NameIDFormat>
		<md:NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress</md:NameIDFormat>
		<md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="http://keycloak:8080/realms/my-realm/protocol/saml"/>
		<md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="http://keycloak:8080/realms/my-realm/protocol/saml"/>
		<md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:SOAP" Location="http://keycloak:8080/realms/my-realm/protocol/saml"/>
		<md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Artifact" Location="http://keycloak:8080/realms/my-realm/protocol/saml"/>
	</md:IDPSSODescriptor>
</md:EntityDescriptor>
`
var TestDataLogoutResponse = `
<samlp:LogoutResponse xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol"
                      xmlns="urn:oasis:names:tc:SAML:2.0:assertion"
                      xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion"
                      Destination="https://fx-tester-be:8000/saml/slo"
                      ID="ID_17bf1cac-f9bc-4da8-a27c-9646c9f0d6ae"
                      InResponseTo="test-authn-request-id"
                      IssueInstant="2024-10-06T12:51:57.429Z"
                      Version="2.0"
                      >
    <Issuer>http://keycloak:8080/realms/my-realm</Issuer>
    <dsig:Signature xmlns:dsig="http://www.w3.org/2000/09/xmldsig#">
        <dsig:SignedInfo>
            <dsig:CanonicalizationMethod Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#" />
            <dsig:SignatureMethod Algorithm="http://www.w3.org/2001/04/xmldsig-more#rsa-sha256" />
            <dsig:Reference URI="#ID_17bf1cac-f9bc-4da8-a27c-9646c9f0d6ae">
                <dsig:Transforms>
                    <dsig:Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature" />
                    <dsig:Transform Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#" />
                </dsig:Transforms>
                <dsig:DigestMethod Algorithm="http://www.w3.org/2001/04/xmlenc#sha256" />
                <dsig:DigestValue>H7/QgdokKORMTlgHhQk3qpVZnFoIl/EDt+r8b7A3/CU=</dsig:DigestValue>
            </dsig:Reference>
        </dsig:SignedInfo>
        <dsig:SignatureValue>E6oVE2OHwmfPJgi36SQOpiNz9WF6ZyKf986FDSfWE4P1qjPUTqgV2B3Mf4I/uIhzf/zr+yCBKlVv4sSSO+a8v6tKobBeZjQL1jzT7afNWShW8a9zQK3mINj08kJ050MBct9UJZQjYK6pX7cIQN0PWt5lQRWx3kjjXrAbBJstZ9wACffhZ4O5tgk+OfM5dnmvtR2st20SbMTjxH3Uuult0CYuLxChfA1T4JJGWgbmfIPl9UW/oQpoR00g5s625eLNHMkVcGmhbU3rqyAOnaWo8hyQvOFMjHYbwLuKv6o3M61Sqtn1rVeJlgzBFqGMJDStgYefK9l1ZvZBfH10/7qG/w==</dsig:SignatureValue>
        <dsig:KeyInfo>
            <dsig:KeyName>E0jR9QYbelf5qujtYmCBGzqkgO5chvLrEl1JtI1Te54</dsig:KeyName>
            <dsig:X509Data>
                <dsig:X509Certificate>MIICnzCCAYcCBgGSYbrNhjANBgkqhkiG9w0BAQsFADATMREwDwYDVQQDDAhteS1yZWFsbTAeFw0yNDEwMDYxMjA2MjVaFw0zNDEwMDYxMjA4MDVaMBMxETAPBgNVBAMMCG15LXJlYWxtMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAo7L2jF6JNX05F1DrCCRA9ORwOLAeKPQeUfgOohQJoW4F+UNKp9p8EUGfqnHlIdFR2OGcJyzsUgBmvlGVWAPJlIMqf5Ao6CERhu3k6pRxsmHY4ewSTrWKP5V9wy422Q1uPUR5FdWCP7anWcziRQSr7+eJSoCl0Pjmi3QW2FlvWp/iJU2TjdsEJntHzBQjdm8DtaNb1d7+RZ+WRfjcTqADGIJH+aKD0457RBzJo1rc8ShyeQsZinGRV048xOPSJZ5ki2rb8k6AqBIp7sOsW0JWRA4cxtDXi1RdfLkSTIsimKVLyUTVtm16TuAcxyrxxWh4BmFkFXzBF1JRsrsUh7B7CQIDAQABMA0GCSqGSIb3DQEBCwUAA4IBAQAVHFN5Vm3lDV2g3LPeRfEYJcYg00Lyh9DrlYv1GiwcwTDRI7QvL9GoMf0mATQMsZlQeH+1zuzyGFKkDD7pqLVNGQt3tvDwf4V6OlZMqSM5b7n1reFAidbCIlRAdSL8KYl2Y6useEZb4T/gjXp1z+6/5if2IzZj4aIoYkfTZ2YcvieMmsBUmG/uabqhMT/Wk5r8qpUXyLNl2PhLD7AJOYJm87HoxEI185JkR4awpF1NAJs+zZHODe+Mj1fq9IofxLB8F3ko+g1ywltpOXF8bIONOHsM2GST87NYuRYZfKVn6zBKC9ku+apSJav5JX6QzO2be2SKB2vQ7PS7kpa5AMPU</dsig:X509Certificate>
            </dsig:X509Data>
        </dsig:KeyInfo>
    </dsig:Signature>
    <samlp:Status>
        <samlp:StatusCode Value="urn:oasis:names:tc:SAML:2.0:status:Success" />
    </samlp:Status>
</samlp:LogoutResponse>`
