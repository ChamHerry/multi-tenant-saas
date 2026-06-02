package me

import (
	"encoding/binary"
	"fmt"
	"net/netip"
	"strings"

	"multi-tenant-saas/api/me/v1"
	"multi-tenant-saas/internal/service"
)

func toSessionItem(session *service.AuthSession, currentSessionID string) v1.SessionItem {
	return v1.SessionItem{
		SessionID:    session.ID,
		IsCurrent:    session.ID == currentSessionID,
		Device:       parseSessionDevice(session.UserAgent),
		IP:           maskSessionIP(session.IP),
		Location:     nil,
		LastActiveAt: session.LastUsedAt,
		CreatedAt:    session.CreatedAt,
		ExpiresAt:    session.ExpiresAt,
	}
}

func parseSessionDevice(userAgent string) v1.SessionDevice {
	raw := strings.TrimSpace(userAgent)
	lower := strings.ToLower(raw)
	device := v1.SessionDevice{
		Browser: "Unknown",
		OS:      "Unknown",
	}
	if raw == "" {
		return device
	}
	switch {
	case strings.Contains(lower, "edg/"):
		device.Browser = "Edge"
		device.BrowserVersion = tokenAfter(raw, "Edg/")
	case strings.Contains(lower, "opr/"):
		device.Browser = "Opera"
		device.BrowserVersion = tokenAfter(raw, "OPR/")
	case strings.Contains(lower, "firefox/"):
		device.Browser = "Firefox"
		device.BrowserVersion = tokenAfter(raw, "Firefox/")
	case strings.Contains(lower, "fxios/"):
		device.Browser = "Firefox"
		device.BrowserVersion = tokenAfter(raw, "FxiOS/")
	case strings.Contains(lower, "crios/"):
		device.Browser = "Chrome"
		device.BrowserVersion = tokenAfter(raw, "CriOS/")
	case strings.Contains(lower, "chrome/"):
		device.Browser = "Chrome"
		device.BrowserVersion = tokenAfter(raw, "Chrome/")
	case strings.Contains(lower, "safari/") && strings.Contains(lower, "version/"):
		device.Browser = "Safari"
		device.BrowserVersion = tokenAfter(raw, "Version/")
	case strings.Contains(lower, "safari/"):
		device.Browser = "Safari"
		device.BrowserVersion = tokenAfter(raw, "Safari/")
	}
	switch {
	case strings.Contains(lower, "iphone") || strings.Contains(lower, "ipad") || strings.Contains(lower, "ipod"):
		device.OS = "iOS"
	case strings.Contains(lower, "android"):
		device.OS = "Android"
	case strings.Contains(lower, "windows"):
		device.OS = "Windows"
	case strings.Contains(lower, "mac os x") || strings.Contains(lower, "macintosh"):
		device.OS = "macOS"
	case strings.Contains(lower, "linux"):
		device.OS = "Linux"
	}
	device.IsMobile = strings.Contains(lower, "mobile") ||
		strings.Contains(lower, "android") ||
		strings.Contains(lower, "iphone") ||
		strings.Contains(lower, "ipad") ||
		strings.Contains(lower, "ipod")
	return device
}

func tokenAfter(value, marker string) string {
	start := strings.Index(value, marker)
	if start < 0 {
		return ""
	}
	rest := value[start+len(marker):]
	end := len(rest)
	for i, r := range rest {
		if r == ' ' || r == ')' || r == ';' {
			end = i
			break
		}
	}
	return rest[:end]
}

func maskSessionIP(ip string) string {
	addr, err := netip.ParseAddr(strings.TrimSpace(ip))
	if err != nil {
		return "unknown"
	}
	if addr.Is4() {
		octets := addr.As4()
		return fmt.Sprintf("%d.%d.*.*", octets[0], octets[1])
	}
	segments := addr.As16()
	return fmt.Sprintf("%x:%x:%x:%x::*",
		binary.BigEndian.Uint16(segments[0:2]),
		binary.BigEndian.Uint16(segments[2:4]),
		binary.BigEndian.Uint16(segments[4:6]),
		binary.BigEndian.Uint16(segments[6:8]),
	)
}
