package resolved_config

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/config_version"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/overrides_objects/v3"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/logs_aggregator_functions/implementations/vector"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	// public so it can be used as default in CLI engine manager
	DefaultDockerClusterName = "docker"

	defaultMinikubeClusterName = "minikube"

	defaultMinikubeClusterKubernetesClusterNameStr = "minikube"
	defaultMinikubeStorageClass                    = "standard"
	defaultMinikubeEnclaveDataVolumeMB             = uint(10)
	defaultLogsAggregatorImage                     = "timberio/vector:0.31.0-debian"
	DefaultCloudConfigApiUrl                       = "cloud.kurtosis.com"
	DefaultCloudConfigPort                         = uint(8080)
	// TODO: We'll need to pull this more dynamic. For now placing here:
	//  Certificate chain obtained by running: openssl s_client -connect cloud.kurtosis.com:8080 -showcerts
	DefaultCertificateChain = "-----BEGIN CERTIFICATE-----\nMIIF0TCCBLmgAwIBAgIQDyigPWbHPvH8PY0tWs+GfzANBgkqhkiG9w0BAQsFADA8\nMQswCQYDVQQGEwJVUzEPMA0GA1UEChMGQW1hem9uMRwwGgYDVQQDExNBbWF6b24g\nUlNBIDIwNDggTTAxMB4XDTIzMDcyMTAwMDAwMFoXDTI0MDgxODIzNTk1OVowHTEb\nMBkGA1UEAxMSY2xvdWQua3VydG9zaXMuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOC\nAQ8AMIIBCgKCAQEAm5pEA+3RLt32aCSorHdiLUVRGJ5lAWBVUmS/5QBDNs6oYPYe\nV2oaHwgb0CxVcjhe+OzYeukJOY9g7uKsLAbTMtoKrrqqm8FuOnr1FKWV3/aopGCA\nKkUwQFf24oSeEDoA9SzLlJolVHWxOMiwPgq0LMg7vmIGmGCeXW6IOWQ6t5DLz9Mg\naUIunrRt9CsiMp9fEJzip4RkGfQL9t/B3Y5dtctNW/NHhmn0hFwdKM6NFetzR8JU\nmywfDTBlhkVy6PcGklIJCtbB02VifcnwYLkmlG4dddCzR6whn06h4KYcbIRtAAhs\nCUnVbi+8jn2OqvKSWJ0RTnNQ45wIVu2GBnbgoQIDAQABo4IC7DCCAugwHwYDVR0j\nBBgwFoAUgbgOY4qJEhjl+js7UJWf5uWQE4UwHQYDVR0OBBYEFOrSXY4CXs9tNuMg\nksZe0C3z83OtMB0GA1UdEQQWMBSCEmNsb3VkLmt1cnRvc2lzLmNvbTAOBgNVHQ8B\nAf8EBAMCBaAwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMDsGA1UdHwQ0\nMDIwMKAuoCyGKmh0dHA6Ly9jcmwucjJtMDEuYW1hem9udHJ1c3QuY29tL3IybTAx\nLmNybDATBgNVHSAEDDAKMAgGBmeBDAECATB1BggrBgEFBQcBAQRpMGcwLQYIKwYB\nBQUHMAGGIWh0dHA6Ly9vY3NwLnIybTAxLmFtYXpvbnRydXN0LmNvbTA2BggrBgEF\nBQcwAoYqaHR0cDovL2NydC5yMm0wMS5hbWF6b250cnVzdC5jb20vcjJtMDEuY2Vy\nMAwGA1UdEwEB/wQCMAAwggF/BgorBgEEAdZ5AgQCBIIBbwSCAWsBaQB3AO7N0GTV\n2xrOxVy3nbTNE6Iyh0Z8vOzew1FIWUZxH7WbAAABiXj17bMAAAQDAEgwRgIhAMRo\nVj0REFx0sDfWWgLLGr74Vb3ZFIG4UP2e3RnFJvzYAiEAmiI6Yn8IUBFDK0XVzSXu\nEOOk1lG1P6Joa1to8z9u7t4AdwBIsONr2qZHNA/lagL6nTDrHFIBy1bdLIHZu7+r\nOdiEcwAAAYl49e2uAAAEAwBIMEYCIQC+A2CnA4MPZJkoQev4Sh97dmozlPGNZIOD\nSvCANNx+/wIhAOk5geC6d42rDwE8hclRiGwIlXYacLGHqPKPEWgvQHLKAHUA2ra/\naz+1tiKfm8K7XGvocJFxbLtRhIU0vaQ9MEjX+6sAAAGJePXthQAABAMARjBEAiBD\ngHWN1z3GQBEZb7UAccg1tLEHGHwZTeMvAC+JJZHzigIgZOIagJoMAWCD+n7IfHWR\nCAdI6Z5FF7GFsIJwd0/ytgMwDQYJKoZIhvcNAQELBQADggEBACjM3hpxhf10xU6q\nDFJ6r8ayq/C02fRss+gF1hFTl3aJOngIQenHocb0xqTqaOKsm68MpxVI0fIXTWGe\nwYTpOIYXekHcftCJrgE8b3+kTtRp9cihnalq1MrkchWuN8eGZ4kgjCl9MYKV+7/u\nYG8Kzg4OxPwhEcYUgmPavhG2+K6RjyB1rR2KtEp7kI8Nn5UmI86Sty0PWY9+xaVw\nmvs1l/K58Y+kW/hJXnY93UWckQn3qV5nU/dA0zJkj63+JaZ2+MVeo1VHonjufLvX\nBT1NfrF+vGDF7ULMkPbSrLzMlbl6ULYqIEARJQHr2BouJuNScp9z3vZXHCiqkjaY\nGiZS750=\n-----END CERTIFICATE-----\n-----BEGIN CERTIFICATE-----\nMIIEXjCCA0agAwIBAgITB3MSOAudZoijOx7Zv5zNpo4ODzANBgkqhkiG9w0BAQsF\nADA5MQswCQYDVQQGEwJVUzEPMA0GA1UEChMGQW1hem9uMRkwFwYDVQQDExBBbWF6\nb24gUm9vdCBDQSAxMB4XDTIyMDgyMzIyMjEyOFoXDTMwMDgyMzIyMjEyOFowPDEL\nMAkGA1UEBhMCVVMxDzANBgNVBAoTBkFtYXpvbjEcMBoGA1UEAxMTQW1hem9uIFJT\nQSAyMDQ4IE0wMTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAOtxLKnL\nH4gokjIwr4pXD3i3NyWVVYesZ1yX0yLI2qIUZ2t88Gfa4gMqs1YSXca1R/lnCKeT\nepWSGA+0+fkQNpp/L4C2T7oTTsddUx7g3ZYzByDTlrwS5HRQQqEFE3O1T5tEJP4t\nf+28IoXsNiEzl3UGzicYgtzj2cWCB41eJgEmJmcf2T8TzzK6a614ZPyq/w4CPAff\nnAV4coz96nW3AyiE2uhuB4zQUIXvgVSycW7sbWLvj5TDXunEpNCRwC4kkZjK7rol\njtT2cbb7W2s4Bkg3R42G3PLqBvt2N32e/0JOTViCk8/iccJ4sXqrS1uUN4iB5Nmv\nJK74csVl+0u0UecCAwEAAaOCAVowggFWMBIGA1UdEwEB/wQIMAYBAf8CAQAwDgYD\nVR0PAQH/BAQDAgGGMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjAdBgNV\nHQ4EFgQUgbgOY4qJEhjl+js7UJWf5uWQE4UwHwYDVR0jBBgwFoAUhBjMhTTsvAyU\nlC4IWZzHshBOCggwewYIKwYBBQUHAQEEbzBtMC8GCCsGAQUFBzABhiNodHRwOi8v\nb2NzcC5yb290Y2ExLmFtYXpvbnRydXN0LmNvbTA6BggrBgEFBQcwAoYuaHR0cDov\nL2NydC5yb290Y2ExLmFtYXpvbnRydXN0LmNvbS9yb290Y2ExLmNlcjA/BgNVHR8E\nODA2MDSgMqAwhi5odHRwOi8vY3JsLnJvb3RjYTEuYW1hem9udHJ1c3QuY29tL3Jv\nb3RjYTEuY3JsMBMGA1UdIAQMMAowCAYGZ4EMAQIBMA0GCSqGSIb3DQEBCwUAA4IB\nAQCtAN4CBSMuBjJitGuxlBbkEUDeK/pZwTXv4KqPK0G50fOHOQAd8j21p0cMBgbG\nkfMHVwLU7b0XwZCav0h1ogdPMN1KakK1DT0VwA/+hFvGPJnMV1Kx2G4S1ZaSk0uU\n5QfoiYIIano01J5k4T2HapKQmmOhS/iPtuo00wW+IMLeBuKMn3OLn005hcrOGTad\nhcmeyfhQP7Z+iKHvyoQGi1C0ClymHETx/chhQGDyYSWqB/THwnN15AwLQo0E5V9E\nSJlbe4mBlqeInUsNYugExNf+tOiybcrswBy8OFsd34XOW3rjSUtsuafd9AWySa3h\nxRRrwszrzX/WWGm6wyB+f7C4\n-----END CERTIFICATE-----\n-----BEGIN CERTIFICATE-----\nMIIEkjCCA3qgAwIBAgITBn+USionzfP6wq4rAfkI7rnExjANBgkqhkiG9w0BAQsF\nADCBmDELMAkGA1UEBhMCVVMxEDAOBgNVBAgTB0FyaXpvbmExEzARBgNVBAcTClNj\nb3R0c2RhbGUxJTAjBgNVBAoTHFN0YXJmaWVsZCBUZWNobm9sb2dpZXMsIEluYy4x\nOzA5BgNVBAMTMlN0YXJmaWVsZCBTZXJ2aWNlcyBSb290IENlcnRpZmljYXRlIEF1\ndGhvcml0eSAtIEcyMB4XDTE1MDUyNTEyMDAwMFoXDTM3MTIzMTAxMDAwMFowOTEL\nMAkGA1UEBhMCVVMxDzANBgNVBAoTBkFtYXpvbjEZMBcGA1UEAxMQQW1hem9uIFJv\nb3QgQ0EgMTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALJ4gHHKeNXj\nca9HgFB0fW7Y14h29Jlo91ghYPl0hAEvrAIthtOgQ3pOsqTQNroBvo3bSMgHFzZM\n9O6II8c+6zf1tRn4SWiw3te5djgdYZ6k/oI2peVKVuRF4fn9tBb6dNqcmzU5L/qw\nIFAGbHrQgLKm+a/sRxmPUDgH3KKHOVj4utWp+UhnMJbulHheb4mjUcAwhmahRWa6\nVOujw5H5SNz/0egwLX0tdHA114gk957EWW67c4cX8jJGKLhD+rcdqsq08p8kDi1L\n93FcXmn/6pUCyziKrlA4b9v7LWIbxcceVOF34GfID5yHI9Y/QCB/IIDEgEw+OyQm\njgSubJrIqg0CAwEAAaOCATEwggEtMA8GA1UdEwEB/wQFMAMBAf8wDgYDVR0PAQH/\nBAQDAgGGMB0GA1UdDgQWBBSEGMyFNOy8DJSULghZnMeyEE4KCDAfBgNVHSMEGDAW\ngBScXwDfqgHXMCs4iKK4bUqc8hGRgzB4BggrBgEFBQcBAQRsMGowLgYIKwYBBQUH\nMAGGImh0dHA6Ly9vY3NwLnJvb3RnMi5hbWF6b250cnVzdC5jb20wOAYIKwYBBQUH\nMAKGLGh0dHA6Ly9jcnQucm9vdGcyLmFtYXpvbnRydXN0LmNvbS9yb290ZzIuY2Vy\nMD0GA1UdHwQ2MDQwMqAwoC6GLGh0dHA6Ly9jcmwucm9vdGcyLmFtYXpvbnRydXN0\nLmNvbS9yb290ZzIuY3JsMBEGA1UdIAQKMAgwBgYEVR0gADANBgkqhkiG9w0BAQsF\nAAOCAQEAYjdCXLwQtT6LLOkMm2xF4gcAevnFWAu5CIw+7bMlPLVvUOTNNWqnkzSW\nMiGpSESrnO09tKpzbeR/FoCJbM8oAxiDR3mjEH4wW6w7sGDgd9QIpuEdfF7Au/ma\neyKdpwAJfqxGF4PcnCZXmTA5YpaP7dreqsXMGz7KQ2hsVxa81Q4gLv7/wmpdLqBK\nbRRYh5TmOTFffHPLkIhqhBGWJ6bt2YFGpn6jcgAKUj6DiAdjd4lpFw85hdKrCEVN\n0FE6/V1dN2RMfjCyVSRCnTawXZwXgWHxyvkQAiSr6w10kY17RSlQOYiypok1JR4U\nakcjMS9cmvqtmg5iUaQqqcT5NJ0hGA==\n-----END CERTIFICATE-----\n-----BEGIN CERTIFICATE-----\nMIIEdTCCA12gAwIBAgIJAKcOSkw0grd/MA0GCSqGSIb3DQEBCwUAMGgxCzAJBgNV\nBAYTAlVTMSUwIwYDVQQKExxTdGFyZmllbGQgVGVjaG5vbG9naWVzLCBJbmMuMTIw\nMAYDVQQLEylTdGFyZmllbGQgQ2xhc3MgMiBDZXJ0aWZpY2F0aW9uIEF1dGhvcml0\neTAeFw0wOTA5MDIwMDAwMDBaFw0zNDA2MjgxNzM5MTZaMIGYMQswCQYDVQQGEwJV\nUzEQMA4GA1UECBMHQXJpem9uYTETMBEGA1UEBxMKU2NvdHRzZGFsZTElMCMGA1UE\nChMcU3RhcmZpZWxkIFRlY2hub2xvZ2llcywgSW5jLjE7MDkGA1UEAxMyU3RhcmZp\nZWxkIFNlcnZpY2VzIFJvb3QgQ2VydGlmaWNhdGUgQXV0aG9yaXR5IC0gRzIwggEi\nMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDVDDrEKvlO4vW+GZdfjohTsR8/\ny8+fIBNtKTrID30892t2OGPZNmCom15cAICyL1l/9of5JUOG52kbUpqQ4XHj2C0N\nTm/2yEnZtvMaVq4rtnQU68/7JuMauh2WLmo7WJSJR1b/JaCTcFOD2oR0FMNnngRo\nOt+OQFodSk7PQ5E751bWAHDLUu57fa4657wx+UX2wmDPE1kCK4DMNEffud6QZW0C\nzyyRpqbn3oUYSXxmTqM6bam17jQuug0DuDPfR+uxa40l2ZvOgdFFRjKWcIfeAg5J\nQ4W2bHO7ZOphQazJ1FTfhy/HIrImzJ9ZVGif/L4qL8RVHHVAYBeFAlU5i38FAgMB\nAAGjgfAwge0wDwYDVR0TAQH/BAUwAwEB/zAOBgNVHQ8BAf8EBAMCAYYwHQYDVR0O\nBBYEFJxfAN+qAdcwKziIorhtSpzyEZGDMB8GA1UdIwQYMBaAFL9ft9HO3R+G9FtV\nrNzXEMIOqYjnME8GCCsGAQUFBwEBBEMwQTAcBggrBgEFBQcwAYYQaHR0cDovL28u\nc3MyLnVzLzAhBggrBgEFBQcwAoYVaHR0cDovL3guc3MyLnVzL3guY2VyMCYGA1Ud\nHwQfMB0wG6AZoBeGFWh0dHA6Ly9zLnNzMi51cy9yLmNybDARBgNVHSAECjAIMAYG\nBFUdIAAwDQYJKoZIhvcNAQELBQADggEBACMd44pXyn3pF3lM8R5V/cxTbj5HD9/G\nVfKyBDbtgB9TxF00KGu+x1X8Z+rLP3+QsjPNG1gQggL4+C/1E2DUBc7xgQjB3ad1\nl08YuW3e95ORCLp+QCztweq7dp4zBncdDQh/U90bZKuCJ/Fp1U1ervShw3WnWEQt\n8jxwmKy6abaVd38PMV4s/KCHOkdp8Hlf9BRUpJVeEXgSYCfOn8J3/yNTd126/+pZ\n59vPr5KW7ySaNRB6nJHGDn2Z9j8Z3/VyVOEVqQdZe4O/Ui5GjLIAZHYcSNPYeehu\nVsyuLAOQ1xk4meTKCRlb/weWsKh/NEnfVqn3sF/tM+2MR7cwA130A4w=\n-----END CERTIFICATE-----"
	portNumberUpperBound    = uint(65535)
)

