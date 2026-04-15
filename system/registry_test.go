package system

import (
	"testing"
)

func TestParseRegistryKey(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantHive  string
		wantSub   string
		wantValue string
		wantErr   bool
	}{
		{
			name:      "standard path",
			input:     `HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Run`,
			wantHive:  "HKLM",
			wantSub:   `SOFTWARE\Microsoft\Windows\CurrentVersion`,
			wantValue: "Run",
		},
		{
			name:      "path with spaces",
			input:     `HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\ProductName`,
			wantHive:  "HKLM",
			wantSub:   `SOFTWARE\Microsoft\Windows NT\CurrentVersion`,
			wantValue: "ProductName",
		},
		{
			name:      "default value with trailing backslash",
			input:     `HKLM\SOFTWARE\Microsoft\Windows\`,
			wantHive:  "HKLM",
			wantSub:   `SOFTWARE\Microsoft\Windows`,
			wantValue: "",
		},
		{
			name:      "HKCU hive",
			input:     `HKCU\Software\Microsoft\Windows\CurrentVersion\Explorer\Advanced\Hidden`,
			wantHive:  "HKCU",
			wantSub:   `Software\Microsoft\Windows\CurrentVersion\Explorer\Advanced`,
			wantValue: "Hidden",
		},
		{
			name:      "HKCR hive",
			input:     `HKCR\.txt\Content Type`,
			wantHive:  "HKCR",
			wantSub:   ".txt",
			wantValue: "Content Type",
		},
		{
			name:      "HKU hive",
			input:     `HKU\.DEFAULT\Software\Test`,
			wantHive:  "HKU",
			wantSub:   `.DEFAULT\Software`,
			wantValue: "Test",
		},
		{
			name:      "HKCC hive",
			input:     `HKCC\System\CurrentControlSet\Setting`,
			wantHive:  "HKCC",
			wantSub:   `System\CurrentControlSet`,
			wantValue: "Setting",
		},
		{
			name:      "single segment value under hive",
			input:     `HKLM\ValueOnly`,
			wantHive:  "HKLM",
			wantSub:   "",
			wantValue: "ValueOnly",
		},
		{
			name:      "lowercase hive is normalized",
			input:     `hklm\SOFTWARE\Test`,
			wantHive:  "HKLM",
			wantSub:   "SOFTWARE",
			wantValue: "Test",
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "no backslash",
			input:   "HKLM",
			wantErr: true,
		},
		{
			name:    "invalid hive",
			input:   `INVALID\SOFTWARE\Test`,
			wantErr: true,
		},
		{
			name:    "empty subkey after hive",
			input:   `HKLM\`,
			wantErr: true,
		},
		{
			name:      "explicit separator with backslashes in value name",
			input:     `HKLM\SOFTWARE\Policies\Microsoft\Windows\NetworkProvider\HardenedPaths::\\*\NETLOGON`,
			wantHive:  "HKLM",
			wantSub:   `SOFTWARE\Policies\Microsoft\Windows\NetworkProvider\HardenedPaths`,
			wantValue: `\\*\NETLOGON`,
		},
		{
			name:      "explicit separator with SYSVOL UNC path",
			input:     `HKLM\SOFTWARE\Policies\Microsoft\Windows\NetworkProvider\HardenedPaths::\\*\SYSVOL`,
			wantHive:  "HKLM",
			wantSub:   `SOFTWARE\Policies\Microsoft\Windows\NetworkProvider\HardenedPaths`,
			wantValue: `\\*\SYSVOL`,
		},
		{
			name:      "explicit separator with simple value name",
			input:     `HKLM\SOFTWARE\Microsoft\Windows::ProductName`,
			wantHive:  "HKLM",
			wantSub:   `SOFTWARE\Microsoft\Windows`,
			wantValue: "ProductName",
		},
		{
			name:      "explicit separator with empty value name (default value)",
			input:     `HKLM\SOFTWARE\Microsoft\Windows::`,
			wantHive:  "HKLM",
			wantSub:   `SOFTWARE\Microsoft\Windows`,
			wantValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRegistryKey(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Hive != tt.wantHive {
				t.Errorf("Hive = %q, want %q", got.Hive, tt.wantHive)
			}
			if got.SubKey != tt.wantSub {
				t.Errorf("SubKey = %q, want %q", got.SubKey, tt.wantSub)
			}
			if got.ValueName != tt.wantValue {
				t.Errorf("ValueName = %q, want %q", got.ValueName, tt.wantValue)
			}
		})
	}
}
