package user

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/gofiber/fiber/v2"
)

const (
	userSettingsDeviceKeyHeader    = "X-Device-Key"
	maxUserSettingsDeviceKeyLength = 128
	maxStoredDeviceSettingsBuckets = 16
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

func normalizeDeviceUsageOrder(order []string, devices map[string]model.Devices) []string {
	if len(devices) == 0 {
		return nil
	}

	normalized := make([]string, 0, len(devices))
	seen := make(map[string]struct{}, len(devices))

	for _, key := range order {
		if _, ok := devices[key]; !ok {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, key)
	}

	if len(seen) == len(devices) {
		return normalized
	}

	remaining := make([]string, 0, len(devices)-len(seen))
	for key := range devices {
		if _, ok := seen[key]; ok {
			continue
		}
		remaining = append(remaining, key)
	}
	sort.Strings(remaining)

	return append(normalized, remaining...)
}

func touchDeviceUsageKey(order []string, key string) []string {
	if key == "" {
		return order
	}

	next := make([]string, 0, len(order)+1)
	for _, existing := range order {
		if existing == key {
			continue
		}
		next = append(next, existing)
	}

	return append(next, key)
}

func trimStoredDeviceSettings(devices map[string]model.Devices, order []string, limit int) (map[string]model.Devices, []string) {
	if len(devices) == 0 {
		return nil, nil
	}
	if limit <= 0 {
		return nil, nil
	}

	order = normalizeDeviceUsageOrder(order, devices)
	for len(devices) > limit && len(order) > 0 {
		oldest := order[0]
		order = order[1:]
		delete(devices, oldest)
	}

	if len(devices) == 0 {
		return nil, nil
	}

	return devices, normalizeDeviceUsageOrder(order, devices)
}

func mergeStoredDeviceSettings(current, incoming model.UserSettingsData, deviceKey string) model.UserSettingsData {
	merged := cloneDevicesByKey(current.DevicesByKey)
	order := normalizeDeviceUsageOrder(current.DeviceUsageOrder(), merged)

	incomingKeys := make([]string, 0, len(incoming.DevicesByKey))
	for key, devices := range incoming.DevicesByKey {
		if merged == nil {
			merged = make(map[string]model.Devices)
		}
		merged[key] = devices
		incomingKeys = append(incomingKeys, key)
	}
	sort.Strings(incomingKeys)
	for _, key := range incomingKeys {
		order = touchDeviceUsageKey(order, key)
	}
	if deviceKey != "" {
		if merged == nil {
			merged = make(map[string]model.Devices)
		}
		merged[deviceKey] = incoming.Devices
		order = touchDeviceUsageKey(order, deviceKey)
	}

	merged, order = trimStoredDeviceSettings(merged, order, maxStoredDeviceSettingsBuckets)
	incoming.DevicesByKey = merged
	incoming.SetDeviceUsageOrder(order)
	return incoming
}

func userSettingsDataFromRecord(record model.UserSettings) (model.UserSettingsData, error) {
	return model.UnmarshalStoredUserSettingsData(record.Settings)
}

func (e *entity) loadStoredUserSettings(ctx context.Context, userID int64) (model.UserSettingsData, error) {
	record, err := e.uset.GetUserSettings(ctx, userID, 0)
	if err != nil {
		return model.UserSettingsData{}, err
	}
	return userSettingsDataFromRecord(record)
}
