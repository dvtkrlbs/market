package api

//go:generate go tool oapi-codegen --config common/oapi-codegen.yaml common/openapi.yaml
//go:generate go tool oapi-codegen --config health/oapi-codegen.yaml health/openapi.yaml
//go:generate go tool oapi-codegen --config entries/oapi-codegen.yaml entries/openapi.yaml
//go:generate go tool oapi-codegen --config orders/oapi-codegen.yaml orders/openapi.yaml
