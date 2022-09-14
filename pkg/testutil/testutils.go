/*
Copyright 2021 Adevinta
*/

package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"sync"

	// This package is intended to be used by tests in other packages so they don't have to interact
	// directly with the db so makes sense to import the driver here.
	_ "github.com/lib/pq"

	"github.com/go-kit/kit/log"
	testfixtures "gopkg.in/testfixtures.v2"

	"github.com/adevinta/vulcan-api/pkg/api"
)

const (
	TestDBName     = "vulcanito_test"
	TestDBUser     = "vulcanito_test"
	TestDBPassword = "vulcanito_test"
	DBdialect      = "postgres"
)

var (
	TestDBconnString      = fmt.Sprintf("user=%s password=%s sslmode=disable dbname=%s", TestDBUser, TestDBPassword, TestDBName)
	dbconnStringWithoutDB = fmt.Sprintf("user=%s password=%s sslmode=disable", TestDBUser, TestDBPassword)
	setupDBOnce           sync.Once
	setupDBError          error
)

// SetupDB initializes the db to be used in tests.
func SetupDB(dbDirPath string) error {
	setupDBOnce.Do(func() {
		setupDBError = setupDB(dbDirPath)
	})
	return setupDBError
}

func setupDB(dbDirPath string) error {
	err := ensureDB()
	if err != nil {
		return err
	}
	err = runFlywayCmd(dbDirPath, "clean")
	if err != nil {
		return err
	}
	return runFlywayCmd(dbDirPath, "migrate")
}

func ensureDB() error {
	db, err := sql.Open(DBdialect, dbconnStringWithoutDB)
	if err != nil {
		return nil
	}
	defer db.Close() // nolint: errcheck
	r, err := db.Exec("select  * from pg_database where datname = $1", TestDBName)
	if err != nil {
		return err

	}
	affected, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		// Create the database.
		// The postgres driver doesn't support params in a query that creates a db.
		// We have to use string concatenation to build the statement but we are not vulnerable to a SQL injection because
		// this function should only be executed under a test and, in any case, the db name is defined in a constant.
		_, err := db.Exec("CREATE DATABASE " + TestDBName)
		if err != nil {
			return err
		}
	}
	return nil
}

func runFlywayCmd(dbDirPath, flywayCommand string) error {

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	dir := path.Join(wd, dbDirPath)
	cmdName := "docker"
	cmdArgs := []string{
		"run",
		"--net=host",
		"-v",
		dir + ":/scripts",
		"flyway",
		"-user=" + TestDBUser,
		"-password=" + TestDBPassword,
		"-url=jdbc:postgresql://localhost:5432/" + TestDBName,
		"-baselineOnMigrate=true",
		"-cleanDisabled=false",
		"-locations=filesystem:/scripts/",
		flywayCommand}

	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error executing flyway command, command output:\n%s.\n Error:\n %s", output, err)
	}
	return nil
}

// LoadFixtures ...
func LoadFixtures(fixturesDir string) error {
	db, err := sql.Open(DBdialect, TestDBconnString)
	if err != nil {
		return err
	}
	defer db.Close() // nolint: errcheck
	fixtures, err := testfixtures.NewFolder(db, &testfixtures.PostgreSQL{}, fixturesDir)
	if err != nil {
		return err
	}
	return fixtures.Load()
}

// CreateTestDatabase builds an empty database in the default local test
// server. The name of the DB will be "vulcanito_<name>_test", where <name>
// corresponds to the name of the function calling this one.
func CreateTestDatabase(name string) (string, error) {
	dialect := "postgres"
	dsn := "host=localhost port=5432 user=vulcanito_test dbname=vulcanito_test password=vulcanito_test sslmode=disable"

	db, err := sql.Open(dialect, dsn)
	if err != nil {
		return "", err
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s;", name))
	if err != nil {
		return "", err
	}
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s WITH TEMPLATE vulcanito OWNER vulcanito_test;", name))
	if err != nil {
		return "", err
	}
	dsn = fmt.Sprintf("host=localhost port=5432 user=vulcanito_test dbname=%v password=vulcanito_test sslmode=disable", name)
	return dsn, nil

}

// DBNameForFunc creates the name of a test DB for a function that is calling
// this one the number of levels above in the calling tree equal to the
// specified depth. For instance if a function named FuncA calls function,
// called FuncB that in turn makes the following call: DBNameForFunc(2), this
// function will return the following name: vulcanito_FuncA_test.
func DBNameForFunc(depth int) string {
	pc, _, _, _ := runtime.Caller(depth)
	callerName := strings.Replace(runtime.FuncForPC(pc).Name(), ".", "_", -1)
	callerName = strings.Replace(callerName, "-", "_", -1)
	parts := strings.Split(callerName, "/")
	return strings.ToLower(fmt.Sprintf("vulcanito_%s_test", parts[len(parts)-1]))
}

// PrepareDatabaseLocal creates a new local test database for the calling
// function and populates it the fixtures in the specified path.
func PrepareDatabaseLocal(fixturesPath string, f func(pDialect, connectionString string, logger log.Logger, logMode bool, defaults map[string][]string) (api.VulcanitoStore, error)) (api.VulcanitoStore, error) {
	dbName := DBNameForFunc(2)
	dsnLocal, err := CreateTestDatabase(dbName)
	if err != nil {
		return nil, err
	}

	err = loadFixtures(fixturesPath, dsnLocal)
	if err != nil {
		return nil, err
	}

	testStoreLocal, err := f("postgres", dsnLocal, log.NewNopLogger(), false, map[string][]string{})
	if err != nil {
		return nil, err
	}
	return testStoreLocal, nil
}

// loadFixtures DESTROYS ALL THE DATA in the database pointed by the specified
// dsn and loads the fixtures stored in the specified path into it.
func loadFixtures(fixturesPath string, dsn string) error {
	dbLocal, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer dbLocal.Close()

	fixturesLocal, err := testfixtures.NewFolder(dbLocal, &testfixtures.PostgreSQL{}, fixturesPath)
	if err != nil {
		return err
	}

	return fixturesLocal.Load()
}

func ErrToStr(err error) string {
	result := ""
	if err != nil {
		result = err.Error()
	}
	return result
}
