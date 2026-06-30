module github.com/ncobase/ncore/oss/aliyun

go 1.25.3

replace github.com/ncobase/ncore/oss => ../

require (
	github.com/aliyun/alibabacloud-oss-go-sdk-v2 v1.4.0
	github.com/ncobase/ncore/oss v0.2.5
)

require (
	github.com/stretchr/testify v1.11.1 // indirect
	golang.org/x/time v0.15.0 // indirect
)
