package v2

type GatewaysConfig struct {
    Ingress *GatewayConfig
    Egress *GatewayConfig
    AdditionalGateways map[string]GatewayConfig
}

type GatewayConfig struct {
    Runtime *ComponentRuntimeConfig
    
}