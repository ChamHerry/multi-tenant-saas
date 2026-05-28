package uuid

import (
	"strings"
	"testing"
)

// TestFormatUUIDFromSeed_Deterministic 测试确定性：相同种子产生相同 UUID
func TestFormatUUIDFromSeed_Deterministic(t *testing.T) {
	seed := "test-account-id::17cf0fd3-d51b-4b59-977d-b899dafb3022"

	uuid1 := FormatUUIDFromSeed(seed)
	uuid2 := FormatUUIDFromSeed(seed)

	if uuid1 != uuid2 {
		t.Errorf("相同种子应产生相同 UUID\n  第一次: %s\n  第二次: %s", uuid1, uuid2)
	}
}

// TestFormatUUIDFromSeed_Isolation 测试账户隔离：不同 accountId 产生不同 UUID
func TestFormatUUIDFromSeed_Isolation(t *testing.T) {
	sessionId := "17cf0fd3-d51b-4b59-977d-b899dafb3022"

	uuid1 := FormatUUIDFromSeed("account-1::" + sessionId)
	uuid2 := FormatUUIDFromSeed("account-2::" + sessionId)

	if uuid1 == uuid2 {
		t.Errorf("不同账户的相同 session 应产生不同 UUID\n  账户1: %s\n  账户2: %s", uuid1, uuid2)
	}
}

// TestFormatUUIDFromSeed_UUIDFormat 测试 UUID 格式：8-4-4-4-12
func TestFormatUUIDFromSeed_UUIDFormat(t *testing.T) {
	uuid := FormatUUIDFromSeed("test::session")

	// 验证长度：36 字符（32 hex + 4 连字符）
	if len(uuid) != 36 {
		t.Errorf("UUID 长度应为 36，实际为 %d", len(uuid))
	}

	// 验证格式：8-4-4-4-12
	parts := strings.Split(uuid, "-")
	if len(parts) != 5 {
		t.Errorf("UUID 应有 5 个部分，实际为 %d", len(parts))
	}

	expectedLengths := []int{8, 4, 4, 4, 12}
	for i, part := range parts {
		if len(part) != expectedLengths[i] {
			t.Errorf("UUID 第 %d 部分长度应为 %d，实际为 %d", i+1, expectedLengths[i], len(part))
		}
	}
}

// TestFormatUUIDFromSeed_V4Version 测试 UUID v4 版本位
func TestFormatUUIDFromSeed_V4Version(t *testing.T) {
	uuid := FormatUUIDFromSeed("test::session")

	// UUID v4 第三段第一个字符应为 '4'
	parts := strings.Split(uuid, "-")
	if len(parts) < 3 {
		t.Fatal("UUID 格式错误")
	}

	versionChar := parts[2][0]
	if versionChar != '4' {
		t.Errorf("UUID v4 版本位应为 '4'，实际为 '%c'", versionChar)
	}
}

// TestFormatUUIDFromSeed_Variant 测试 UUID 变体位
func TestFormatUUIDFromSeed_Variant(t *testing.T) {
	uuid := FormatUUIDFromSeed("test::session")

	// UUID 变体位：第四段第一个字符应为 '8', '9', 'a', 或 'b'
	parts := strings.Split(uuid, "-")
	if len(parts) < 4 {
		t.Fatal("UUID 格式错误")
	}

	variantChar := parts[3][0]
	validVariants := "89ab"
	if !strings.ContainsRune(validVariants, rune(variantChar)) {
		t.Errorf("UUID 变体位应为 '8', '9', 'a' 或 'b'，实际为 '%c'", variantChar)
	}
}

// TestFormatUUIDFromSeed_ConsistentWithJS 测试与 JS 版本一致性
//
// 预期值来自 claude-relay-service 的 formatUuidFromSeed 函数
// Node.js 测试代码:
//
//	const crypto = require('crypto')
//	function formatUuidFromSeed(seed) {
//	  const digest = crypto.createHash('sha256').update(String(seed)).digest()
//	  const bytes = Buffer.from(digest.subarray(0, 16))
//	  bytes[6] = (bytes[6] & 0x0f) | 0x40
//	  bytes[8] = (bytes[8] & 0x3f) | 0x80
//	  const hex = Array.from(bytes).map(byte => byte.toString(16).padStart(2, '0')).join('')
//	  return `${hex.slice(0,8)}-${hex.slice(8,12)}-${hex.slice(12,16)}-${hex.slice(16,20)}-${hex.slice(20)}`
//	}
func TestFormatUUIDFromSeed_ConsistentWithJS(t *testing.T) {
	testCases := []struct {
		seed     string
		expected string // 由 JS 代码计算得出
	}{
		// 测试用例 1: 简单种子
		{
			seed:     "test-account::test-session",
			expected: "74254044-d33e-4164-887b-fc96d815bbeb",
		},
		// 测试用例 2: 默认值场景
		{
			seed:     "unknown-scheduler::default",
			expected: "0ea80264-aea1-44ab-b084-a35d53076f94",
		},
		// 测试用例 3: 真实格式
		{
			seed:     "acc-123::17cf0fd3-d51b-4b59-977d-b899dafb3022",
			expected: "e61b976a-ef4c-4091-af78-3b7701149519",
		},
		// 测试用例 4: 空种子
		{
			seed:     "",
			expected: "e3b0c442-98fc-4c14-9afb-f4c8996fb924",
		},
	}

	for _, tc := range testCases {
		result := FormatUUIDFromSeed(tc.seed)

		if result != tc.expected {
			t.Errorf("种子 %q:\n  期望: %s\n  实际: %s", tc.seed, tc.expected, result)
		} else {
			t.Logf("✅ 种子 %q → %s (与 JS 一致)", tc.seed, result)
		}
	}
}

// TestFormatUUIDFromSeed_EmptySeed 测试空种子
func TestFormatUUIDFromSeed_EmptySeed(t *testing.T) {
	uuid := FormatUUIDFromSeed("")

	// 空种子也应产生有效 UUID
	if len(uuid) != 36 {
		t.Errorf("空种子应产生有效 UUID，实际长度为 %d", len(uuid))
	}

	// 空种子的 UUID 应是确定性的
	uuid2 := FormatUUIDFromSeed("")
	if uuid != uuid2 {
		t.Errorf("空种子应产生相同 UUID\n  第一次: %s\n  第二次: %s", uuid, uuid2)
	}

	t.Logf("空种子 → %s", uuid)
}

// TestFormatUUIDFromSeed_LongSeed 测试长种子
func TestFormatUUIDFromSeed_LongSeed(t *testing.T) {
	// 超长种子（模拟复杂 accountId + sessionId）
	longSeed := strings.Repeat("a", 1000) + "::" + strings.Repeat("b", 1000)

	uuid := FormatUUIDFromSeed(longSeed)

	if len(uuid) != 36 {
		t.Errorf("长种子应产生有效 UUID，实际长度为 %d", len(uuid))
	}

	t.Logf("长种子（%d 字符）→ %s", len(longSeed), uuid)
}

// BenchmarkFormatUUIDFromSeed 性能基准测试
func BenchmarkFormatUUIDFromSeed(b *testing.B) {
	seed := "account-id-12345678::session-17cf0fd3-d51b-4b59-977d-b899dafb3022"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatUUIDFromSeed(seed)
	}
}
