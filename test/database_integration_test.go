package config

import (
    "database/sql"
    "os"
    "testing"

    _ "github.com/go-sql-driver/mysql"
    "github.com/spf13/viper"
)

// These tests are integration tests that attempt to connect to the real
// MySQL database defined in `app.conf.json`. They are skipped by default
// unless the environment variable `RUN_DB_TEST` is set to "true".

func shouldRunDBTests() bool {
    return os.Getenv("RUN_DB_TEST") == "true"
}

func loadConfigForTest(t *testing.T) {
    // Initialize configuration: this will load .env (if present) and
    // optionally app.conf.json. Environment variables override file.
    InitConfiguration()
}

func openDBFromViper(t *testing.T) *sql.DB {
    dbUser := viper.GetString("DATABASE.USER")
    dbPass := viper.GetString("DATABASE.PASSWORD")
    dbName := viper.GetString("DATABASE.NAME")
    dbHost := viper.GetString("DATABASE.HOST")
    dbPort := viper.GetString("DATABASE.PORT")
    dbDriver := viper.GetString("DATABASE.DRIVER")

    dsn := dbUser + ":" + dbPass + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName + "?parseTime=true&loc=Asia%2FJakarta"

    db, err := sql.Open(dbDriver, dsn)
    if err != nil {
        t.Fatalf("gagal membuka koneksi database: %v", err)
    }

    if err := db.Ping(); err != nil {
        db.Close()
        t.Fatalf("gagal ping database: %v", err)
    }

    return db
}

func TestDatabase_Connection_And_Tables(t *testing.T) {
    if !shouldRunDBTests() {
        t.Skip("skipping DB integration tests; set RUN_DB_TEST=true to run")
    }

    loadConfigForTest(t)
    db := openDBFromViper(t)
    defer db.Close()

    // Check tables exist: users and record
    tests := []struct{
        name string
        table string
    }{
        {"users table", "users"},
        {"record table", "record"},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            var count int
            // information_schema is portable for MySQL
            err := db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = ?", viper.GetString("DATABASE.NAME"), tc.table).Scan(&count)
            if err != nil {
                t.Fatalf("gagal cek table %s: %v", tc.table, err)
            }
            if count == 0 {
                t.Fatalf("table %s tidak ditemukan di database %s", tc.table, viper.GetString("DATABASE.NAME"))
            }
        })
    }
}

func TestDatabase_TablesHaveRows(t *testing.T) {
    if !shouldRunDBTests() {
        t.Skip("skipping DB integration tests; set RUN_DB_TEST=true to run")
    }

    loadConfigForTest(t)
    db := openDBFromViper(t)
    defer db.Close()

    // We check row counts but do not fail the test if zero; we log the result.
    checks := []struct{
        table string
    }{
        {"users"},
        {"record"},
    }

    for _, c := range checks {
        t.Run("rows_in_"+c.table, func(t *testing.T) {
            var cnt int64
            err := db.QueryRow("SELECT COUNT(*) FROM "+c.table).Scan(&cnt)
            if err != nil {
                t.Fatalf("gagal hitung rows di table %s: %v", c.table, err)
            }
            if cnt == 0 {
                t.Logf("table %s kosong (0 rows)", c.table)
            } else {
                t.Logf("table %s memiliki %d baris", c.table, cnt)
            }
        })
    }
}
