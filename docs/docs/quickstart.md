--
title: Quickstart
sidebar_label: Quickstart
slug: /quickstart
--

Introduction
============
Welcome to the Kurtosis quickstart!

If arrived here by chance and you're curious as to what Kurtosis _is_, [see here](TODO).

If you're ready to get going, here's what you need to know:

1. This guide will give you basic Kurtosis competency by building a Kurtosis package, step by step.
1. You need to [have Kurtosis and its prerequisites installed](TODO), but you do not need any other knowledge.
1. 


- Explain how they can get help
    - `kurtosis docs`
    - `kurtosis feedback`
- Explain that every command takes in `-h`

- TODO install shit

Get a Postgres
==============
- Kurtosis as a "build system for distributed applications"
- Kurtosis uses Starlark for distributed application definitions
    - TODO link out
    - Stress that it's just a subset of Python
- Let's see it in action
- Create new directory
- TODO explain enclaves

Create a file called `main.star` inside it with the following contents:
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

Now, [run](TODO KURTOSIS RUN) the file you just created:

```bash
kurtosis run --enclave-identifier quickstart main.star
```

Kurtosis will create an [enclave](TODO LINK) called `quickstart` and the Starlark code will run inside.

Now if you [inspect](TODO TODO) the `quickstart` enclave...

```bash
kurtosis enclave inspect quickstart
```

...you'll see that a Postgres instance has been started:

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

Note that Kurtosis did not execute anything until _after_ interpretation and validation completed. You can think of interpretation and validation like Kurtosis' "compilation" for your distributed system: we can catch many errors before any containers run, which shortens the dev loop and reduces the resource burden on your machine.

We call this approach [multi-phase execution](TODO). While it has powerful benefits, the major gotcha for new Kurtosis users is that _you cannot reference execution-time values like IP address in Starlark_ because they simply don't exist at interpretation time. We'll see how to work around this limitation later.

Add some data
=============
A database without data is a fancy heater, so let's add some. 

Our two options for seeding a Postgres database are:

1. Making a sequence of PSQL commands via the `psql` binary
1. Using `pg_restore` to load a package of data

Both are possible in Kurtosis, but for this tutorial we'll do the second one using a seed data TAR of DVD rental information [courtesy of postgresqltutorial.com](https://www.postgresqltutorial.com/postgresql-getting-started/postgresql-sample-database/). 

Normally seeding a database would require downloading the seed data to your machine, starting Postgres, and writing a pile of Bash to copy the seed data to the Postgres server and run a `pg_restore`. If you forgot to check if the database is available, you may get flakes when you try to use the seeding logic in a test. You could try Docker Compose to volume-moun the data TAR into the Postgres server, but you'd still need to handle Postgres availability and sequencing the `pg_restore` afterwards.

Enter Kurtosis. Kurtosis Starlark scripts can use data as a first-class primitive, and sequence tasks such as `pg_restore` into the plan. Let's see it in action, and we'll explain what's happening afterwards.

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
echo 'name: "github.com/YOUR-GITHUB-USERNAME/kurtosis-quickstart"' > kurtosis.yml
```

You're almost ready to run, but you still have the `quickstart` enclave hanging around from the previous section. [Blow it away](TODO KURTOSIS CLEAN) and rerun:

```bash
kurtosis clean -a && kurtosis run --enclave-identifier quickstart .
```

(This will be your dev loop for the rest of the quickstart.)

The output should also look more interesting as our plan has grown bigger:

```text
INFO[2023-03-14T21:16:30-03:00] Destroying enclaves...
INFO[2023-03-14T21:16:30-03:00] Enclaves successfully destroyed
INFO[2023-03-14T21:16:31-03:00] Creating a new enclave for Starlark to run inside...
INFO[2023-03-14T21:16:36-03:00] Enclave 'quickstart' created successfully
INFO[2023-03-14T21:16:36-03:00] Executing Starlark package at '/Users/argos/Library/CloudStorage/GoogleDrive-thetallmonkey@gmail.com/My Drive/project-support/quickstart-new-iteration/iterations' as the passed argument '.' looks like a directory
INFO[2023-03-14T21:16:36-03:00] Compressing package 'github.com/ME/kurtosis-quickstart' at '.' for upload
INFO[2023-03-14T21:16:36-03:00] Uploading and executing package 'github.com/ME/kurtosis-quickstart'

> upload_files src="github.com/kurtosis-tech/examples/data-package/dvd-rental-data.tar"
Files with artifact name 'yearning-boulder' uploaded with artifact UUID '924402c17eb94bbd9b8f7657a9a7aba1'

> add_service service_name="postgres" config=ServiceConfig(image="postgres:15.2-alpine", ports={"postgres": PortSpec(number=5432, application_protocol="postgresql")}, files={"/seed-data": "yearning-boulder"}, env_vars={"POSTGRES_DB": "app_db", "POSTGRES_PASSWORD": "password", "POSTGRES_USER": "app_user"})
Service 'postgres' added with service UUID '06e34fe9c2374bed84b55de45d2b353c'

