sqltest
=======

Check database schema of your Go app behaves as expected.

Example usage
-------------

In example we use `DefaultFileSet`, which finds .sql files in testdata directory.

Structure of files in this example:

```
migrations/schema_test.go
migrations/initial.sql
migrations/testdata/emp_log_salary.sql
```

```go
// migrations/schema_test.go
package migrations

import (
	"testing"

	"github.com/shagohead/sqltest"
	"github.com/shagohead/sqltestpgx"
)

// TestSchema tests database schema with queries from testdata/*.sql
func TestSchema(t *testing.T) {
	set, err := sqltest.DefaultFileSet()
	if err != nil {
		t.Fatal(err)
	}
	for name, test := range set.All() {
		t.Run(name, func(t *testing.T) {
			// dbtest.StartTx is a helper which creates and rollbacks transactions for tests.
			err := test.Run(sqltestpgx.Tx(dbtest.StartTx(t)))
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
```

Database migration in which table emp_log populated by trigger on writes in emp table.

```sql
-- migrations/initial.sql
CREATE TABLE emp (
  user_id integer PRIMARY KEY GENERATED AS DEFAULT BY IDENTITY,
  salary integer NOT NULL
);

CREATE TABLE emp_log (
  id bigint PRIMARY KEY GENERATED AS DEFAULT BY IDENTITY,
  user_id integer NOT NULL,
  salary integer NOT NULL
);

CREATE FUNCTION log_emp() RETURNS trigger AS $$
BEGIN
    IF (TG_OP = 'INSERT' OR OLD.salary <> NEW.salary) THEN
        INSERT INTO emp_log (user_id, salary) VALUES (NEW.user_id, NEW.salary);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER log_emp AFTER INSERT OR UPDATE OF salary ON log_emp FOR EACH ROW EXECUTE FUNCTION log_emp();
```

Test file contains queries separated by `;\n`.

For testing purposes we have two statements: `define` and `assert`.
First one declares and defines named query.
Second one calls that query by its name and compare representation of results.

If query do not use that keywords, it just invokes and checks for error occur.

```sql
-- migrations/emp_log_salary.sql
define get_last_log
SELECT user_id, salary FROM emp_log ORDER BY id DESC;

INSERT INTO emp VALUES (1, 125800);
assert get_last_log [1 125800];

INSERT INTO emp VALUES (2, 220000);
assert get_last_log [2 220000];

UPDATE emp SET salary 157000 WHERE user_id = 1;
assert get_last_log [1 157000];
```
