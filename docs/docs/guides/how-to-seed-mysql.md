---
title: How to programmatically seed MySQL databases for integration tests
sidebar_label: Programmatically seeding MySQL databases
slug: /how-to-seed-mysql
toc_max_heading_level: 2
sidebar_position: 6
---

Introduction
------------
In this guide, you will programmatically seed a MySQL database for integration tests in an ExpressJS application. ExpressJS can be replaced with any containerized microservice framework you wish to use.

Specifically, to run a set of integration tests, an application must satisfy a series of requirements:

- Our test environments need to be instantiated and run the same way for every test, or else they’ll be flaky.
- The seed data can easily be changed for our application across tests, so we can test across different scenarios.
- We need to be able to easily link our databases into our microservices in different tests - so that teams building different features, or different microservices, can reference the databases as well without redoing work.
- Ideally, we seed our databases in a way that is reproducible, parameterizable, and composable. One way to do this would involve writing configuration files, code, and shell scripts over Docker, or on bare metal - in which case we’d have to build these guarantees into our scripts.

However, in this guide, we’re going to use a free tool, [Kurtosis][kurtosis-website], that already has reproducibility, composability, and parameterizability built-in. Kurtosis is a composable build system for writing reproducible multi-container test environments, and can run on your own laptop or in your favorite CI provider.

Setup
-----
Before you proceed, make sure you have:
* [Installed and started the Docker engine on your local machine][starting-docker]
* [Installed the Kurtosis CLI (or upgraded it to the latest release, if you already have the CLI installed)][installing-the-cli]

:::tip Use the Starlark VS Code Extension
Feel free to use the [official Kurtosis Starlark VS Code extension][vscode-plugin] when writing Starlark with VSCode for features like syntax highlighting, method signature suggestions, hover preview for functions, and auto-completion for Kurtosis custom types.
:::

Instantiate a MySQL instance
----------------------------
A Kurtosis package (written in [Starlark][starlark]) will be used to abstract away the interface with MySQL. To create one from the command line, begin by creating a simple folder:

```bash
mkdir mysql-package && cd mysql-package
```

Then, inside the "mysql-package" folder, create a `mysql-package/kurtosis.yml` file to declare where this package will live. This converts your directory into a [Kurtosis package](https://docs.kurtosis.com/concepts-reference/packages). 

The content of your `kurtosis.yml` file should look like:

```yml
name: github.com/kurtosis-tech/mysql-package
```

Next, in the same folder, create a file for the main implementation named `mysql.star`. In this `.star` file, the MySQL service will be initialized by defining the required environment variables, exposing the ports, and allowing users to pass an `init db` script.

Begin your `mysql.star` file with the following:
```python
MYSQL_IMAGE = "mysql:8.0.32"
ROOT_DEFAULT_PASSWORD = "root"
MYSQL_DEFAULT_PORT = 3036

def create_database(plan, database_name, database_user, database_password, seed_script_artifact = None):
    files = {}
    # Given that scripts on /docker-entrypoint-initdb.d/ are executed sorted by filename
    if seed_script_artifact != None:
        files["/docker-entrypoint-initdb.d"] = seed_script_artifact

    # Give your service the name passed in by a user
    service_name = "mysql-{}".format(database_name)

    # Define a readiness check for your database to ensure the service is ready to receive traffic and connections after starting
    db_ready_check = ReadyCondition(
        recipe = ExecRecipe(
            command = ["mysql", "-u", database_user, "-p{}".format(database_password), database_name]
            ),
        field = "code",
        assertion = "==",
        target_value = 0,
        timeout = "30s",
    )
    
    # Add MySQL service
    mysql_service = plan.add_service(
        name = service_name,
        # Define the service configurations
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
            ready_conditions = db_ready_check,
        ),
    )
```
Last but not least, add the following to return a struct that encapsulates all the required database information for later usage by this and other packages:

```python
. . .
   return {
       service = mysql_service,
       name = database_name,
       user = database_user,
       password = database_password,
   }
. . .
```
:::note
Note that we made this function generic enough so any team can leverage this package, independent of the application, by passing the desired database name, credentials, and setup script, abstracting away MySQL specifics and only interacting with the result of create_database.
:::

You now have all the instructions needed for Kurtosis to instantiate a MySQL database. However, a databse is not very useful unless its queryable by other users. For consumers of this package to run a SQL query after database creation, add the following block that uses the result of `create_database` as an input:

```python
def run_sql(plan, database, sql_query):
   exec_result = plan.exec(
       service_name = database.service.name,
   )
   return exec_result["output"]
```
And thats it! What you've done here is write a set of instructions that anyone can consume and use, with Kurtosis, to instantiate a MySQL database and make a simple query against it to test functionality. Your local `mysql.star` should look identical to this one on [Github](https://github.com/kurtosis-tech/mysql-package/blob/main/mysql.star)

Seed the database with data upon startup
----------------------------------------
With the MySQL package ready to go, you will now write a new package that seeds it with data. In this specific case, you will write a package that spins up a database pre-populated with blog data to test our application.

Start by creating a folder called "blog-mysql-seed" next to your "mysql-package" folder. Inside your "blog-mysql-seed" folder, create a `blog-mysql-seed/kurtosis.yml` file with the following contents:
```yml
name: github.com/kurtosis-tech/awesome-kurtosis/blog-mysql-seed
```

Next, create 2 `.sql` files: "setup.sql" to create tables and "seed.sql" to populate them, and save them in the same folder.
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

Finally, create a `main.star` file with the following contents:

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
    db = mysql.create_database(plan, args["database"], args["username"], args["password"], seed_script_artifact = setup_sql)
    mysql.run_sql(plan, db, seed_sql)
    mysql.run_sql(plan, db, SELECT_SQL_EXAMPLE)
```

This new `main.star` file can be run from the "blog-mysql-seed" folder. The Kurtosis package is actually parameterized so that if the application code changes and expects, for example, a different "db" name, you can easily modify the behavior of the package. 

Your final script should look like this: [blog-mysql-seed/main.star](https://github.com/kurtosis-tech/awesome-kurtosis/blob/main/blog-mysql-seed/main.star)

To show that it works, you can run the following (from the "blog-mysql-seed" folder):
```bash
kurtosis run main.star '{"database": "bd", "username": "abc", "password": "123"}'
```

Your output should look something like this:
```bash
> exec recipe=ExecRecipe(command=["sh", "-c", "mysql -u hi -pbye -e 'SELECT * FROM post' bd"])
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

Congrats! You've just written two Kurtosis packages: one to instantiate the database and another to seed it with tables and data. 

To recap, the [`mysql-package/mysql.star`](https://github.com/kurtosis-tech/mysql-package/blob/main/mysql.star) file on Github can be used as a building block for downstream Kurtosis packages to consume. You used it to set up an empty MySQL database instance while the [blog-mysql-seed/main.star](https://github.com/kurtosis-tech/awesome-kurtosis/blob/main/blog-mysql-seed/main.star) package was used to create a database and seed it with data.

<!--------------------------REFERENCES------------------------------>
[installing-the-cli]: ./installing-the-cli.md#ii-install-the-cli
[starting-docker]: ./installing-the-cli.md#i-install--start-docker
[vscode-plugin]: https://marketplace.visualstudio.com/items?itemName=Kurtosis.kurtosis-extension
[kurtosis-website]: https://www.kurtosis.com/
[starlark]: ../concepts-reference/starlark.md