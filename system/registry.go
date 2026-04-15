package system

import (
	"errors"
	"strings"
)

type Registry interface {
	Key() string
	Exists() (bool, error)
	Value() (string, error)
	Type() (string, error)
}

var ErrRegistryUnsupported = errors.New("registry resource is only supported on Windows")

// registryPathParts holds the parsed components of a registry key path.
type registryPathParts struct {
	Hive      string
	SubKey    string
	ValueName string
}

// parseRegistryKey splits a full registry path into hive, subkey, and
// value name.
//
// Two formats are supported:
//
// Standard format: HIVE\subkey\path\ValueName
// The last backslash-separated segment is the value name. A trailing
// backslash indicates the default value.
//
// Explicit format: HIVE\subkey\path::ValueName
// Use "::" to explicitly separate the subkey from the value name. This
// is required when the value name itself contains backslashes (e.g.
// HardenedPaths entries like "\\*\NETLOGON").
//
// Examples:
//   - "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Run"
//     -> Hive="HKLM", SubKey="SOFTWARE\Microsoft\Windows\CurrentVersion", ValueName="Run"
//   - "HKLM\SOFTWARE\Microsoft\Windows\"
//     -> Hive="HKLM", SubKey="SOFTWARE\Microsoft\Windows", ValueName=""
//   - "HKLM\SOFTWARE\Policies\Microsoft\Windows\NetworkProvider\HardenedPaths::\\*\NETLOGON"
//     -> Hive="HKLM", SubKey="SOFTWARE\Policies\Microsoft\Windows\NetworkProvider\HardenedPaths",
//        ValueName="\\*\NETLOGON"
func parseRegistryKey(key string) (registryPathParts, error) {
	if key == "" {
		return registryPathParts{}, errors.New("empty registry key")
	}

	parts := strings.SplitN(key, `\`, 2)
	if len(parts) < 2 {
		return registryPathParts{}, errors.New("invalid registry key: missing subkey path")
	}

	hive := strings.ToUpper(parts[0])
	switch hive {
	case "HKLM", "HKCU", "HKCR", "HKU", "HKCC":
	default:
		return registryPathParts{}, errors.New("invalid registry hive: " + parts[0])
	}

	rest := parts[1]
	if rest == "" {
		return registryPathParts{}, errors.New("invalid registry key: empty subkey path")
	}

	// Check for explicit "::" separator first. This handles value names
	// that contain backslashes (e.g. HardenedPaths UNC entries).
	if idx := strings.Index(rest, "::"); idx >= 0 {
		return registryPathParts{
			Hive:      hive,
			SubKey:    rest[:idx],
			ValueName: rest[idx+2:],
		}, nil
	}

	// Standard format: split at the last backslash.
	// Trailing backslash means default value (empty value name).
	lastSep := strings.LastIndex(rest, `\`)
	if lastSep < 0 {
		// Single segment after hive: treat as value name under root of hive
		return registryPathParts{
			Hive:      hive,
			SubKey:    "",
			ValueName: rest,
		}, nil
	}

	return registryPathParts{
		Hive:      hive,
		SubKey:    rest[:lastSep],
		ValueName: rest[lastSep+1:],
	}, nil
}
