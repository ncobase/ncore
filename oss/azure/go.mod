module github.com/ncobase/ncore/oss/azure

go 1.25.3

replace github.com/ncobase/ncore/oss => ../

require (
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.6.4
	github.com/ncobase/ncore/oss v0.2.5
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.22.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.12.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	golang.org/x/net v0.56.0 // indirect
	golang.org/x/text v0.38.0 // indirect
)
