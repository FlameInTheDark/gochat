package user

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/gofiber/fiber/v2"
)

const (
	userSettingsDeviceKeyHeader    = "X-Device-Key"
	maxUserSettingsDeviceKeyLength = 128
)

func getUserSettingsDeviceKey(c *fiber.Ctx) (string, error) {
	deviceKey := strings.TrimSpace(c.Get(userSettingsDeviceKeyHeader))
	if deviceKey == "" {
		return "", nil
	}
	if len(deviceKey) > maxUserSettingsDeviceKeyLength {
		return "", fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("%s is too long", userSettingsDeviceKeyHeader))
	}
	return deviceKey, nil
}

func applyDeviceScopedSettings(settings *model.UserSettingsData, deviceKey string) {
	if settings == nil || deviceKey == "" {
		return
	}
	settings.Devices = settings.DevicesForKey(deviceKey)
}

func cloneDevicesByKey(src map[string]model.Devices) map[string]model.Devices {
	if len(src) == 0 {
		return nil
	}

	dst := make(map[string]model.Devices, len(src))
	for key, devices := range src {
		dst[key] = devices
	}
	return dst
}

func mergeStoredDeviceSettings(current, incoming model.UserSettingsData, deviceKey string) model.UserSettingsData {
	merged := cloneDevicesByKey(current.DevicesByKey)
	for key, devices := range incoming.DevicesByKey {
		if merged == nil {
			merged = make(map[string]model.Devices)
		}
		merged[key] = devices
	}
	if deviceKey != "" {
		if merged == nil {
			merged = make(map[string]model.Devices)
		}
		merged[deviceKey] = incoming.Devices
	}
	incoming.DevicesByKey = merged
	return incoming
}

func userSettingsDataFromRecord(record model.UserSettings) (model.UserSettingsData, error) {
	var settings model.UserSettingsData
	if len(record.Settings) > 0 {
		if err := json.Unmarshal(record.Settings, &settings); err != nil {
			return model.UserSettingsData{}, err
		}
	}
	settings.NormalizeCollections()
	return settings, nil
}

func (e *entity) loadStoredUserSettings(ctx context.Context, userID int64) (model.UserSettingsData, error) {
	record, err := e.uset.GetUserSettings(ctx, userID, 0)
	if err != nil {
		return model.UserSettingsData{}, err
	}
	return userSettingsDataFromRecord(record)
}
