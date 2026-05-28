package uuid

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/util/guid"
)

// Generate 生成仿UUID格式的ID（36位，带连字符）
// 基于 guid.S() 格式化为: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
// 保留 GoFrame 的时间戳和追踪特性，同时拥有UUID的美观格式
func Generate() string {
	s := guid.S()
	if len(s) != 32 {
		return s
	}
	// 格式化为 8-4-4-4-12 的UUID样式
	return s[0:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:32]
}

// GenerateShort 生成短代码（4位）
// 适用于验证码、短标识等场景
func GenerateShort() string {
	// 使用完整guid的后4位，避免重复
	s := guid.S()
	return s[len(s)-4:]
}

// GenerateMedium 生成中长度代码（8位）
// 适用于订单号、任务ID等场景
func GenerateMedium() string {
	// 使用完整guid的后8位，避免重复
	s := guid.S()
	return s[len(s)-8:]
}

// GenerateCompact 生成紧凑格式UUID（32位，无连字符）
// 格式: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
func GenerateCompact() string {
	// uuid := guid.S()
	return guid.S()
}

// GenerateV4 生成 RFC 4122 UUID v4 字符串。
func GenerateV4() string {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return Generate()
	}

	randomBytes[6] = (randomBytes[6] & 0x0f) | 0x40
	randomBytes[8] = (randomBytes[8] & 0x3f) | 0x80

	hexStr := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hexStr[0:8], hexStr[8:12], hexStr[12:16], hexStr[16:20], hexStr[20:32])
}

// GenerateClientId 生成客户端标识
//
// 用途:
//   - Claude Code 请求去特征化
//   - 构造 metadata.user_id 字段
//   - 格式: user_{client_id}_account_{account_id}_session_{session_id}
//
// 技术规格:
//   - 使用 crypto/rand 生成 32 字节随机数
//   - 输出 64 位小写十六进制字符串
//   - 密码学安全，难以预测
//
// 返回:
//   - string: 64位小写十六进制字符串
//
// 示例输出: "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef12345678"
func GenerateClientId() string {
	randomBytes := make([]byte, 32) // 32字节 = 64位hex
	_, err := rand.Read(randomBytes)
	if err != nil {
		// 降级策略：使用时间戳 + UUID 组合（极少触发）
		compact := GenerateCompact()
		return fmt.Sprintf("%016x%s", time.Now().UnixNano(), strings.ReplaceAll(compact, "-", "")[:48])
	}
	return hex.EncodeToString(randomBytes)
}

// FormatUUIDFromSeed 从种子字符串生成确定性 UUID
//
// 算法（与 claude-relay-service/requestIdentityService.js 对齐）：
//  1. 使用 SHA-256 哈希种子
//  2. 取前 16 字节
//  3. 设置 UUID v4 版本位和变体位
//  4. 格式化为标准 UUID 格式
//
// 参数:
//   - seed: 种子字符串（如 "accountId::sessionTail"）
//
// 返回:
//   - string: 标准 UUID v4 格式（8-4-4-4-12）
//
// 特性:
//   - 确定性：相同种子始终产生相同 UUID
//   - 隐私保护：无法从 UUID 反推原始种子
//   - 格式兼容：符合 RFC 4122 UUID v4 规范
//
// 示例:
//
//	FormatUUIDFromSeed("test::session") → "e3b0c442-98fc-4c52-..."
func FormatUUIDFromSeed(seed string) string {
	// 1. SHA-256 哈希
	digest := sha256.Sum256([]byte(seed))

	// 2. 取前 16 字节（复制以避免修改原数组）
	bytes := make([]byte, 16)
	copy(bytes, digest[:16])

	// 3. 设置 UUID v4 版本位（第 7 字节的高 4 位 = 0100）
	bytes[6] = (bytes[6] & 0x0f) | 0x40

	// 4. 设置 UUID 变体位（第 9 字节的高 2 位 = 10）
	bytes[8] = (bytes[8] & 0x3f) | 0x80

	// 5. 格式化为 UUID 字符串（8-4-4-4-12）
	hexStr := hex.EncodeToString(bytes)
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hexStr[0:8], hexStr[8:12], hexStr[12:16], hexStr[16:20], hexStr[20:32])
}
