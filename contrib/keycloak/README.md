# Keycloak as SAML provider


## Why ?

During initialization, the Vulcan API needs to connect to an authentication service that supports the SAML protocol, otherwise the API wont start.
This folder provides the necessary files to run a local instance of Keycloak as a SAML provider, so we can run the Vulcan API without having to connect to a remote/external service.

A default Vulcan user will be provisioned as follows:

```
Realm: Vulcan

user: vulcan
pass: vulcan
```


## Instructions

To start a Keycloar container, navigate to this folder and run the following command:
```
./start-keycloak.sh 
```

The admin console will be available at:
```
http://localhost:8083/auth/admin/

user: admin
pass: admin
```

## Vulcan API SAML config

In order to configure the Vulcan API to connect on Keycloak, configure the `saml` section from the `config` file like this:

```
[saml]
saml_metadata = "http://localhost:8083/auth/realms/vulcan/protocol/saml/descriptor"
saml_issuer = "http://localhost:8080/api/v1/login/callback"
saml_callback = "http://localhost:8080/api/v1/login/callback"
saml_trusted_domains = ["localhost"]
```

# Demo

See [Keycloak and Vulcan in action](keycloak-vulcan-demo.mp4)