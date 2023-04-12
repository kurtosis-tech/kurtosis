---
title: How To: programmatically seed MySQL databases for integration tests
sidebar_label: Programmatically seeding MySQL databases
slug: /how-to-seed-mysql
toc_max_heading_level: 2
sidebar_position: 6
---

Introduction
------------
In this guide, you will programmatically seed a MySQL database for integration tests in an ExpressJS application. ExpressJS can be replaced with any containerized microservice framework you wish.

In general, to run a set of integration tests, an application must satisfy a series of requirements:

- Our test environments need to be instantiated and run the same way every test, or else they’ll be flaky.
- The seed data can easily be changed for our application across tests, so we can test across different scenarios.
- We need to be able to easily link our databases into our microservices in different tests - so that teams building different features, or different microservices, can reference the databases as well without redoing work.
- Ideally, we seed our databases in a way that is reproducible, parameterizable, and composable. One way to do this would involve writing configuration files, code, and shell scripts over Docker, or on bare metal - in which case we’d have to build these guarantees into our scripts.

However, in this guide, we’re going to use a free tool, Kurtosis, that already has reproducibility, composability, and parameterizability built-in. Kurtosis is a composable build system for writing reproducible multi-container test environments, and can run on your own laptop or in your favorite CI provider.

# How to:

