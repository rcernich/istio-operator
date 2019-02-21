package v1alpha3

import (
	"encoding/json"
)

func (gc GatewaysConfig) MarshalJSON() ([]byte, error) {
	var data, gatewayData, enabledData, globalData []byte
    var err error
    // marshal components
	enabledData, err = json.Marshal(gc.EnabledField)
	if err != nil {
		return []byte{}, err
	}

    globalData, err = json.Marshal(gc.Global)
	if err != nil {
		return []byte{}, err
	}
	if len(globalData) > 2 {
		globalData = append([]byte(`"global":`), globalData...)
	}

    if len(gc.Gateways) > 0 {
		gatewayData, err = json.Marshal(gc.Gateways)
		if err != nil {
			return []byte{}, err
		}
	} else {
		gatewayData = []byte("{}")
    }

    // assemble components
    enabledLen := len(enabledData)
    globalLen := len(globalData)
    gatewayLen := len(gatewayData)
	data = make([]byte, 0, enabledLen+globalLen+gatewayLen)
	if enabledLen > 2 {
        data = append(data, enabledData[:enabledLen-1]...)
	} else {
		data = append(data, '{')
	}
	if globalLen > 2 {
        if len(data) > 1 {
            data = append(data, byte(','))
        }
        data = append(data, globalData...)
    }
    if gatewayLen > 2 {
        if len(data) > 1 {
            data = append(data, byte(','))
        }
        data = append(data, gatewayData[1:]...)
    } else {
        data = append(data, '}')
    }
	return data, nil
}

func (gc *GatewaysConfig) UnmarshalJSON(data []byte) error {
    rawKeyedData := map[string]json.RawMessage{}
    err := json.Unmarshal(data, &rawKeyedData)
    if err != nil {
        return err
    }
    if value, ok := rawKeyedData["enabled"]; ok {
        err = json.Unmarshal(value, &gc.Enabled)
        if err != nil {
            return err
        }
        delete(rawKeyedData, "enabled")
    }
    if value, ok := rawKeyedData["global"]; ok {
        err = json.Unmarshal(value, &gc.Global)
        if err != nil {
            return err
        }
        delete(rawKeyedData, "global")
    }
    gc.Gateways = map[string]GatewayConfig{}
    for key, value := range rawKeyedData {
        g := GatewayConfig{}
        err = json.Unmarshal(value, &g)
        if err != nil {
            return err
        }
        gc.Gateways[key] = g
    }
	return nil
}