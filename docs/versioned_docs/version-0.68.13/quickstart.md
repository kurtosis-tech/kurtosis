---
title: Quickstart
sidebar_label: Quickstart
slug: /quickstart
---

Introduction
------------
Welcome to the [Kurtosis][homepage] quickstart!

If you arrived here by chance and you're curious as to what Kurtosis _is_, [see here][what-is-kurtosis-explanation].

If you're ready to get going, this guide will take ~15 minutes and will give you basic Kurtosis competency by walking you through building a Kurtosis package.

You need to [have Kurtosis installed][installing-kurtosis-guide] (or [upgraded to latest][upgrading-kurtosis-guide] if you already have it), but you do not need any other knowledge.

:::tip
If you get stuck at any point during this quickstart, there are many, many options available:

- Every Kurtosis command accepts a `-h` flag to print helptext
- The `kurtosis discord` command will open up our Discord
- `kurtosis feedback --github` will take you to opening a Github issue
- `kurtosis feedback --email` will open an email to us
- `kurtosis feedback --calendly` will open a booking link for a personal session with [our cofounder Kevin][kevin-linked]

**Don't suffer in silence - we want to hear from you!**
:::

Hello, World
------------
First, create and `cd` into a directory to hold the project you'll be working on:

```bash
mkdir kurtosis-quickstart && cd kurtosis-quickstart
```

:::tip
All code blocks in this quickstart can be copied by hovering over the block and clicking the clipboard that appears in the right.
:::

Next, create a file called `main.star` inside your new directory with the following contents:

```python
def run(plan, args):
    plan.print("Hello, world")
```

Finally, [run][kurtosis-run-reference] the script.

```bash
kurtosis run --enclave-identifier quickstart main.star
```

Kurtosis will work for a bit, and then deliver you results:

```text
INFO[2023-03-15T04:27:01-03:00] Creating a new enclave for Starlark to run inside...
INFO[2023-03-15T04:27:05-03:00] Enclave 'quickstart' created successfully

> print msg="Hello, world"
Hello, world

Starlark code successfully run. No output was returned.
INFO[2023-03-15T04:27:05-03:00] ===================================================
INFO[2023-03-15T04:27:05-03:00] ||          Created enclave: quickstart          ||
INFO[2023-03-15T04:27:05-03:00] ===================================================
UUID:                                 a78f2ce1ca68
Enclave Name:                         quickstart
Enclave Status:                       RUNNING
Creation Time:                        Wed, 15 Mar 2023 04:27:01 -03
API Container Status:                 RUNNING
API Container Host GRPC Port:         127.0.0.1:62828
API Container Host GRPC Proxy Port:   127.0.0.1:62829

========================================== User Services ==========================================
UUID   Name   Ports   Status
```

Congratulations - you've written your first Kurtosis code!

### Review
:::info
We'll use these "Review" sections to explain what happened in the section. If you just want the action, feel free to skip them.
:::

In this section, we created a `.star` file that prints `Hello, world`. The `.star` extension corresponds to [the Starlark language developed at Google][starlark-github-repo], a dialect of Python for configuring the [Bazel build system][bazel-github]. [Kurtosis uses Starlark for the same purpose of configuring builds][starlark-explanation], except that we're building distributed systems rather than binaries or JARs.

When you ran the Starlark, you got `Created enclave: quickstart`. An [enclave][enclaves-explanation] is a Kurtosis primitive that can be thought of as an ephemeral house for a distributed application. The distributed applications that you define with Starlark will run inside enclaves. 

Enclaves are intended to be easy to create, easy to destroy, cheap to run, and isolated from each other. Use enclaves liberally!

Run Postgres
--------------
The heart of any application is the database. To introduce you to Kurtosis, we'll start by launching a Postgres server using Kurtosis.

Replace the contents of your `main.star` file with the following:

```python
POSTGRES_PORT_ID = "postgres"
POSTGRES_DB = "app_db"
POSTGRES_USER = "app_user"
POSTGRES_PASSWORD = "password"

def run(plan, args):
    # Add a Postgres server
    postgres = plan.add_service(
        "postgres",
        ServiceConfig(
            image = "postgres:15.2-alpine",
            ports = {
                POSTGRES_PORT_ID: PortSpec(5432, application_protocol = "postgresql"),
            },
            env_vars = {
                "POSTGRES_DB": POSTGRES_DB,
                "POSTGRES_USER": POSTGRES_USER,
                "POSTGRES_PASSWORD": POSTGRES_PASSWORD,
            },
        ),
    )
```

You're almost ready to run, but you still have the `quickstart` enclave hanging around from the previous section. [Blow it away][kurtosis-clean-reference] and rerun:

```bash
kurtosis clean -a && kurtosis run --enclave-identifier quickstart main.star
```

:::info
This clean-and-run process will be your dev loop for the rest of the quickstart.
:::

You'll see in the result that the `quickstart` enclave now contains a Postgres instance:

