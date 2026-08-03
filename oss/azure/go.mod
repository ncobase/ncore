module github.com/ncobase/ncore/oss/azure

go 1.25.3

replace github.com/ncobase/ncore/oss => ../

require (
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.6.4
	github.com/ncobase/ncore/oss v0.2.6
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.22.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.12.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	golang.org/x/net v0.56.0 // indirect
	golang.org/x/text v0.39.0 // indirect
)
