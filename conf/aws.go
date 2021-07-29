package conf

type AWSConfig struct {
	Region          string
	AccessKey       string
	AccessSecretKey string
	Bucket          string
}

func getAWSConfig() *AWSConfig {
	return &AWSConfig{
		Region:          c.GetString("aws.region"),
		AccessKey:       c.GetString("aws.access_key"),
		AccessSecretKey: c.GetString("aws.access_secret"),
		Bucket:          c.GetString("aws.bucket"),
	}
}
