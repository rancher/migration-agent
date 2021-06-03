package migrate

type MigrationConfig struct {
	KubeConfig          string
	DataDir             string
	Snapshot            string
	EtcdS3Endpoint      string
	EtcdS3EndpointCA    string
	EtcdS3SkipSSLVerify bool
	EtcdS3AccessKey     string
	EtcdS3SecretKey     string
	EtcdS3Region        string
	EtcdS3BucketName    string
	EtcdS3Folder        string
	NodeName			string
}