```text
UUID:                                 a30106a0bb87
Enclave Name:                         quickstart
Enclave Status:                       RUNNING
Creation Time:                        Tue, 14 Mar 2023 20:23:54 -03
API Container Status:                 RUNNING
API Container Host GRPC Port:         127.0.0.1:59271
API Container Host GRPC Proxy Port:   127.0.0.1:59272

========================================== User Services ==========================================
UUID           Name       Ports                                                Status
b6fc024deefe   postgres   postgres: 5432/tcp -> postgresql://127.0.0.1:59299   RUNNING
```

### Review
So what actually happened?

1. **Interpretation:** Kurtosis ran your Starlark to build [a plan](https://docs.kurtosis.com/reference/plan) for what you wanted done (in this case, starting a Postgres instance)
1. **Validation:** Kurtosis ran several validations against your plan, including validating that the Postgres image exists
1. **Execution:** Kurtosis executed the validated plan inside the enclave to start a Postgres container

Note that Kurtosis did not execute anything until _after_ Interpretation and Validation completed. You can think of Interpretation and Validation like Kurtosis' "compilation" for your distributed system: we can catch many errors before any containers run, which shortens the dev loop and reduces the resource burden on your machine.

We call this approach [multi-phase runs][multi-phase-runs-reference]. While it has powerful benefits, the stumbling point for new Kurtosis users is that _you cannot reference Execution values like IP address in Starlark_ because they simply don't exist at Interpretation time. We'll see how to work around this limitation later.

Add some data
-------------
A database without data is a fancy heater, so let's add some. 

Our two options for seeding a Postgres database are:

1. Making a sequence of PSQL commands via the `psql` binary
1. Using `pg_restore` to load a package of data

Both are possible in Kurtosis, but for this tutorial we'll do the second one using a seed data TAR of DVD rental information, [courtesy of postgresqltutorial.com](https://www.postgresqltutorial.com/postgresql-getting-started/postgresql-sample-database/). 

Normally seeding a database would require downloading the seed data to your machine, starting Postgres, and writing a pile of Bash to copy the seed data to the Postgres server and run a `pg_restore`. If you forgot to check if the database is available, you may get flakes when you try to use the seeding logic in a test. 

You could try Docker Compose to volume-mount the data TAR into the Postgres server, but you'd still need to handle Postgres availability and sequencing the `pg_restore` afterwards.

By contrast, Kurtosis Starlark scripts can use data as a first-class primitive and sequence tasks such as `pg_restore` into the plan. 

Let's see it in action, and we'll explain what's happening afterwards.

Replace your `main.star` with the following:

```python
data_package_module = import_module("github.com/kurtosis-tech/examples/data-package/main.star")

POSTGRES_PORT_ID = "postgres"
POSTGRES_DB = "app_db"
POSTGRES_USER = "app_user"
POSTGRES_PASSWORD = "password"

SEED_DATA_DIRPATH = "/seed-data"

def run(plan, args):
    # Make data available for use in Kurtosis
    data_package_module_result = data_package_module.run(plan, struct())

    # Add a Postgres server
    postgres = plan.add_service(
        "postgres",
        ServiceConfig(
            image = "postgres:15.2-alpine",
            ports = {
                POSTGRES_PORT_ID: PortSpec(5432, application_protocol = "postgresql"),
            },
            env_vars = {
                "POSTGRES_DB": POSTGRES_DB,
                "POSTGRES_USER": POSTGRES_USER,
                "POSTGRES_PASSWORD": POSTGRES_PASSWORD,
            },
            files = {
                SEED_DATA_DIRPATH: data_package_module_result.files_artifact,
            }
        ),
    )

    # Wait for Postgres to become available
    postgres_flags = ["-U", POSTGRES_USER,"-d", POSTGRES_DB]
    plan.wait(
        service_name = "postgres",
        recipe = ExecRecipe(command = ["psql"] + postgres_flags + ["-c", "\\l"]),
        field = "code",
        assertion = "==",
        target_value = 0,
        timeout = "5s",
    )

    # Load the data into Postgres
    plan.exec(
        service_name = "postgres",
        recipe = ExecRecipe(command = ["pg_restore"] + postgres_flags + [
            "--no-owner",
            "--role=" + POSTGRES_USER,
            SEED_DATA_DIRPATH + "/" + data_package_module_result.tar_filename,
        ]),
    )
```

Next to your `main.star`, add a file called `kurtosis.yml` with the following contents:

```bash
name: "github.com/YOUR-GITHUB-USERNAME/kurtosis-quickstart"
```

Rerun:

```bash
kurtosis clean -a && kurtosis run --enclave-identifier quickstart .
```

(Note that the final argument is now `.` and not `main.star`)

The output should also look more interesting as our plan has grown bigger:

```text
INFO[2023-03-15T04:34:06-03:00] Cleaning enclaves...
INFO[2023-03-15T04:34:06-03:00] Successfully removed the following enclaves:
60601dd9906e40d6af5f16b233a56ae7	quickstart
INFO[2023-03-15T04:34:06-03:00] Successfully cleaned enclaves
INFO[2023-03-15T04:34:06-03:00] Cleaning old Kurtosis engine containers...
INFO[2023-03-15T04:34:06-03:00] Successfully cleaned old Kurtosis engine containers
INFO[2023-03-15T04:34:06-03:00] Creating a new enclave for Starlark to run inside...
INFO[2023-03-15T04:34:10-03:00] Enclave 'quickstart' created successfully
INFO[2023-03-15T04:34:10-03:00] Executing Starlark package at '/tmp/kurtosis-quickstart' as the passed argument '.' looks like a directory
INFO[2023-03-15T04:34:10-03:00] Compressing package 'github.com/YOUR-GITHUB-USERNAME/kurtosis-quickstart' at '.' for upload
INFO[2023-03-15T04:34:10-03:00] Uploading and executing package 'github.com/YOUR-GITHUB-USERNAME/kurtosis-quickstart'

> upload_files src="github.com/kurtosis-tech/examples/data-package/dvd-rental-data.tar"
Files with artifact name 'howling-thunder' uploaded with artifact UUID '32810fc8c131414882c52b044318b2fd'

> add_service service_name="postgres" config=ServiceConfig(image="postgres:15.2-alpine", ports={"postgres": PortSpec(number=5432, application_protocol="postgresql")}, files={"/seed-data": "howling-thunder"}, env_vars={"POSTGRES_DB": "app_db", "POSTGRES_PASSWORD": "password", "POSTGRES_USER": "app_user"})
Service 'postgres' added with service UUID 'f1d9cab2ca344d1fbb0fc00b2423f45f'

> wait recipe=ExecRecipe(command=["psql", "-U", "app_user", "-d", "app_db", "-c", "\\l"]) field="code" assertion="==" target_value=0 timeout="5s"
Wait took 2 tries (1.135498667s in total). Assertion passed with following:
Command returned with exit code '0' and the following output:
--------------------
                                                List of databases
   Name    |  Owner   | Encoding |  Collate   |   Ctype    | ICU Locale | Locale Provider |   Access privileges
-----------+----------+----------+------------+------------+------------+-----------------+-----------------------
 app_db    | app_user | UTF8     | en_US.utf8 | en_US.utf8 |            | libc            |
 postgres  | app_user | UTF8     | en_US.utf8 | en_US.utf8 |            | libc            |
 template0 | app_user | UTF8     | en_US.utf8 | en_US.utf8 |            | libc            | =c/app_user          +
           |          |          |            |            |            |                 | app_user=CTc/app_user
 template1 | app_user | UTF8     | en_US.utf8 | en_US.utf8 |            | libc            | =c/app_user          +
           |          |          |            |            |            |                 | app_user=CTc/app_user
(4 rows)


--------------------

> exec recipe=ExecRecipe(command=["pg_restore", "-U", "app_user", "-d", "app_db", "--no-owner", "--role=app_user", "/seed-data/dvd-rental-data.tar"])
Command returned with exit code '0' with no output

Starlark code successfully run. No output was returned.
INFO[2023-03-15T04:34:21-03:00] ===================================================
INFO[2023-03-15T04:34:21-03:00] ||          Created enclave: quickstart          ||
INFO[2023-03-15T04:34:21-03:00] ===================================================
UUID:                                 995fe0ca69fe
Enclave Name:                         quickstart
Enclave Status:                       RUNNING
Creation Time:                        Wed, 15 Mar 2023 04:34:06 -03
API Container Status:                 RUNNING
API Container Host GRPC Port:         127.0.0.1:62893
API Container Host GRPC Proxy Port:   127.0.0.1:62894

========================================== User Services ==========================================
UUID           Name       Ports                                                Status
f1d9cab2ca34   postgres   postgres: 5432/tcp -> postgresql://127.0.0.1:62914   RUNNING
```

Does our Postgres have data now? Let's find out by logging into the database:

```bash
kurtosis service shell quickstart postgres
```

This will open a shell on the Postgres container. From there, listing the tables in the Postgres...

``` bash
psql -U app_user -d app_db -c '\dt'
```

...will reveal that many new tables now exist:

```text
             List of relations
 Schema |     Name      | Type  |  Owner
--------+---------------+-------+----------
 public | actor         | table | app_user
 public | address       | table | app_user
 public | category      | table | app_user
 public | city          | table | app_user
 public | country       | table | app_user
 public | customer      | table | app_user
 public | film          | table | app_user
 public | film_actor    | table | app_user
 public | film_category | table | app_user
 public | inventory     | table | app_user
 public | language      | table | app_user
 public | payment       | table | app_user
 public | rental        | table | app_user
 public | staff         | table | app_user
 public | store         | table | app_user
(15 rows)
```

Feel free to explore the Postgres container. When you're done run either `exit` or press Ctrl-D.

### Review
So what did we just do?

Kurtosis' first-class data primitive is called a [files artifact][files-artifacts-reference]. Each files artifact is a TGZ of arbitrary files, living inside the enclave. So long as a files artifact exists, Kurtosis knows how to mount its contents on a service. We used this feature to mount the seed data into the Postgres instance via the `ServiceConfig.files` option:

```python
postgres = plan.add_service(
    "postgres",
    ServiceConfig(
        # ...omitted...
        files = {
            SEED_DATA_DIRPATH: data_package_module_result.files_artifact,
        }
    ),
)
```

But where did the data come from? 

There are many ways to create files artifacts in an enclave. The simplest is to upload files from your local machine using [the `kurtosis files upload` command][kurtosis-files-upload-reference]. A more advanced way is to upload files using [the `upload_files` Starlark instruction][upload-files-reference] on the plan.

But... you never downloaded the seed data on your local machine. In fact, you didn't need you to because we leveraged one of the most powerful features of Kurtosis: composition. 

Kurtosis has [a built-in packaging/dependency system][how-do-imports-work-explanation] that allows Starlark code to depend on other Starlark code via Github repositories. When you created the `kurtosis.yml` file, you linked your code into the packaging system: you told Kurtosis that your code is a part of a [Kurtosis package][packages-reference], which allowed your code to depend on external Starlark code.

This line at the top of your `main.star`...

```python
data_package_module = import_module("github.com/kurtosis-tech/examples/data-package/main.star")
```

...created a dependency on [the external Kurtosis package living here][data-package-example]. 

Your code then called that dependency here...

```python
data_package_module_result = data_package_module.run(plan, struct())
```

...which in turn ran [the code in the `main.star` of that external package][data-package-example-main.star]. That Kurtosis package happens to contain [the seed data][data-package-example-seed-tar], and it uses the `upload_data` Starlark instruction on the plan to make the seed data available via a files artifact. From there, all we needed to do was mount it on the `postgres` service.

This ability to modularize your distributed application logic using only a Github repo is one of Kurtosis' most loved features. We won't dive into all the usecases now, but [the examples repo][examples-repo] can serve as a good source of inspiration.


Add an API
----------
Databases don't come alone, however. In this section we'll add a [PostgREST API][postgrest] in front of the database and see how Kurtosis handles inter-service dependencies.

Replace the contents of your `main.star` with this:

```python
data_package_module = import_module("github.com/kurtosis-tech/examples/data-package/main.star")

POSTGRES_PORT_ID = "postgres"
POSTGRES_DB = "app_db"
POSTGRES_USER = "app_user"
POSTGRES_PASSWORD = "password"

SEED_DATA_DIRPATH = "/seed-data"

POSTGREST_PORT_ID = "http"

def run(plan, args):
    # Make data available for use in Kurtosis
    data_package_module_result = data_package_module.run(plan, struct())

    # Add a Postgres server
    postgres = plan.add_service(
        "postgres",
        ServiceConfig(
            image = "postgres:15.2-alpine",
            ports = {
                POSTGRES_PORT_ID: PortSpec(5432, application_protocol = "postgresql"),
            },
            env_vars = {
                "POSTGRES_DB": POSTGRES_DB,
                "POSTGRES_USER": POSTGRES_USER,
                "POSTGRES_PASSWORD": POSTGRES_PASSWORD,
            },
            files = {
                SEED_DATA_DIRPATH: data_package_module_result.files_artifact,
            }
        ),
    )

    # Wait for Postgres to become available
    postgres_flags = ["-U", POSTGRES_USER,"-d", POSTGRES_DB]
    plan.wait(
        service_name = "postgres",
        recipe = ExecRecipe(command = ["psql"] + postgres_flags + ["-c", "\\l"]),
        field = "code",
        assertion = "==",
        target_value = 0,
        timeout = "5s",
    )

    # Load the data into Postgres
    plan.exec(
        service_name = "postgres",
        recipe = ExecRecipe(command = ["pg_restore"] + postgres_flags + [
            "--no-owner",
            "--role=" + POSTGRES_USER,
            SEED_DATA_DIRPATH + "/" + data_package_module_result.tar_filename,
        ]),
    )

    # Add PostgREST
    postgres_url = "postgresql://{}:{}@{}:{}/{}".format(
        "postgres",
        POSTGRES_PASSWORD,
        postgres.ip_address,
        postgres.ports[POSTGRES_PORT_ID].number,
        POSTGRES_DB,
    )
    postgrest = plan.add_service(
        service_name = "postgrest",
        config = ServiceConfig(
            image = "postgrest/postgrest:v10.2.0.20230209",
            env_vars = {
                "PGRST_DB_URI": postgres_url,
                "PGRST_DB_ANON_ROLE": POSTGRES_USER,
            },
            ports = {POSTGREST_PORT_ID: PortSpec(3000, application_protocol = "http")},
        )
    )

    # Wait for PostgREST to become available
    plan.wait(
        service_name = "postgrest",
        recipe = GetHttpRequestRecipe(
            port_id = POSTGREST_PORT_ID,
            endpoint = "/actor?limit=5",
        ),
        field = "code",
        assertion = "==",
        target_value = 200,
        timeout = "5s",
    )
```

Now, run the same dev loop command as before (and don't worry about the result!):

```bash
kurtosis clean -a && kurtosis run --enclave-identifier quickstart .
```

We just got a failure, just like we might when building a real system!

```text
> wait recipe=GetHttpRequestRecipe(port_id="http", endpoint="/actor", extract="") field="code" assertion="==" target_value=200 timeout="5s"
There was an error executing Starlark code
An error occurred executing instruction (number 6) at github.com/ME/kurtosis-quickstart[77:14]:
wait(recipe=GetHttpRequestRecipe(port_id="http", endpoint="/actor", extract=""), field="code", assertion="==", target_value=200, timeout="5s", service_name="postgrest")
 --- at /home/circleci/project/core/server/api_container/server/startosis_engine/startosis_executor.go:62 (StartosisExecutor.Execute.func1) ---
Caused by: Wait timed-out waiting for the assertion to become valid. Waited for '8.183602629s'. Last assertion error was:
<nil>
 --- at /home/circleci/project/core/server/api_container/server/startosis_engine/kurtosis_instruction/wait/wait.go:263 (WaitCapabilities.Execute) ---

Error encountered running Starlark code.
```

Here, Kurtosis is telling us that the `wait` instruction on line `77` of our `main.star` (the one for ensuring PostgREST is up) is timing out.

The enclave state is usually a good place to start. If we look at the bottom of our output we'll see the following state of the enclave:

```text
UUID:                                 5b360f940bcc
Enclave Name:                         quickstart
Enclave Status:                       RUNNING
Creation Time:                        Tue, 14 Mar 2023 22:15:19 -03
API Container Status:                 RUNNING
API Container Host GRPC Port:         127.0.0.1:59814
API Container Host GRPC Proxy Port:   127.0.0.1:59815

========================================== User Services ==========================================
UUID           Name        Ports                                                Status
45b355fc810b   postgres    postgres: 5432/tcp -> postgresql://127.0.0.1:59821   RUNNING
80987420176f   postgrest   http: 3000/tcp                                       STOPPED
```

The problem is clear now: the `postgrest` service status is `STOPPED` rather than `RUNNING`. When we grab the PostgREST logs...

```bash
kurtosis service logs quickstart postgrest
```

:::tip
Kurtosis [has tab-completion][installing-tab-complete-guide] - even for dynamic values like enclave name - to make writing commands like this faster.
:::

...we can see that the PostgREST is dying:

```text
15/Mar/2023:01:15:30 +0000: Attempting to connect to the database...
15/Mar/2023:01:15:30 +0000: {"code":"PGRST000","details":"FATAL:  password authentication failed for user \"postgres\"\n","hint":null,"message":"Database connection error. Retrying the connection."}
15/Mar/2023:01:15:30 +0000: FATAL:  password authentication failed for user "postgres"

postgrest: thread killed
```

Looking back to our Starlark, we can see the problem: we're creating the Postgres database with a user called `app_user`, but we're telling PostgREST to try and connect through a user called `postgres`:

```python
POSTGRES_USER = "app_user"

# ...

def run(plan, args):
    # ...

    # Add a Postgres server
    postgres = plan.add_service(
        "postgres",
        ServiceConfig(
            # ...
            env_vars = {
                # ...
                "POSTGRES_USER": POSTGRES_USER,
                # ...
            },
            # ...
        ),
    )

    # ...

    postgres_url = "postgresql://{}:{}@{}:{}/{}".format(
        "postgres",   # <---------- THE PROBLEM 
        POSTGRES_PASSWORD,
        postgres.ip_address,
        postgres.ports[POSTGRES_PORT_ID].number,
        POSTGRES_DB,
    )
```

Replace that `"postgres"` with `POSTGRES_USER` in your `main.star` file to use the correct username, and then rerun your dev loop:

```bash
kurtosis clean -a && kurtosis run --enclave-identifier quickstart .
```

Now at the bottom of the output we can see that the PostgREST service is `RUNNING` correctly:

```text
UUID:                                 11c0ac047299
Enclave Name:                         quickstart
Enclave Status:                       RUNNING
Creation Time:                        Tue, 14 Mar 2023 22:30:02 -03
API Container Status:                 RUNNING
API Container Host GRPC Port:         127.0.0.1:59876
API Container Host GRPC Proxy Port:   127.0.0.1:59877

========================================== User Services ==========================================
UUID           Name        Ports                                                Status
ce90b471a982   postgres    postgres: 5432/tcp -> postgresql://127.0.0.1:59883   RUNNING
98094b33cd9a   postgrest   http: 3000/tcp -> http://127.0.0.1:59887             RUNNING
```

### Review
In this section, we declared a new PostgREST service with a dependency on the Postgres service.

Yet... PostgREST needs to know the IP address or hostname of the Postgres service, and we said earlier that Starlark (the Interpretation phase) can never know Execution values. How can this be?

Answer: Execution-time values are represented at Interpretation time as [future references][future-references-reference] - special Starlark strings like `{{kurtosis:6670e781977d41409f9eb2833977e9df:ip_address.runtime_value}}` that Kurtosis will replace at Execution time with the actual value. In this case, the `postgres_url` variable here...

```python
postgres_url = "postgresql://{}:{}@{}:{}/{}".format(
    POSTGRES_USER,
    POSTGRES_PASSWORD,
    postgres.ip_address,
    postgres.ports[POSTGRES_PORT_ID].number,
    POSTGRES_DB,
)
```

...used the `postgres.ip_address` and `postgres.ports[POSTGRES_PORT_ID].number` future references returned by adding the Postgres service, so that when `postgres_url` was used as an environment variable during PostgREST startup...

```python
postgrest = plan.add_service(
    service_name = "postgrest",
    config = ServiceConfig(
        # ...
        env_vars = {
            "PGRST_DB_URI": postgres_url, # <-------- HERE
            "PGRST_DB_ANON_ROLE": POSTGRES_USER,
        },
        # ...
    )
)
```

...Kurtosis simply swapped in the correct Postgres container Execution-time values. While future references take some getting used to, [we've found the feedback loop speedup to be very worth it][why-multi-phase-runs-explanation].

Modifying data
--------------
Now that we have an API, we should be able to interact with the data.

[Inspect][kurtosis-enclave-inspect-reference] your enclave:

```bash
kurtosis enclave inspect quickstart
```

:::tip
Tab completion can fill enclave names too.
:::

Notice how Kurtosis automatically exposed the PostgREST container's `http` port to your machine:

```text
28a923400e50   postgrest   http: 3000/tcp -> http://127.0.0.1:59992             RUNNING
```

:::info
In this output, the `http` port is exposed as URL `http://127.0.0.1:59992` but your output will be different.
:::

You can paste the URL from your output into your browser (or Cmd+click it in iTerm) to verify that you are indeed talking to the PostgREST inside your `quickstart` enclave:

```json
{"swagger":"2.0","info":{"description":"","title":"standard public schema","version":"10.2.0.20230209 (pre-release) (a1e2fe3)"},"host":"0.0.0.0:3000","basePath":"/","schemes":["http"],"consumes":["application/json","application/vnd.pgrst.object+json","text/csv"],"produces":["application/json","application/vnd.pgrst.object+json","text/csv"],"paths":{"/":{"get":{"tags":["Introspection"],"summary":"OpenAPI description (this document)","produces":["application/openapi+json","application/json"],"responses":{"200":{"description":"OK"}}}},"/actor":{"get":{"tags":["actor"],"parameters":[{"$ref":"#/parameters/rowFilter.actor.actor_id"},{"$ref":"#/parameters/rowFilter.actor.first_name"},{"$ref":"#/parameters/rowFilter.actor.last_name"},{"$ref":"#/parameters/rowFilter.actor.last_update"},{"$ref":"#/parameters/select"},{"$ref":"#/parameters/order"},{"$ref":"#/parameters/range"},{"$ref":"#/parameters/rangeUnit"},{"$ref":"#/parameters/offset"},{"$ref":"#/parameters/limit"},{"$ref":"#/parameters/preferCount"}], ...
```

Now make a request to insert a row into the database (replacing the `http://127.0.0.1:59992` portion of the URL with the correct URL from your `enclave inspect` output)...

```bash
curl -XPOST -H "content-type: application/json" <YOUR_HTTP_URL>/actor --data '{"first_name": "Kevin", "last_name": "Bacon"}'
```

...and then query for it (again replacing `http://127.0.0.1:59992` with your correct URL)...

```bash
curl -XGET "http://127.0.0.1:<YOUR_URL>/actor?first_name=eq.Kevin&last_name=eq.Bacon"
```

...to get it back:

```text
[{"actor_id":201,"first_name":"Kevin","last_name":"Bacon","last_update":"2023-03-15T02:08:14.315732"}]
```

Of course, it'd be much nicer to formalize this in Kurtosis. Replace your `main.star` with the following:

```python
data_package_module = import_module("github.com/kurtosis-tech/examples/data-package/main.star")

POSTGRES_PORT_ID = "postgres"
POSTGRES_DB = "app_db"
POSTGRES_USER = "app_user"
POSTGRES_PASSWORD = "password"

SEED_DATA_DIRPATH = "/seed-data"

POSTGREST_PORT_ID = "http"

def run(plan, args):
    # Make data available for use in Kurtosis
    data_package_module_result = data_package_module.run(plan, struct())

    # Add a Postgres server
    postgres = plan.add_service(
        "postgres",
        ServiceConfig(
            image = "postgres:15.2-alpine",
            ports = {
                POSTGRES_PORT_ID: PortSpec(5432, application_protocol = "postgresql"),
            },
            env_vars = {
                "POSTGRES_DB": POSTGRES_DB,
                "POSTGRES_USER": POSTGRES_USER,
                "POSTGRES_PASSWORD": POSTGRES_PASSWORD,
            },
            files = {
                SEED_DATA_DIRPATH: data_package_module_result.files_artifact,
            }
        ),
    )

    # Wait for Postgres to become available
    postgres_flags = ["-U", POSTGRES_USER,"-d", POSTGRES_DB]
    plan.wait(
        service_name = "postgres",
        recipe = ExecRecipe(command = ["psql"] + postgres_flags + ["-c", "\\l"]),
        field = "code",
        assertion = "==",
        target_value = 0,
        timeout = "5s",
    )

    # Load the data into Postgres
    plan.exec(
        service_name = "postgres",
        recipe = ExecRecipe(command = ["pg_restore"] + postgres_flags + [
            "--no-owner",
            "--role=" + POSTGRES_USER,
            SEED_DATA_DIRPATH + "/" + data_package_module_result.tar_filename,
        ]),
    )

    # Add PostgREST
    postgres_url = "postgresql://{}:{}@{}:{}/{}".format(
        POSTGRES_USER,
        POSTGRES_PASSWORD,
        postgres.hostname,
        postgres.ports[POSTGRES_PORT_ID].number,
        POSTGRES_DB,
    )
    postgrest = plan.add_service(
        service_name = "postgrest",
        config = ServiceConfig(
            image = "postgrest/postgrest:v10.2.0.20230209",
            env_vars = {
                "PGRST_DB_URI": postgres_url,
                "PGRST_DB_ANON_ROLE": POSTGRES_USER,
            },
            ports = {POSTGREST_PORT_ID: PortSpec(3000, application_protocol = "http")},
        )
    )

    # Wait for PostgREST to become available
    plan.wait(
        service_name = "postgrest",
        recipe = GetHttpRequestRecipe(
            port_id = POSTGREST_PORT_ID,
            endpoint = "/actor?limit=5",
        ),
        field = "code",
        assertion = "==",
        target_value = 200,
        timeout = "5s",
    )

    # Insert data
    if args != None:
        insert_data(plan, args)

def insert_data(plan, data):
    plan.request(
        service_name = "postgrest",
        recipe = PostHttpRequestRecipe(
            port_id = POSTGREST_PORT_ID,
            endpoint = "/actor",
            content_type = "application/json",
            body = json.encode(data),
        )
    )
```

Now clean and run, only this time with extra args to `kurtosis run`:

```bash
kurtosis clean -a && kurtosis run --enclave-identifier quickstart . '[{"first_name":"Kevin", "last_name": "Bacon"}, {"first_name":"Steve", "last_name":"Buscemi"}]'
```

Using the new `http` URL on the `postgrest` service in the output, query for the rows you just added (replacing `http://127.0.0.1:59992` with your URL)...

```bash
curl -XGET "<YOUR_HTTP_URL>/actor?or=(last_name.eq.Buscemi,last_name.eq.Bacon)"
```

...to yield:

```text
[{"actor_id":201,"first_name":"Kevin","last_name":"Bacon","last_update":"2023-03-15T02:29:53.454697"},
 {"actor_id":202,"first_name":"Steve","last_name":"Buscemi","last_update":"2023-03-15T02:29:53.454697"}]
```

### Review
How did this work?

Mechanically, we first create a JSON string of data using Starlark's `json.encode` builtin. Then we use [the `request` Starlark instruction][request-reference] to shove the string at PostgREST, which writes it to the database:

```python
plan.request(
    service_name = "postgrest",
    recipe = PostHttpRequestRecipe(
        port_id = POSTGREST_PORT_ID,
        endpoint = "/actor",
        content_type = "application/json",
        body = json.encode(data),
    )
)
```

At a higher level, Kurtosis automatically deserialized the `[{"first_name":"Kevin", "last_name": "Bacon"}, {"first_name":"Steve", "last_name":"Buscemi"}]` string passed as a parameter to `kurtosis run`, and put the deserialized object in the `args` parameter to the `run` function in `main.star`:

```python
def run(plan, args):
```

Publishing
----------
Congratulations - you've written your very first distributed application in Kurtosis! Now it's time to share it with the world.

The Kurtosis packaging system uses Github as its package repository, just like Go modules. Also like Go modules, Kurtosis packages need their name to match their location on Github.

Update the `name` key of the `kurtosis.yml` file to replace `YOUR-GITHUB-USERNAME` with your Github username:

```yaml
# You'll need to update this
name: "github.com/YOUR-GITHUB-USERNAME/kurtosis-quickstart"
```

Create a new repository on Github, owned by you, named `kurtosis-quickstart` by clicking [here](https://github.com/new).

Hook your Starlark up to the Github repository (replacing `YOUR-GITHUB-USERNAME` with your Github username):

```bash
git init -b main && git remote add origin https://github.com/YOUR-GITHUB-USERNAME/kurtosis-quickstart.git
```

Finally, commit and push your changes:

```bash
git add . && git commit -m "Initial commit" && git push origin main
```

Now that your package is live, any Kurtosis user can run it using their CLI without even cloning your repo:

```bash
kurtosis clean -a && kurtosis run --enclave-identifier quickstart github.com/YOUR-GITHUB-USERNAME/kurtosis-quickstart
```

(Parameterization will still work, of course.)

### Review
Publishing a Kurtosis package is as simple as verifying the `name` key in `kurtosis.yml` matches your repo and pushing it to Github. That package will then be available to every `kurtosis run`, as well as every Starlark script via the `import_module` composition flow.

<!-- TODO TODO TDOO
Testing
=======
- Use the package in some Starlark tests
- Show how we can use the `insert_data` function independently, to operate on an existing environment
- Show the `enclave dump` usecase - very useful for test logs!!!
-->

Conclusion
----------
In this tutorial you have:

- Started a Postgres database
- Seeded it by importing a third-party Starlark package
- Added an API server
- Inserted & queried data via the API
- Parameterized data insertion
- Published your package
- Consumed your package directly from Github

Along the way you've learned about several Kurtosis concepts:

- [The CLI][cli-reference]
- [Enclaves][enclaves-explanation]
- [Starlark][starlark-explanation]
- [Multi-phase runs][multi-phase-runs-reference]
- The [plan][plan-reference]
- [Files artifacts][files-artifacts-reference]
- [Kurtosis packages][packages-reference]
- [Future references][future-references-reference]

Now that you've reached the end, we'd love to hear from you - what went well for you, and what didn't? You can file issues and feature requests on Github...

```bash
kurtosis feedback --github
```

...or you can email us via the CLI...

```bash
kurtosis feedback --email
```

...and you can even schedule a personal session with [our cofounder Kevin][kevin-linked] via:

```bash
kurtosis feedback --calendly
```

We use all feedback to fuel product development, so please don't hesitate to get in touch!

Finally, if liked what you saw and want to engage with us, you can:

- [Star us on Github](https://github.com/kurtosis-tech/kurtosis) (this helps a lot!)
- [Join our Discord](https://discord.com/channels/783719264308953108/783719264308953111) (also available with the `kurtosis discord` CLI command)
- [Reach out to us on Twitter](https://twitter.com/KurtosisTech)

Or you can simply dive deeper into the docs:

- [Read about the architecture][architecture-explanation]
- [Explore the full catalog of Starlark commands][starlark-instructions-reference]
- [Explore the various CLI commands][cli-reference]
- Explore [Kurtosis-provided packages being used in production][kurtosis-managed-packages]
- [Search GitHub for Kurtosis packages in the wild][wild-kurtosis-packages]

<!-- !!!!!!!!!!!!!!!!!!!!!!!!!!! ONLY LINKS BELOW HERE !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! -->

<!--------------------------- Guides ------------------------------------>
[installing-kurtosis-guide]: ./guides/installing-the-cli.md
[upgrading-kurtosis-guide]: ./guides/upgrading-the-cli.md
[installing-tab-complete-guide]: ./guides/adding-tab-completion.md

<!--------------------------- Explanations ------------------------------------>
[architecture-explanation]: ./explanations/architecture.md
[enclaves-explanation]: ./explanations/architecture.md#enclaves
[services-explanation]: ./explanations/architecture.md#services
[starlark-explanation]: ./explanations/starlark.md
[reusable-environment-definitions-explanation]: ./explanations/reusable-environment-definitions.md
[what-is-kurtosis-explanation]: ./explanations/what-is-kurtosis.md
[how-do-imports-work-explanation]: ./explanations/how-do-kurtosis-imports-work.md
[why-multi-phase-runs-explanation]: ./explanations/why-multi-phase-runs.md

<!--------------------------- Reference ------------------------------------>
<!-- CLI Commands Reference -->
[cli-reference]: ./reference/cli/cli.md
[kurtosis-run-reference]: ./reference/cli/run-starlark.md
[kurtosis-clean-reference]: ./reference/cli/clean.md
[kurtosis-clean-reference]: ./reference/cli/clean.md
[kurtosis-enclave-inspect-reference]: ./reference/cli/enclave-inspect.md
[kurtosis-files-upload-reference]: ./reference/cli/files-upload.md

<!-- SL Instructions Reference-->
[starlark-instructions-reference]: ./reference/starlark-instructions.md
[upload-files-reference]: ./reference/starlark-instructions.md#upload_files
[request-reference]: ./reference/starlark-instructions.md#request

<!-- Reference -->
[multi-phase-runs-reference]: ./reference/multi-phase-runs.md
[kurtosis-yml-reference]: ./reference/kurtosis-yml.md
[packages-reference]: ./reference/packages.md
[runnable-packages-reference]: ./reference/packages.md#runnable-packages
[locators-reference]: ./reference/locators.md
[plan-reference]: ./reference/plan.md
[future-references-reference]: ./reference/future-references.md
[files-artifacts-reference]: ./reference/files-artifacts.md

<!--------------------------- Other ------------------------------------>
<!-- Examples repo -->
[examples-repo]: https://github.com/kurtosis-tech/examples
[data-package-example]: https://github.com/kurtosis-tech/examples/tree/main/data-package
[data-package-example-main.star]: https://github.com/kurtosis-tech/examples/blob/main/data-package/main.star
[data-package-example-seed-tar]: https://github.com/kurtosis-tech/examples/blob/main/data-package/dvd-rental-data.tar

<!-- Misc -->
[homepage]: https://kurtosis.com
[kevin-linked]: https://www.linkedin.com/in/kevintoday/
[kurtosis-managed-packages]: https://github.com/kurtosis-tech?q=in%3Aname+package&type=all&language=&sort=
[wild-kurtosis-packages]: https://github.com/search?q=filename%3Akurtosis.yml&type=code
[bazel-github]: https://github.com/bazelbuild/bazel/
[starlark-github-repo]: https://github.com/bazelbuild/starlark
[postgrest]: https://postgrest.org/en/stable/
