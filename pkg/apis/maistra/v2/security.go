package v2

type SecurityConfig struct {
	MutualTLSConfig MutualTLSConfig
}

type MutualTLSConfig struct {
	// .Values.global.mtls.auto
	Auto                 bool
	Trust                TrustConfig
	CertificateAuthority CertificateAuthorityConfig
	Identity             IdentityConfig
	ControlPlane         ControlPlaneMTLSConfig
}

type TrustConfig struct {
	//.Values.global.trustDomain, maps to trustDomain
	// The trust domain corresponds to the trust root of a system.
	// Refer to https://github.com/spiffe/spiffe/blob/master/standards/SPIFFE-ID.md#21-trust-domain
	// XXX: can this be consolidated with clusterDomainSuffix?
	Domain string
	// .Values.global.trustDomainAliases, maps to trustDomainAliases
	//  Any service with the identity "td1/ns/foo/sa/a-service-account", "td2/ns/foo/sa/a-service-account",
	//  or "td3/ns/foo/sa/a-service-account" will be treated the same in the Istio mesh.
	AdditionalDomains []string
}

type CertificateAuthorityConfig struct {
	// .Values.global.pilotCertProvider (istiod, kubernetes, custom)
	Type CertificateAuthorityType
	// each of these produces a CAEndpoint, i.e. CA_ADDR
	Istiod *IstiodCertificateAuthorityConfig
	Custom *CustomCertificateAuthorityConfig
}

type CertificateAuthorityType string

const (
	CertificateAuthorityTypeIstiod CertificateAuthorityType = "Istiod"
	CertificateAuthorityTypeCustom CertificateAuthorityType = "Custom"
)

type IstiodCertificateAuthorityConfig struct {
	// .Values.global.jwtPolicy, local=first-party-jwt, external=third-party-jwt
	Type IstioCertificateSignerType

	SelfSigned *IstioSelfSignedCertificateSignerConfig
	PrivateKey *IstioPrivateKeyCertificateSignerConfig
	// default TTL for generated workload certificates, used if CSR is not specified (< 0)
	// env DEFAULT_WORKLOAD_CERT_TTL
	// defaults to 24 hours
	WorkloadCertTTLDefault string
	// maximum TTL for generated workload certificates
	// env MAX_WORKLOAD_CERT_TTL
	// defaults to 90 days
	WorkloadCertTTLMax string
}

type IstioCertificateSignerType string

const (
	IstioCertificateSignerTypePrivateKey IstioCertificateSignerType = "PrivateKey"
	IstioCertificateSignerTypeSelfSigned IstioCertificateSignerType = "SelfSigned"
)

// nothing in here is currently configurable, except RootCADir
type IstioPrivateKeyCertificateSignerConfig struct {
	// hard coded to use a secret named cacerts
	EncryptionSecret string
	// ROOT_CA_DIR, defaults to /etc/cacerts
	// Mount directory for encryption secret
	// XXX: currently, not configurable in the charts
	RootCADir string
	// hard coded to ca-key.pem
	SigningKeyFile string
	// hard coded to ca-cert.pem
	SigningCertFile string
	// hard coded to root-cert.pem
	RootCertFile string
	// hard coded to cert-chain.pem
	CertChainFile string
}

type IstioSelfSignedCertificateSignerConfig struct {
	// TTL for self-signed root certificate
	// env CITADEL_SELF_SIGNED_CA_CERT_TTL
	// default is 10 years
	TTL string
	// grace period percentile for self-signed cert
	// env CITADEL_SELF_SIGNED_ROOT_CERT_GRACE_PERIOD_PERCENTILE
	// default is 20%
	GracePeriod string
	// interval with which certificate is checked for cert rotation
	// env CITADEL_SELF_SIGNED_ROOT_CERT_CHECK_INTERVAL
	// default is 1 hour, zero or negative value disables cert rotation
	CheckPeriod string
	// use jitter for cert rotation
	// env CITADEL_ENABLE_JITTER_FOR_ROOT_CERT_ROTATOR
	// defaults to true
	EnableJitter bool
	// currently uses TrustDomain
	Org string
}

type CustomCertificateAuthorityConfig struct {
	// .Values.global.caAddress
	// XXX: assumption is this is a grpc endpoint that provides methods like istio.v1.auth.IstioCertificateService/CreateCertificate
	Address string
}

type IdentityConfig struct {
	// .Values.global.jwtPolicy
	Type       IdentityConfigType
	Kubernetes *KubernetesIdentityConfig
	ThirdParty *ThirdPartyIdentityConfig
}

type IdentityConfigType string

const (
	IdentityConfigTypeKubernetes IdentityConfigType = "Kubernetes" // first-party-jwt
	IdentityConfigTypeThirdParty IdentityConfigType = "ThirdParty" // third-party-jwt
)

type KubernetesIdentityConfig struct {
	// jwtPolicy=first-party-jwt, uses /var/run/secrets/kubernetes.io/serviceaccount/token
}

type ThirdPartyIdentityConfig struct {
	// default /var/run/secrets/tokens/istio-token
	// XXX: projects service account token with specified audience (istio-ca)
	// XXX: not configurable
	TokenPath string
	// env TOKEN_ISSUER, defaults to iss in specified token
	Issuer string
	// env AUDIENCE
	// .Values.global.sds.token.aud, defaults to istio-ca
	Audience string
}

type ControlPlaneMTLSConfig struct {
	// .Values.global.controlPlaneSecurityEnabled
    EnableControlPlaneSecurity bool
    // .Values.global.pilotCertProvider
    // Provider used to generate serving certs for istiod (pilot)
	CertProvider               ControlPlaneCertProviderType
}

type ControlPlaneCertProviderType string

const (
	ControlPlaneCertProviderTypeIstiod     ControlPlaneCertProviderType = "Istiod"
    ControlPlaneCertProviderTypeKubernetes ControlPlaneCertProviderType = "Kubernetes"
    // Not quite sure what this means. Presumably, the key and cert chain have been mounted specially
	ControlPlaneCertProviderTypeCustom     ControlPlaneCertProviderType = "Custom"
)