/*
KurtosisConfig should be the interface other modules use to access
the latest configuration values available in Kurtosis CLI configuration.

From the standpoint of the rest of our code, this is the evergreen config value.
This prevents code using configuration from needing to completely change
everytime configuration versions change.

Under the hood, the KurtosisConfig is responsible for reconciling the user's overrides
with the default values for the configuration. It can be thought of as a "resolver" for
the overrides on top of the default config.
*/
type KurtosisConfig struct {
	// Only necessary to store for when we serialize overrides
	overrides *v3.KurtosisConfigV3

	shouldSendMetrics    bool
	clusters             map[string]*KurtosisClusterConfig
	cloudConfig          *KurtosisCloudConfig
	logsAggregatorConfig *KurtosisLogsAggregatorConfig
}

// NewKurtosisConfigFromOverrides constructs a new KurtosisConfig that uses the given overrides
// We leave the overrides as an interface which "quarantines" all versioned config structs into this
// package
func NewKurtosisConfigFromOverrides(uncastedOverrides interface{}) (*KurtosisConfig, error) {
	overrides, err := castUncastedOverrides(uncastedOverrides)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred casting the uncasted config overrides")
	}

	config := &KurtosisConfig{
		overrides:            overrides,
		shouldSendMetrics:    false,
		clusters:             nil,
		cloudConfig:          nil,
		logsAggregatorConfig: nil,
	}

	// Get latest config version
	latestConfigVersion := config_version.ConfigVersion_v0
	for _, configVersion := range config_version.ConfigVersionValues() {
		if uint(configVersion) > uint(latestConfigVersion) {
			latestConfigVersion = configVersion
		}
	}

	// Ensure that the overrides are storing the latest config version
	// From this point onwards, it should be impossible to have the incorrect config version
	config.overrides.ConfigVersion = latestConfigVersion

	// --------------------- Validation --------------------------
	if overrides.ShouldSendMetrics == nil {
		return nil, stacktrace.NewError("An explicit election about sending metrics must be made")
	}
	shouldSendMetrics := *overrides.ShouldSendMetrics

	allClusterOverrides := getDefaultKurtosisClusterConfigOverrides()
	if overrides.KurtosisClusters != nil {
		allClusterOverrides = overrides.KurtosisClusters
	}

	if len(allClusterOverrides) == 0 {
		return nil, stacktrace.NewError("At least one Kurtosis cluster must be specified")
	}

	allClusterConfigs := map[string]*KurtosisClusterConfig{}
	for clusterId, overridesForCluster := range allClusterOverrides {
		clusterConfig, err := NewKurtosisClusterConfigFromOverrides(clusterId, overridesForCluster)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating a Kurtosis cluster config object from overrides: %+v", overridesForCluster)
		}
		allClusterConfigs[clusterId] = clusterConfig
	}

	cloudConfig := &KurtosisCloudConfig{
		ApiUrl:           DefaultCloudConfigApiUrl,
		Port:             DefaultCloudConfigPort,
		CertificateChain: DefaultCertificateChain,
	}
	if overrides.CloudConfig != nil {
		if overrides.CloudConfig.ApiUrl != nil {
			if len(*overrides.CloudConfig.ApiUrl) < 1 {
				return nil, stacktrace.NewError("The CloudConfig ApiUrl must be nonempty")
			} else {
				cloudConfig.ApiUrl = *overrides.CloudConfig.ApiUrl
			}
		}
		if overrides.CloudConfig.Port != nil {
			if *overrides.CloudConfig.Port > portNumberUpperBound {
				return nil, stacktrace.NewError("The CloudConfig Port must be an integer and <= 65535")
			} else {
				cloudConfig.Port = *overrides.CloudConfig.Port
			}
		}
		if overrides.CloudConfig.CertificateChain != nil {
			if len(*overrides.CloudConfig.CertificateChain) < 1 {
				return nil, stacktrace.NewError("The CloudConfig CertificateChain must be nonempty")
			} else {
				cloudConfig.CertificateChain = *overrides.CloudConfig.CertificateChain
			}
		}
	}

	logsAggregatorConfig := &KurtosisLogsAggregatorConfig{
		Image: defaultLogsAggregatorImage,
	}

	if overrides.LogsAggregator != nil {
		if overrides.LogsAggregator.Image != nil && *overrides.LogsAggregator.Image != "" {
			logsAggregatorConfig.Image = *overrides.LogsAggregator.Image
		}

		if len(overrides.LogsAggregator.Sinks) > 0 {
			for sinkId, sinkConfig := range overrides.LogsAggregator.Sinks {
				// The validation logic should be independent of the aggregator library, but we are only using vector
				// for now
				if sinkId == vector.DefaultSinkId {
					return nil, stacktrace.NewError("The LogsAggregator Sinks had a sink named %s which is reserved for Kurtosis default sink", vector.DefaultSinkId)
				}

				// We don't allow users to pass in sink inputs because there is only one source (fluent_bit) and source
				// configuration is currently not supported
				_, found := sinkConfig["inputs"]
				if found {
					return nil, stacktrace.NewError("The LogsAggregator Sink %s must not specify \"inputs\" field", sinkId)
				}
			}

			logsAggregatorConfig.Sinks = overrides.LogsAggregator.Sinks
		}
	}

	return &KurtosisConfig{
		overrides:            overrides,
		shouldSendMetrics:    shouldSendMetrics,
		clusters:             allClusterConfigs,
		cloudConfig:          cloudConfig,
		logsAggregatorConfig: logsAggregatorConfig,
	}, nil
}

