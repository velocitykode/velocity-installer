package generator

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	cli "github.com/velocitykode/velocity-cli"
)

type dbEnv struct {
	Connection string
	Host       string
	Port       string
	Database   string
	Username   string
	Password   string
}

func readDBEnv(projectPath string) (dbEnv, error) {
	f, err := os.Open(filepath.Join(projectPath, ".env"))
	if err != nil {
		return dbEnv{}, err
	}
	defer f.Close()

	env := dbEnv{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		switch strings.TrimSpace(key) {
		case "DB_CONNECTION":
			env.Connection = value
		case "DB_HOST":
			env.Host = value
		case "DB_PORT":
			env.Port = value
		case "DB_DATABASE":
			env.Database = value
		case "DB_USERNAME":
			env.Username = value
		case "DB_PASSWORD":
			env.Password = value
		}
	}
	return env, scanner.Err()
}

// ensureDatabaseReady performs a preflight check and attempts to create the
// target database. It returns ready=true when migrations can proceed, and
// ready=false (with a user-facing message already printed) when migrations
// should be skipped. Errors are only returned for unexpected failures like
// being unable to read the .env file.
func ensureDatabaseReady(projectPath string) (ready bool, err error) {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return false, err
	}
	env, err := readDBEnv(absPath)
	if err != nil {
		return false, fmt.Errorf("read .env: %w", err)
	}

	// sqlite is file-based — nothing to preflight.
	if env.Connection == "" || env.Connection == "sqlite" {
		return true, nil
	}

	target := net.JoinHostPort(env.Host, env.Port)
	cli.Info(fmt.Sprintf("Checking %s server at %s", env.Connection, target))
	conn, dialErr := net.DialTimeout("tcp", target, 3*time.Second)
	if dialErr != nil {
		cli.Warning(fmt.Sprintf("%s not reachable at %s", env.Connection, target))
		cli.Muted(fmt.Sprintf("Start %s, create database '%s', then run: cd %s && go run . migrate",
			env.Connection, env.Database, projectPath))
		return false, nil
	}
	_ = conn.Close()
	cli.Success(fmt.Sprintf("%s server reachable", env.Connection))

	cli.Info(fmt.Sprintf("Creating database %s", env.Database))
	created, createErr := createDatabase(env)
	if createErr != nil {
		cli.Warning(fmt.Sprintf("Could not create database %s: %s", env.Database, createErr))
		cli.Muted(fmt.Sprintf("Create it manually, then run: cd %s && go run . migrate", projectPath))
		return false, nil
	}
	if created {
		cli.Success(fmt.Sprintf("Database %s created", env.Database))
	} else {
		cli.Success(fmt.Sprintf("Database %s already exists", env.Database))
	}
	return true, nil
}

func createDatabase(env dbEnv) (created bool, err error) {
	switch env.Connection {
	case "postgres":
		return createPostgresDatabase(env)
	case "mysql":
		return createMySQLDatabase(env)
	default:
		return false, fmt.Errorf("unsupported database %q", env.Connection)
	}
}

func createPostgresDatabase(env dbEnv) (bool, error) {
	if _, err := exec.LookPath("psql"); err != nil {
		return false, fmt.Errorf("psql not found on PATH")
	}

	baseEnv := append(os.Environ(), "PGPASSWORD="+env.Password)

	check := exec.Command("psql",
		"-h", env.Host, "-p", env.Port, "-U", env.Username,
		"-d", "postgres",
		"-tAc", fmt.Sprintf("SELECT 1 FROM pg_database WHERE datname='%s'", env.Database),
	)
	check.Env = baseEnv
	checkOut, checkErr := check.CombinedOutput()
	if checkErr != nil {
		return false, fmt.Errorf("%s", firstLine(checkOut, checkErr))
	}
	if strings.TrimSpace(string(checkOut)) == "1" {
		return false, nil
	}

	create := exec.Command("psql",
		"-h", env.Host, "-p", env.Port, "-U", env.Username,
		"-d", "postgres",
		"-c", fmt.Sprintf(`CREATE DATABASE "%s"`, env.Database),
	)
	create.Env = baseEnv
	if out, err := create.CombinedOutput(); err != nil {
		return false, fmt.Errorf("%s", firstLine(out, err))
	}
	return true, nil
}

func createMySQLDatabase(env dbEnv) (bool, error) {
	if _, err := exec.LookPath("mysql"); err != nil {
		return false, fmt.Errorf("mysql not found on PATH")
	}

	cmd := exec.Command("mysql",
		"-h", env.Host, "-P", env.Port, "-u", env.Username,
		"-N", "-B",
		"-e", fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", env.Database),
	)
	cmd.Env = append(os.Environ(), "MYSQL_PWD="+env.Password)
	if out, err := cmd.CombinedOutput(); err != nil {
		return false, fmt.Errorf("%s", firstLine(out, err))
	}
	// IF NOT EXISTS hides whether it already existed; treat as "exists" to keep
	// messaging honest rather than claiming we created it.
	return false, nil
}

func firstLine(out []byte, err error) string {
	msg := strings.TrimSpace(string(out))
	if msg == "" {
		return err.Error()
	}
	if idx := strings.IndexByte(msg, '\n'); idx >= 0 {
		msg = msg[:idx]
	}
	return msg
}
