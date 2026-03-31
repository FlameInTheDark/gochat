package model

import "encoding/json"

type storedUserSettingsData struct {
	UserSettingsData
	DeviceUsageOrder []string `json:"_device_usage_order,omitempty"`
}

func MarshalStoredUserSettingsData(settings UserSettingsData) ([]byte, error) {
	payload := storedUserSettingsData{
		UserSettingsData: settings,
		DeviceUsageOrder: settings.DeviceUsageOrder(),
	}
	return json.Marshal(payload)
}

func UnmarshalStoredUserSettingsData(data []byte) (UserSettingsData, error) {
	var payload storedUserSettingsData
	if len(data) == 0 {
		payload.NormalizeCollections()
		return payload.UserSettingsData, nil
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return UserSettingsData{}, err
	}

	payload.SetDeviceUsageOrder(payload.DeviceUsageOrder)
	payload.NormalizeCollections()
	return payload.UserSettingsData, nil
}