First, install Kurtosis following the instructions at the [docs](https://docs.kurtosis.com/install). Kurtosis uses the Starlark language to programmatically define our system, seed our data, and test behavior.

Now we will create a Starlark Package to abstract away the interface with MySQL. We will start by creating a repository named "mysql-package".

Inside that repository, we will write a `mysql-package/kurtosis.yml` file to declare where this package will live. In our case, it will look like this:


```yml
name: github.com/kurtosis-tech/mysql-package
```

After that, we will create a file for the main implementation named "mysql-package/main.star". There, we want to first initialize the MySQL service by defining the required environment variables, exposing the ports, and allowing users to pass an init db script:

```python
MYSQL_IMAGE = "mysql:8.0.32"
ROOT_DEFAULT_PASSWORD = "root"
MYSQL_DEFAULT_PORT = 3036


def create_database(plan, database_name, database_user, database_password, seed_script_artifact = None):
    files = {}
    # Given that scripts on /docker-entrypoint-initdb.d/ are executed sorted by filename
    if seed_script_artifact != None:
        files["/docker-entrypoint-initdb.d"] = seed_script_artifact
    service_name = "mysql-{}".format(database_name)
    # Add service
    mysql_service = plan.add_service(
        name = service_name,
        config = ServiceConfig(
            image = MYSQL_IMAGE,
            ports = {
                "db": PortSpec(
                    number = MYSQL_DEFAULT_PORT,
                    transport_protocol = "TCP",
                    application_protocol = "http",
                ),
            },
            files = files,
            env_vars = {
                "MYSQL_ROOT_PASSWORD": ROOT_DEFAULT_PASSWORD,
                "MYSQL_DATABASE": database_name,
                "MYSQL_USER": database_user,
                "MYSQL_PASSWORD":  database_password,
            },
        )
    )
. . .
```

We also want to wait until MySQL is available before moving forward
```python
  . . .
  # Wait for MySQL to become available
   plan.wait(
       service_name = service_name,
       recipe = ExecRecipe(command = ["mysql", "-u", database_user, "-p{}".format(database_password), database_name]),
       field = "code",
       assertion = "==",
       target_value = 0,
       timeout = "30s",
    )
   . . .
```

Last but not least, we would like to return a struct that encapsulates all the required database information for later usage by this and other packages.

```python
. . .
   return {
       service=mysql_service,
       name=database_name,
       user=database_user,
       password=database_password,
   }
. . .
```

Note that we made this function generic enough so any team can leverage this package, independent of the application, by passing the desired database name, credentials, and setup script, abstracting away MySQL specifics and only interacting with the result of create_database.

We would also like to offer a way for consumers of this package to run a SQL query after database creation. We can use the result of create_database as input, since this encapsulates all details of the database. This is what it looks like:


```python
def run_sql(plan, database, sql_query):
   exec_result = plan.exec(
       service_name = database.service.name,
       recipe = ExecRecipe(command = ["sh", "-c", "mysql -u {} -p{} -e '{}' {}".format(database.user, database.password, sql_query, database.name)]),
   )
   return exec_result["output"]
```

The final script looks like this: [mysql-package/mysql.star](https://github.com/kurtosis-tech/mysql-package/blob/main/mysql.star)

With the MySQL package written, we will now write a package that consumes it. In our case, we will write a package that spins up a database pre-populated with blog data to test our application.

We will start by creating a repository called "blog-seeder". Inside that repository, we will again write a "blog-seeder/kurtosis.yml" file. Other than that, we will include two SQL files: "setup.sql" to create tables and "seed.sql" to populate them.

```sql
% setup.sql
CREATE TABLE user (
   user_id INT NOT NULL,
   first_name VARCHAR(255) NOT NULL,
   PRIMARY KEY (user_id)
);
CREATE TABLE post (
   post_id INT NOT NULL AUTO_INCREMENT,
   content VARCHAR(255) NOT NULL,
   author_user_id INT NOT NULL,
   PRIMARY KEY (post_id),
   FOREIGN KEY (author_user_id) REFERENCES user(user_id)
);

% seed.sql
INSERT INTO user (user_id, first_name) VALUES (1, "Peter");
INSERT INTO user (user_id, first_name) VALUES (2, "Jorge");
INSERT INTO post (post_id, content, author_user_id) VALUES (0, "Lorem ipsum dolor sit amet", 2);
```

Now we will write the main file containing the logic of loading these files, in our case it looks like:

```python
mysql = import_module("github.com/kurtosis-tech/mysql-package/mysql.star")
SELECT_SQL_EXAMPLE = "SELECT * FROM post"

def run(plan, args):
    setup_sql = plan.upload_files(
        src = "github.com/kurtosis-tech/awesome-kurtosis/blog-mysql-seed/setup.sql",
    )
    seed_sql = read_file(
        src = "github.com/kurtosis-tech/awesome-kurtosis/blog-mysql-seed/seed.sql",
    )
    db = mysql.create_database(plan, “db”, “user”, “pass”, seed_script_artifact = setup_sql)
    mysql.run_sql(plan, db, seed_sql)
    mysql.run_sql(plan, db, SELECT_SQL_EXAMPLE)
```
This script can be run using the following command line from within “blog-seeder” repo: `kurtosis run .`

Ideally, we would also like to allow callers of this package to parameterize the values of "db", "user", and "password". This way, if the application code changes and expects, for example, a different "db" name, we can easily modify the behavior of the package.

```python
mysql = import_module("github.com/kurtosis-tech/mysql-package/mysql.star")
SELECT_SQL_EXAMPLE = "SELECT * FROM post"

def run(plan, args):
   setup_sql = plan.upload_files(
       src = "github.com/kurtosis-tech/awesome-kurtosis/blog-mysql-seed/setup.sql",
   )
   seed_sql = read_file(
       src = "github.com/kurtosis-tech/awesome-kurtosis/blog-mysql-seed/seed.sql",
   )
   db = mysql.create_database(plan, “db”, “user”, “pass”, seed_script_artifact = setup_sql)
   mysql.run_sql(plan, db, seed_sql)
   mysql.run_sql(plan, db, SELECT_SQL_EXAMPLE)
```

This script can be run using the following command line from within “blog-seeder” repo: `kurtosis run .`

Ideally, we would also like to allow callers of this package to parameterize the values of "db", "user", and "password". This way, if the application code changes and expects, for example, a different "db" name, we can easily modify the behavior of the package.

Final script looks like this: [blog-mysql-seed/main.star](https://github.com/kurtosis-tech/awesome-kurtosis/blob/main/blog-mysql-seed/main.star)

This script can be run using the following command line from within “blog-seeder” repo:

```bash
kurtosis run . '{"username": "abc", "password": "123", "database": "bd"}'
```

From the output

```bash
\> exec recipe=ExecRecipe(command=["sh", "-c", "mysql -u hi -pbye -e 'SELECT * FROM post' bd"])
```

Command returned with exit code '0' and the following output:
```console
--------------------
mysql: [Warning] Using a password on the command line interface can be insecure.
post_id content author_user_id
1       Lorem ipsum dolor sit amet      2

--------------------

Starlark code successfully run. No output was returned.
```