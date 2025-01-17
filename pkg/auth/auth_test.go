package auth_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mashfeii/chirpy/pkg/auth"
)

func TestHashPassword(t *testing.T) {
	t.Parallel()

	type args struct {
		password string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Simple hashing",
			args: args{
				password: "password",
			},
			wantErr: false,
		},
		{
			name: "More complex characters",
			args: args{
				password: "123Xren'ads;kfje234-82u341jkfljfsf",
			},
			wantErr: false,
		},
		{
			name: "Error: too long password",
			args: args{
				password: "123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := auth.HashPassword(tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && auth.CheckPasswordHash(tt.args.password, got) != nil {
				t.Error("Hash does not match")
			}
		})
	}
}

func TestJWT(t *testing.T) {
	type args struct {
		userID      uuid.UUID
		tokenSecret string
		expiresIn   time.Duration
	}

	tests := []struct {
		name         string
		args         args
		want         string
		wantErr      bool
		tokenInvalid bool
	}{
		{
			name: "Default creation",
			args: args{
				userID:      uuid.New(),
				tokenSecret: "secret",
				expiresIn:   time.Second * 2,
			},
			wantErr:      false,
			tokenInvalid: false,
		},
		{
			name: "Expired token",
			args: args{
				userID:      uuid.New(),
				tokenSecret: "secret",
				expiresIn:   time.Second,
			},
			want:         uuid.Nil.String(),
			wantErr:      false,
			tokenInvalid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := auth.MakeJWT(tt.args.userID, tt.args.tokenSecret, tt.args.expiresIn)

			if (err != nil) != tt.wantErr {
				t.Errorf("MakeJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if strings.Count(got, ".") != 2 {
				t.Errorf("MakeJWT() = %v, invalid format", got)
			}

			time.Sleep(time.Second)

			validated, err := auth.ValidateJWT(got, tt.args.tokenSecret)

			if (err != nil) != tt.tokenInvalid {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, tt.tokenInvalid)
				return
			}

			if err == nil && validated != tt.args.userID {
				t.Errorf("ValidateJWT() = %v, want %v", validated, tt.args.userID)
			}
		})
	}
}

func TestMakeRefreshToken(t *testing.T) {
	for i := range 5 {
		t.Run(fmt.Sprintf("Test_case_%d", i+1), func(t *testing.T) {
			t.Parallel()

			if got := auth.MakeRefreshToken(); len(got) != 64 {
				t.Errorf("MakeRefreshToken() = %v, invalid format", got)
			}
		})
	}
}

func TestGetAuthorizationToken(t *testing.T) {
	type args struct {
		headers http.Header
		apiType string
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Bearer token",
			args: args{
				headers: http.Header{
					"Authorization": []string{"Bearer token"},
				},
				apiType: "Bearer",
			},
			want:    "token",
			wantErr: false,
		},
		{
			name: "APIKey token",
			args: args{
				headers: http.Header{
					"Authorization": []string{"ApiKey token"},
				},
				apiType: "ApiKey",
			},
			want:    "token",
			wantErr: false,
		},
		{
			name: "Empty token",
			args: args{
				headers: http.Header{
					"Authorization": []string{"Bearer"},
				},
				apiType: "Bearer",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Invalid format",
			args: args{
				headers: http.Header{
					"Authorization": []string{"Bearerhello"},
				},
				apiType: "Bearer",
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := auth.GetAuthorizationToken(tt.args.headers, tt.args.apiType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAuthorizationToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("GetAuthorizationToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