// NOTE: We probably want to remove this function entirely
func NewKurtosisConfigFromRequiredFields(shouldSendMetrics bool) (*KurtosisConfig, error) {
	overrides := &v3.KurtosisConfigV3{
		ConfigVersion:     0,
		ShouldSendMetrics: &shouldSendMetrics,
		KurtosisClusters:  nil,
		CloudConfig:       nil,
		LogsAggregator:    nil,
	}
	result, err := NewKurtosisConfigFromOverrides(overrides)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Kurtosis config with did-accept-metrics flag '%v'", shouldSendMetrics)
	}
	return result, nil
}

func NewKurtosisConfigWithMetricsSetFromExistingConfig(config *KurtosisConfig, shouldSendMetrics bool) *KurtosisConfig {
	newConfig := &KurtosisConfig{
		overrides:         config.overrides,
		shouldSendMetrics: shouldSendMetrics,
		clusters:          config.clusters,
		cloudConfig:       config.cloudConfig,
	}
	newConfig.overrides.ShouldSendMetrics = &shouldSendMetrics
	return newConfig
}

func (kurtosisConfig *KurtosisConfig) GetShouldSendMetrics() bool {
	return kurtosisConfig.shouldSendMetrics
}

func (kurtosisConfig *KurtosisConfig) GetKurtosisClusters() map[string]*KurtosisClusterConfig {
	return kurtosisConfig.clusters
}

