package wifi

import (
	"crypto/rand"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/windows/registry"
)

const adapterClassKey = `SYSTEM\CurrentControlSet\Control\Class\{4d36e972-e325-11ce-bfc1-08002be10318}`

type WifiAdapter struct {
	ClassSubkey    string
	DriverDesc     string
	ConnectionName string
}

func Detect() (WifiAdapter, error) {
	interfaceDesc, connectionName, err := probePowershellAdapter()
	if err != nil {
		return WifiAdapter{}, err
	}
	subkey, err := findClassSubkey(interfaceDesc)
	if err != nil {
		return WifiAdapter{}, err
	}
	return WifiAdapter{
		ClassSubkey:    subkey,
		DriverDesc:     interfaceDesc,
		ConnectionName: connectionName,
	}, nil
}

func RandomMac() (string, error) {
	rawBytes := make([]byte, 6)
	if _, err := rand.Read(rawBytes); err != nil {
		return "", fmt.Errorf("generate random MAC: %w", err)
	}
	rawBytes[0] = (rawBytes[0] & 0xFC) | 0x02
	return fmt.Sprintf("%02X%02X%02X%02X%02X%02X",
		rawBytes[0], rawBytes[1], rawBytes[2],
		rawBytes[3], rawBytes[4], rawBytes[5]), nil
}

func NormalizeMac(rawInput string) (string, error) {
	hexOnly := strings.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') || (r >= 'A' && r <= 'F') || (r >= 'a' && r <= 'f') {
			return r
		}
		return -1
	}, rawInput)
	cleaned := strings.ToUpper(hexOnly)
	if len(cleaned) != 12 {
		return "", fmt.Errorf("invalid MAC %q: need 12 hex digits (e.g. 02:1A:2B:3C:4D:5E)", rawInput)
	}
	if !isSpoofable(cleaned) {
		return "", fmt.Errorf("invalid MAC %q: 2nd hex digit must be 2, 6, A or E (e.g. 02:xx:xx:xx:xx:xx)", rawInput)
	}
	return cleaned, nil
}

func DisplayMac(macWithoutSeparators string) string {
	parts := make([]string, 0, 6)
	for i := 0; i < 12; i += 2 {
		parts = append(parts, macWithoutSeparators[i:i+2])
	}
	return strings.Join(parts, "-")
}

func SpoofedMac(adapter WifiAdapter) (string, bool) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, adapterClassKey+`\`+adapter.ClassSubkey, registry.QUERY_VALUE)
	if err != nil {
		return "", false
	}
	defer key.Close()
	stored, _, err := key.GetStringValue("NetworkAddress")
	if err != nil || stored == "" {
		return "", false
	}
	return stored, true
}

func CurrentMac(connectionName string) string {
	out, err := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-Command",
		fmt.Sprintf(`Get-NetAdapter -Name "%s" | Select-Object -ExpandProperty MacAddress`, connectionName)).CombinedOutput()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func ApplySpoof(adapter WifiAdapter, macWithoutSeparators string) error {
	if err := writeNetworkAddress(adapter.ClassSubkey, macWithoutSeparators); err != nil {
		return err
	}
	return Restart(adapter.ConnectionName)
}

func Revert(adapter WifiAdapter) error {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, adapterClassKey+`\`+adapter.ClassSubkey,
		registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return fmt.Errorf("open adapter registry key (run as administrator): %w", err)
	}
	defer key.Close()
	if err := key.DeleteValue("NetworkAddress"); err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("clear NetworkAddress: %w", err)
	}
	return Restart(adapter.ConnectionName)
}

func writeNetworkAddress(classSubkey, macWithoutSeparators string) error {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, adapterClassKey+`\`+classSubkey,
		registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return fmt.Errorf("open adapter registry key (run as administrator): %w", err)
	}
	defer key.Close()
	if err := key.SetStringValue("NetworkAddress", macWithoutSeparators); err != nil {
		return fmt.Errorf("write NetworkAddress: %w", err)
	}
	return nil
}

func Restart(connectionName string) error {
	if err := hiddenCmd("netsh", "interface", "set", "interface", "name="+connectionName, "admin=disable").Run(); err != nil {
		return fmt.Errorf("disable adapter %q: %w", connectionName, err)
	}
	time.Sleep(2 * time.Second)
	if err := hiddenCmd("netsh", "interface", "set", "interface", "name="+connectionName, "admin=enable").Run(); err != nil {
		return fmt.Errorf("enable adapter %q: %w", connectionName, err)
	}
	time.Sleep(3 * time.Second)
	return nil
}

func probePowershellAdapter() (interfaceDesc, connectionName string, err error) {
	out, cmdErr := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-Command",
		`Get-NetAdapter -Physical | Where-Object { $_.MediaType -eq 'Native 802.11' -or $_.InterfaceDescription -like '*8821CE*' -or $_.InterfaceDescription -like '*802.11ac Wireless*' } | Select-Object -First 1 | ForEach-Object { $_.InterfaceDescription + '|' + $_.Name }`).CombinedOutput()
	if cmdErr != nil {
		return "", "", fmt.Errorf("auto-detect Wi-Fi adapter: %w", cmdErr)
	}
	parts := strings.Split(strings.TrimSpace(string(out)), "|")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", fmt.Errorf("auto-detect Wi-Fi adapter: no wireless adapter found (8821CE / 802.11ac)")
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}

func findClassSubkey(interfaceDesc string) (string, error) {
	base, err := registry.OpenKey(registry.LOCAL_MACHINE, adapterClassKey, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return "", fmt.Errorf("open adapter class key (run as administrator): %w", err)
	}
	defer base.Close()
	subkeys, err := base.ReadSubKeyNames(-1)
	if err != nil {
		return "", fmt.Errorf("list adapter keys: %w", err)
	}
	for _, subkey := range subkeys {
		child, err := registry.OpenKey(registry.LOCAL_MACHINE, adapterClassKey+`\`+subkey, registry.QUERY_VALUE)
		if err != nil {
			continue
		}
		desc, _, err := child.GetStringValue("DriverDesc")
		child.Close()
		if err != nil {
			continue
		}
		if desc == interfaceDesc {
			return subkey, nil
		}
	}
	return "", fmt.Errorf("registry key for %q not found", interfaceDesc)
}

func isSpoofable(macWithoutSeparators string) bool {
	var firstByte int
	if _, err := fmt.Sscanf(macWithoutSeparators[:2], "%02X", &firstByte); err != nil {
		return false
	}
	return (firstByte&0x01) == 0 && (firstByte&0x02) == 0x02
}

func hiddenCmd(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd
}
