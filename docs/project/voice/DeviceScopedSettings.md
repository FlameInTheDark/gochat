[<- Voice documentation](README.md)

# Device-Scoped Media Settings

`/user/me/settings` now supports device-scoped media preferences so one account can keep different microphone, speaker, and camera selections on different devices.

## Backend contract

- The client may send an optional `X-Device-Key` header on both `GET /user/me/settings` and `POST /user/me/settings`.
- The header value should be a stable opaque key for the current installation or browser profile.
- When the header is present, the backend stores the current `settings.devices` payload into `settings.devices_by_key[deviceKey]`.
- When the header is present on reads, the backend resolves `settings.devices` from `settings.devices_by_key[deviceKey]` and falls back to the legacy top-level `settings.devices` field if no bucket exists yet.
- The backend keeps at most 16 device-specific buckets per user and evicts the least recently updated bucket when a new device key exceeds that limit.
- Existing clients that do not send `X-Device-Key` continue to work and still use the legacy shared `settings.devices` value.

## What the frontend needs to do

1. Generate one stable key per installation or browser profile and keep it in local storage.
2. Send that key as `X-Device-Key` every time the app calls `GET /user/me/settings`.
3. Send the same key as `X-Device-Key` every time the app calls `POST /user/me/settings`.
4. Keep sending the normal full settings payload. The backend still expects `settings.devices` to contain the current device's selected media settings.

## Recommended key strategy

- Prefer a client-generated installation ID over a hardware fingerprint.
- A random UUID created on first launch and stored locally is the safest option.
- If you still want a hash, hash the generated installation ID before sending it, not raw hardware details.
- Avoid hashing labels from `enumerateDevices()` because browser labels can change, may be permission-dependent, and are not reliably stable across reinstalls or browser resets.

## Example

```ts
const DEVICE_KEY_STORAGE = "gochat_device_key";

function getDeviceKey(): string {
  const stored = window.localStorage.getItem(DEVICE_KEY_STORAGE);
  if (stored) {
    return stored;
  }

  const created = crypto.randomUUID();
  window.localStorage.setItem(DEVICE_KEY_STORAGE, created);
  return created;
}

async function getUserSettings() {
  const deviceKey = getDeviceKey();

  return fetch("/user/me/settings", {
    headers: {
      "X-Device-Key": deviceKey,
    },
    credentials: "include",
  });
}
```

## Notes

- `devices_by_key` is persisted on the backend for compatibility, but the frontend can continue reading and writing only `settings.devices`.
- If the frontend skips the header on one request, that request falls back to the legacy shared device settings bucket.
