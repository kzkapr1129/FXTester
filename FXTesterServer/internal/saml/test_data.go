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
