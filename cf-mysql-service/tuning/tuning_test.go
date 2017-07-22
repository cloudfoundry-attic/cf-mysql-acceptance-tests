package tuning_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"
	_ "github.com/go-sql-driver/mysql"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MySQL Server Tuning Configuration", func() {
	var (
		db *sql.DB
	)

	BeforeEach(func() {
		standalone := helpers.TestConfig.Standalone
		connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/",
			standalone.MySQLUsername,
			standalone.MySQLPassword,
			standalone.Host,
			standalone.Port)

		var err error
		db, err = sql.Open("mysql", connectionString)
		Expect(err).ToNot(HaveOccurred())

		err = db.Ping()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := db.Close()
		Expect(err).ToNot(HaveOccurred())
	})

	It("Correctly sets MySQL internal variables based on values in the manifest", func() {
		rows, err := db.Query("SHOW VARIABLES;")
		Expect(err).ToNot(HaveOccurred())

		mysqlVariables := map[string]string{}

		defer rows.Close()
		for rows.Next() {
			var name, value string
			err = rows.Scan(&name, &value)
			Expect(err).ToNot(HaveOccurred())
			mysqlVariables[name] = value
		}

		err = rows.Err()
		Expect(err).ToNot(HaveOccurred())

		buf, err := ioutil.ReadFile(helpers.TestConfig.Tuning.ExpectationFilePath)
		Expect(err).ToNot(HaveOccurred())

		compareConfig := map[string]string{}

		if err := json.Unmarshal(buf, &compareConfig); err != nil {
			Expect(err).ToNot(HaveOccurred())
		}

		for k, v := range compareConfig {
			Expect(mysqlVariables[k]).To(Equal(v), fmt.Sprintf("mismatch in %v", k))
		}
	})
})