func (kurtosisConfig *KurtosisConfig) GetOverrides() *v3.KurtosisConfigV3 {
	return kurtosisConfig.overrides
}

func (kurtosisConfig *KurtosisConfig) GetCloudConfig() *KurtosisCloudConfig {
	return kurtosisConfig.cloudConfig
}

func (kurtosisConfig *KurtosisConfig) GetLogsAggregatorConfig() *KurtosisLogsAggregatorConfig {
	return kurtosisConfig.logsAggregatorConfig
}

// ====================================================================================================
//
//	Private Helpers
//
// ====================================================================================================
// This is a separate helper function so that we can use it to ensure that the
func castUncastedOverrides(uncastedOverrides interface{}) (*v3.KurtosisConfigV3, error) {
	castedOverrides, ok := uncastedOverrides.(*v3.KurtosisConfigV3)
	if !ok {
		return nil, stacktrace.NewError("An error occurred casting the uncasted config overrides to the right version")
	}
	return castedOverrides, nil
}

func getDefaultKurtosisClusterConfigOverrides() map[string]*v3.KurtosisClusterConfigV3 {
	dockerClusterType := KurtosisClusterType_Docker.String()
	minikubeClusterType := KurtosisClusterType_Kubernetes.String()
	minikubeKubernetesClusterName := defaultMinikubeClusterKubernetesClusterNameStr
	minikubeStorageClass := defaultMinikubeStorageClass
	minikubeEnclaveDataVolSizeMB := defaultMinikubeEnclaveDataVolumeMB

	result := map[string]*v3.KurtosisClusterConfigV3{
		DefaultDockerClusterName: {
			Type:   &dockerClusterType,
			Config: nil, // Must be nil for Docker
		},
		defaultMinikubeClusterName: {
			Type: &minikubeClusterType,
			Config: &v3.KubernetesClusterConfigV3{
				KubernetesClusterName:  &minikubeKubernetesClusterName,
				StorageClass:           &minikubeStorageClass,
				EnclaveSizeInMegabytes: &minikubeEnclaveDataVolSizeMB,
			},
		},
	}

	return result
}
