package jwt

import (
	"fmt"
	"testing"
	"time"
)

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name    string
		uid     int64
		wantErr bool
	}{
		{
			name:    "正常生成token",
			uid:     12345,
			wantErr: false,
		},
		{
			name:    "用户ID为0",
			uid:     0,
			wantErr: false,
		},
		{
			name:    "用户ID为负数",
			uid:     -1,
			wantErr: false,
		},
		{
			name:    "大用户ID",
			uid:     999999999,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(tt.uid)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && token == "" {
				t.Error("GenerateToken() returned empty token")
			}
		})
	}
}

func TestParseToken(t *testing.T) {
	// 首先生成一个有效的token用于测试
	validUID := int64(12345)
	validToken, err := GenerateToken(validUID)
	if err != nil {
		t.Fatalf("Failed to generate valid token for testing: %v", err)
	}

	tests := []struct {
		name    string
		token   string
		wantUID int64
		wantErr bool
	}{
		{
			name:    "解析有效token",
			token:   validToken,
			wantUID: validUID,
			wantErr: false,
		},
		{
			name:    "解析空token",
			token:   "",
			wantUID: 0,
			wantErr: true,
		},
		{
			name:    "解析无效token",
			token:   "invalid.token.here",
			wantUID: 0,
			wantErr: true,
		},
		{
			name:    "解析格式错误的token",
			token:   "not.a.valid.jwt.token",
			wantUID: 0,
			wantErr: true,
		},
		{
			name:    "解析过期token",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEyMzQ1LCJleHAiOjE2MzQ1Njc4OTB9.invalid_signature",
			wantUID: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ParseToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if claims == nil {
					t.Error("ParseToken() returned nil claims for valid token")
					return
				}
				if claims.Uid != tt.wantUID {
					t.Errorf("ParseToken() UID = %v, want %v", claims.Uid, tt.wantUID)
				}
			}
		})
	}
}

func TestGenerateAndParseToken(t *testing.T) {
	// 测试生成token后立即解析是否能得到相同的UID
	testUIDs := []int64{1, 100, 1000, 999999}

	for _, uid := range testUIDs {
		t.Run(fmt.Sprintf("UID_%d", uid), func(t *testing.T) {
			// 生成token
			token, err := GenerateToken(uid)
			if err != nil {
				t.Errorf("GenerateToken() failed: %v", err)
				return
			}

			// 解析token
			claims, err := ParseToken(token)
			if err != nil {
				t.Errorf("ParseToken() failed: %v", err)
				return
			}

			// 验证UID是否一致
			if claims.Uid != uid {
				t.Errorf("UID mismatch: got %v, want %v", claims.Uid, uid)
			}

			// 验证token的过期时间
			if claims.ExpiresAt == nil {
				t.Error("Token expiration time is nil")
			} else if claims.ExpiresAt.Time.Before(time.Now()) {
				t.Error("Token has already expired")
			}
		})
	}
}

func TestTokenExpiration(t *testing.T) {
	uid := int64(12345)

	// 生成token
	token, err := GenerateToken(uid)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// 解析token
	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	// 验证token的过期时间是否在合理范围内（24小时）
	expectedExpiry := time.Now().Add(24 * time.Hour)
	actualExpiry := claims.ExpiresAt.Time

	// 允许1分钟的误差
	tolerance := time.Minute
	if actualExpiry.After(expectedExpiry.Add(tolerance)) || actualExpiry.Before(expectedExpiry.Add(-tolerance)) {
		t.Errorf("Token expiration time is not within expected range. Expected around %v, got %v", expectedExpiry, actualExpiry)
	}
}
