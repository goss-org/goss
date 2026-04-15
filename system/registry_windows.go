//go:build windows

package system

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/sys/windows/registry"

	"github.com/goss-org/goss/util"
)

type defRegistryWindows struct {
	key string
}

func NewDefRegistry(_ context.Context, key string, system *System, config util.Config) Registry {
	return &defRegistryWindows{key: key}
}

func (r *defRegistryWindows) Key() string { return r.key }

func (r *defRegistryWindows) Exists() (bool, error) {
	parts, err := parseRegistryKey(r.key)
	if err != nil {
		return false, err
	}

	hive, err := lookupHive(parts.Hive)
	if err != nil {
		return false, err
	}

	k, err := registry.OpenKey(hive, parts.SubKey, registry.QUERY_VALUE)
	if err != nil {
		return false, nil
	}
	defer k.Close()

	if parts.ValueName == "" {
		// Default value: key exists, that's enough
		return true, nil
	}

	_, _, err = k.GetValue(parts.ValueName, nil)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (r *defRegistryWindows) Value() (string, error) {
	parts, err := parseRegistryKey(r.key)
	if err != nil {
		return "", err
	}

	hive, err := lookupHive(parts.Hive)
	if err != nil {
		return "", err
	}

	k, err := registry.OpenKey(hive, parts.SubKey, registry.QUERY_VALUE)
	if err != nil {
		return "", fmt.Errorf("opening registry key: %w", err)
	}
	defer k.Close()

	_, valType, err := k.GetValue(parts.ValueName, nil)
	if err != nil {
		return "", fmt.Errorf("reading registry value: %w", err)
	}

	return formatValue(valType, k, parts.ValueName)
}

func (r *defRegistryWindows) Type() (string, error) {
	parts, err := parseRegistryKey(r.key)
	if err != nil {
		return "", err
	}

	hive, err := lookupHive(parts.Hive)
	if err != nil {
		return "", err
	}

	k, err := registry.OpenKey(hive, parts.SubKey, registry.QUERY_VALUE)
	if err != nil {
		return "", fmt.Errorf("opening registry key: %w", err)
	}
	defer k.Close()

	_, valType, err := k.GetValue(parts.ValueName, nil)
	if err != nil {
		return "", fmt.Errorf("reading registry value: %w", err)
	}

	return typeName(valType), nil
}

func lookupHive(name string) (registry.Key, error) {
	switch name {
	case "HKLM":
		return registry.LOCAL_MACHINE, nil
	case "HKCU":
		return registry.CURRENT_USER, nil
	case "HKCR":
		return registry.CLASSES_ROOT, nil
	case "HKU":
		return registry.USERS, nil
	case "HKCC":
		return registry.CURRENT_CONFIG, nil
	default:
		return 0, fmt.Errorf("unknown registry hive: %s", name)
	}
}

func formatValue(valType uint32, k registry.Key, name string) (string, error) {
	switch valType {
	case registry.SZ, registry.EXPAND_SZ:
		s, _, err := k.GetStringValue(name)
		if err != nil {
			return "", err
		}
		return s, nil
	case registry.DWORD, registry.QWORD:
		v, _, err := k.GetIntegerValue(name)
		if err != nil {
			return "", err
		}
		return strconv.FormatUint(v, 10), nil
	case registry.BINARY:
		b, _, err := k.GetBinaryValue(name)
		if err != nil {
			return "", err
		}
		return hex.EncodeToString(b), nil
	case registry.MULTI_SZ:
		ss, _, err := k.GetStringsValue(name)
		if err != nil {
			return "", err
		}
		return strings.Join(ss, "\n"), nil
	default:
		return "", fmt.Errorf("unsupported registry value type: %d", valType)
	}
}

func typeName(t uint32) string {
	switch t {
	case registry.SZ:
		return "REG_SZ"
	case registry.EXPAND_SZ:
		return "REG_EXPAND_SZ"
	case registry.DWORD:
		return "REG_DWORD"
	case registry.QWORD:
		return "REG_QWORD"
	case registry.BINARY:
		return "REG_BINARY"
	case registry.MULTI_SZ:
		return "REG_MULTI_SZ"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", t)
	}
}
