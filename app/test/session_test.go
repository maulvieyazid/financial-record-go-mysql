package config

import (
    "os"
    "testing"
)

func restoreEnv(key, val string) {
    if val == "" {
        _ = os.Unsetenv(key)
    } else {
        _ = os.Setenv(key, val)
    }
}

func TestSessionStoreOptions_FromEnv(t *testing.T) {
    tests := []struct {
        name            string
        appSecureCookie string // empty = unset
        appEnv          string // empty = unset
        expectSecure    bool
    }{
        {"default_no_env_vars", "", "", false},
        {"secure_cookie_true", "true", "", true},
        {"secure_cookie_TRUE_case_insensitive", "TRUE", "", true},
        {"app_env_production", "", "production", true},
    }

    for _, tc := range tests {
        tc := tc
        t.Run(tc.name, func(t *testing.T) {
            // Simpan env asli
            origSecure := os.Getenv("APP_SECURE_COOKIE")
            origEnv := os.Getenv("APP_ENV")
            defer restoreEnv("APP_SECURE_COOKIE", origSecure)
            defer restoreEnv("APP_ENV", origEnv)

            // Set/unset sesuai testcase
            if tc.appSecureCookie == "" {
                _ = os.Unsetenv("APP_SECURE_COOKIE")
            } else {
                _ = os.Setenv("APP_SECURE_COOKIE", tc.appSecureCookie)
            }
            if tc.appEnv == "" {
                _ = os.Unsetenv("APP_ENV")
            } else {
                _ = os.Setenv("APP_ENV", tc.appEnv)
            }

            // Reinitialize Store based on current envs
            InitStore()

            if Store == nil {
                t.Fatalf("Store == nil setelah init()")
            }
            if Store.Options == nil {
                t.Fatalf("Store.Options == nil setelah init()")
            }

            if Store.Options.Secure != tc.expectSecure {
                t.Errorf("Secure mismatch: got %v, want %v", Store.Options.Secure, tc.expectSecure)
            }
            if Store.Options.Path != "/" {
                t.Errorf("Path mismatch: got %q, want %q", Store.Options.Path, "/")
            }
            if Store.Options.MaxAge != 3600*24 {
                t.Errorf("MaxAge mismatch: got %d, want %d", Store.Options.MaxAge, 3600*24)
            }
            if !Store.Options.HttpOnly {
                t.Errorf("HttpOnly expected true")
            }
        })
    }
}

func TestConstants(t *testing.T) {
    if SESSION_ID != "finacial_record_okt" {
        t.Errorf("SESSION_ID berubah: got %q", SESSION_ID)
    }
    if FLASH_ID != "flash_logout" {
        t.Errorf("FLASH_ID berubah: got %q", FLASH_ID)
    }
}