> wait recipe=ExecRecipe(command=["psql", "-U", "app_user", "-d", "app_db", "-c", "\\l"]) field="code" assertion="==" target_value=0 timeout="5s"
Wait took 2 tries (1.111577583s in total). Assertion passed with following:
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
INFO[2023-03-14T21:16:42-03:00] ===================================================
INFO[2023-03-14T21:16:42-03:00] ||          Created enclave: quickstart          ||
INFO[2023-03-14T21:16:42-03:00] ===================================================
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

Kurtosis' first-class data primitive is called a [files artifact](TDOO). Each files artifact is a TGZ of arbitrary files, living inside the enclave. So long as a files artifact exists, Kurtosis knows how to mount its contents on a service. We used this feature to mount the seed data into the Postgres instance via the `ServiceConfig.files` option:

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

There are many ways to create files artifacts in an enclave. The simplest is to upload files from your local machine using [the `kurtosis files upload` command](TODO TODO). A more advanced way is to upload files using [the `upload_files` Starlark instruction](https://docs.kurtosis.com/reference/starlark-instructions#upload_files) on the plan.

However, you never downloaded the seed data on your local machine. We didn't need you to, because we leveraged one of the most powerful features of Kurtosis: composability. 

Kurtosis has [a built-in packaging/dependency system](TODO TODO) that allows Starlark code to depend on other Starlark code via Github repositories. When you created the `kurtosis.yml` file, you linked your code into the packaging system: you told Kurtosis that your code is a part of a Kurtosis package, which allowed your code to consume external Starlark code.

This line at the top of your `main.star`...

```python
data_package_module = import_module("github.com/kurtosis-tech/examples/data-package/main.star")
```

...created a dependency on [the external Kurtosis package living here](https://github.com/kurtosis-tech/examples/tree/main/data-package). Your code then called that dependency code here...

```python
data_package_module_result = data_package_module.run(plan, struct())
```

...which in turn ran [the code in the `main.star` of that external Kurtosis package](https://github.com/kurtosis-tech/examples/blob/main/data-package/main.star). That package happens to contain [the seed data](https://github.com/kurtosis-tech/examples/blob/main/data-package/dvd-rental-data.tar), and it uses the `upload_data` Starlark instruction on the plan to make the seed data available via a files artifact. From there, all we needed to do was mount it on the `postgres` service.

This ability to modularize your distributed application logic using only a Github repo is one of Kurtosis' most loved features. We won't dive into all the usecases now, but [the examples repo](https://github.com/kurtosis-tech/examples) can serve as a good source of inspiration.


Add an API
==========
Databases don't come alone, however. In this section we'll add a [PostgREST API](https://postgrest.org/en/stable/) in front of the database and see how Kurtosis handles inter-service dependencies.

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
        # TODO replace with postgrest.name when we can support it!!
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

The current status of the enclave is usually a good place to start. Inspecting our enclave...

```bash
kurtosis enclave inspect quickstart
```

...reveals the following output:

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
kurtosis service logs quickstart postgres
```

...we can see that the service is dying:

```text
15/Mar/2023:01:15:30 +0000: Attempting to connect to the database...
15/Mar/2023:01:15:30 +0000: {"code":"PGRST000","details":"FATAL:  password authentication failed for user \"postgres\"\n","hint":null,"message":"Database connection error. Retrying the connection."}
15/Mar/2023:01:15:30 +0000: FATAL:  password authentication failed for user "postgres"

postgrest: thread killed
```

Referencing back to our Starlark, we can see the problem: we're creating the Postgres database with a user called `app_user`, but we're telling PostgREST to try and connect through a user called `postgres`:

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
        "postgres",   # <----- !!! The PROBLEM !!!!
        POSTGRES_PASSWORD,
        postgres.ip_address,
        postgres.ports[POSTGRES_PORT_ID].number,
        POSTGRES_DB,
    )
```

Replace that `"postgres"` with `POSTGRES_USER` to user the correct username, and then rerun your dev loop:

```bash
kurtosis clean -a && kurtosis run --enclave-identifier quickstart .
```

Now if we inspect the enclave...

```bash
kurtosis enclave inspect quickstart
```

...we can see that PostgREST is now `RUNNING` correctly:

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

Yet... PostgREST needs to know the IP address or hostname of the Postgres service, and we said earlier that Starlark (interpretation) can never know execution values. How can this be?

Answer: execution-time values are represented at interpretation time as [future references](TODO) - special Starlark strings like `{{kurtosis:6670e781977d41409f9eb2833977e9df:ip_address.runtime_value}}` that Kurtosis will replace at execution time with the actual value. In this case, the `postgres_url` variable here...

```python
postgres_url = "postgresql://{}:{}@{}:{}/{}".format(
    POSTGRES_USER,
    POSTGRES_PASSWORD,
    postgres.ip_address,
    postgres.ports[POSTGRES_PORT_ID].number,
    POSTGRES_DB,
)
```

...used the `postgres.ip_address` and `postgres.ports[POSTGRES_PORT_ID].number` future references, so that when `postgres_url` was used as an environment variable...

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

...Kurtosis simply swapped in the correct execution-time values when it came time to start the PostgREST container. While future references take some getting used to, we've found the feedback loop speedup to be very worth it.

Modifying data
==============
Now that we have an API, we should be able to interact with the data.

Inspect your enclave once more:

```bash
kurtosis enclave inspect quickstart
```

Notice how Kurtosis automatically exposed the PostgREST container's `http` port to your machine:

```text
28a923400e50   postgrest   http: 3000/tcp -> http://127.0.0.1:59992             RUNNING
```

(In this output, the `http` port is exposed as URL `http://127.0.0.1:59992` but your output will be different.)

You can paste the URL from your output into your browser to verify that you are indeed talking to the PostgREST inside your `quickstart` enclave.

Now make a request to insert a row into the database (replacing the `http://127.0.0.1:59992` portion of the URL with the correct URL from your `enclave inspect` output)...

```bash
curl -XPOST -H "content-type: application/json" http://127.0.0.1:59992/actor --data '{"first_name": "Kevin", "last_name": "Bacon"}'
```

...and then query for it (again replacing `http://127.0.0.1:59992` with your correct URL)...

```bash
curl -XGET "http://127.0.0.1:59992/actor?first_name=eq.Kevin&last_name=eq.Bacon"
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
        # TODO replace with postgrest.name when we can support it!!
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

Inspect your enclave to the get the PostgREST URL...

```bash
kurtosis enclave inspect quickstart
```

...and use the PostgREST `http` URL to query for the rows you just added (replacing `http://127.0.0.1:59992` with your URL)...

```bash
curl -XGET "http://127.0.0.1:59992/actor?or=(last_name.eq.Buscemi,last_name.eq.Bacon)"
```

...yielding:

```text
[{"actor_id":201,"first_name":"Kevin","last_name":"Bacon","last_update":"2023-03-15T02:29:53.454697"},
 {"actor_id":202,"first_name":"Steve","last_name":"Buscemi","last_update":"2023-03-15T02:29:53.454697"}]
```

### Review
How did this work?

Mechanically, [the `request` Starlark instruction](TODO) is being used create a JSON string that's getting shoved at PostgREST, which writes it to the database:

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

At a higher level, Kurtosis automatically deserialized the `[{"first_name":"Kevin", "last_name": "Bacon"}, {"first_name":"Steve", "last_name":"Buscemi"}]` string passed in to `kurtosis run` and fed it as the `args` object to the `run` function in `main.star`:

```python
def run(plan, args):
```

Publishing
==========
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

Now that your package is live, any Kurtosis user can run it without the code being checked out at all:

```bash
kurtosis clean -a && kurtosis run --enclave-identifier quickstart github.com/mieubrisse/kurtosis-quickstart
```

(Parameterization will still work, of course.)

### Review
Publishing a Kurtosis package is as simple as verifying the `name` key in `kurtosis.yml` matches and pushing it to Github. That package will then be available to every `kurtosis run`, as well as every Starlark script via the `import_module` composition flow.

<!-- TODO TODO TDOO
Testing
=======
- Use the package in some Starlark tests
- Show how we can use the `insert_data` function independently, to operate on an existing environment
- Show the `enclave dump` usecase - very useful for test logs!!!
-->

Conclusion
==========
In this tutorial you have:

- Started a Postgres database
- Seeded it by importing a third-party Starlark package
- Added an API server
- Inserted & queried data via the API
- Parameterized data insertion
- Published your package

Along the way you've learned about several Kurtosis concepts:

- Starlark
- Enclaves
- Multi-phase execution
- Files artifacts
- Kurtosis packages
- Future references

- Environments as a first-class concept - easy to create, access, and destroy
- Two ways of manipulating the contents of an environment, [through the CLI][cli-reference] and [through Starlark][starlark-instructions-reference]
- Referencing external resources in Starlark
- Publishing & consuming environment definitions through the concept of [Kurtosis packages][packages-reference]
- Parameterizing environment definitions through the concept of [runnable package][runnable-packages-reference]

These are just the basics of Kurtosis. To dive deeper, you can now:

- Learn more about [the architecture of Kurtosis][architecture-explanation]
- Explore [the catalog of Starlark instructions][starlark-instructions-reference]
- Explore [Kurtosis-provided packages being used in production][kurtosis-managed-packages]
- [Search GitHub for Kurtosis packages in the wild][wild-kurtosis-packages]

:::info
Get a personalized onboarding session with us [here](https://calendly.com/d/zgt-f2c-66p/kurtosis-onboarding).
:::

