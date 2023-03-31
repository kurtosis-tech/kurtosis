---
title: Quickstart
sidebar_label: Quickstart
slug: /quickstart
---

Introduction
------------

<details><summary>###TL;DR Version</summary>
This quickstart is in a "code along" format. If you'd prefer to dive straight into the code and inner workings of the distributed application we build with Kurtosis here, get [set up](#setup) and then run:

:::bash
kurtosis run --enclave quickstart github.com/kurtosis-tech/awesome-kurtosis/quickstart
:::

Then, you can refer to the [recap](#conclusion) and the other `Review` sections in this quickstart guide to learn more about what we did and how we did it.

</details>

Welcome to the [Kurtosis][homepage] quickstart! This guide will take ~15 minutes and will walk you through building a basic Kurtosis package. This guide is in a "code along" format, meaning we assume the user will be following the code examples and running Kurtosis CLI commands on your local machine. Everything you will run in this guide is free, public, and does not contain any sensitive data.

For a quick read on what Kurtosis is and what problems Kurtosis aims to solve, our [introduction page][homepage] will be a great starting point, alongside our [motivations behind starting Kurtosis][why-we-built-kurtosis-explanation].

:::tip What You'll Do

- Start a containerized Postgres database in Kurtosis
- Seed your database with test data using task sequencing
- Connect an API server to your database using dynamic service dependencies
- Parameterize your application setup in order to automate loading data into your API
:::

<details><summary>Getting help and giving feedback</summary>

There are many ways to get help and give feedback. First, every Kurtosis command accepts a `-h` flag to print helptext. If that doesn't help, here are some ways to get support:

- `kurtosis feedback "my feedback"` will take you to our [Github issue creation page](https://github.com/kurtosis-tech/kurtosis/issues/new/choose) to file an issue with pre-filled text `my feedback`.  Passing in `--bug` or `--docs` will take you to the specific Issue template for bug reports or docs issues, respectively
- `kurtosis feedback --calendly` opens a calendly link for a personal help session with [our cofounder Kevin][kevin-linked].
- `kurtosis discord` command will open up our [Discord](https://discord.com/channels/783719264308953108/783719264308953111), where you can get live support via chat with our team.

**Don't suffer in silence - we want to help and hear from you!**

</details>

Setup
-----

#### Requirements
Before you proceed, please make sure you have:
- [Installed and started the Docker engine][installing-docker-guide]
- [Installed the Kurtosis CLI][installing-kurtosis-guide] (or [upgraded to latest][upgrading-kurtosis-guide] if you already have it)

Hello, World
------------
First, create and `cd` into a directory to hold the project you'll be working on:

```bash
mkdir kurtosis-quickstart && cd kurtosis-quickstart
```

Next, create a Starlark file called `main.star` inside your new directory with the following contents (more on Starlark in the "Review" section coming up soon):

```python
def run(plan, args):
    plan.print("Hello, world")
```

:::tip
If you're using Vim, you can add the following to your `.vimrc` to get Starlark syntax highlighting:

```
" Add syntax highlighting for Starlark files
autocmd FileType *.star setlocal filetype=python
```

:::

Finally, [run][kurtosis-run-reference] the script (we'll explain enclaves in the "Review" section too):

```bash
kurtosis run --enclave quickstart main.star
```

Kurtosis will work for a bit, and then deliver you the following result:

```text
INFO[2023-03-15T04:27:01-03:00] Creating a new enclave for Starlark to run inside...
INFO[2023-03-15T04:27:05-03:00] Enclave 'quickstart' created successfully

> print msg="Hello, world"
Hello, world

Starlark code successfully run. No output was returned.
INFO[2023-03-15T04:27:05-03:00] ===================================================
INFO[2023-03-15T04:27:05-03:00] ||          Created enclave: quickstart          ||
INFO[2023-03-15T04:27:05-03:00] ===================================================
Name:                                 quickstart
UUID:                                 a78f2ce1ca68
Status:                               RUNNING
Creation Time:                        Wed, 15 Mar 2023 04:27:01 -03

========================================= Files Artifacts =========================================
UUID   Name

========================================== User Services ==========================================
UUID   Name   Ports   Status
```

Congratulations - you've written your first Kurtosis code!

### Review: Hello, World
:::info
You'll use these "Review" sections to explain what happened in the section. 
:::

In this section, you created a `main.star` file that simply told Kurtosis to print `Hello, world`. The `.star` extension corresponds to [Starlark][starlark-reference], a Python dialect also used by Google and Meta for configuring build systems.

When you ran `main.star`, you got `Created enclave: quickstart`. An [enclave][enclaves-reference] is a Kurtosis primitive that can be thought of as an *ephemeral test environment*, on top of Docker or Kubernetes, for a distributed application. The distributed applications that you define with Starlark will run inside enclaves. If you'd like, you can tear down your enclave and any of their artifacts by running: `kurtosis clean -a` (more on the `kurtosis clean` command [here][kurtosis-clean-reference]).

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
        name = "postgres",
        config = ServiceConfig(
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

Before you run the above command, remember that you still have the `quickstart` enclave hanging around from the previous section. To [clean up the previous enclave][kurtosis-clean-reference] and execute our new `main.star` file above, run:

```bash
kurtosis clean -a && kurtosis run --enclave quickstart main.star
```

:::info
The `--enclave` flag is used to specify the enclave to use for that particular run. If one doesn't exist, Kurtosis will create an enclave with that name - which is what is happening here. Read more about `kurtosis run` [here][kurtosis-run-reference]. 

This entire  "clean-and-run" process will be your dev loop for the rest of the quickstart as you add more services and operations to our distributed application.
:::

You'll see in the result that the `quickstart` enclave now contains a Postgres instance:

```text
Name:                                 quickstart
UUID:                                 a30106a0bb87
Status:                               RUNNING
Creation Time:                        Tue, 14 Mar 2023 20:23:54 -03

========================================= Files Artifacts =========================================
UUID   Name

========================================== User Services ==========================================
UUID           Name       Ports                                                Status
b6fc024deefe   postgres   postgres: 5432/tcp -> postgresql://127.0.0.1:59299   RUNNING
```

### Review: Run Postgres
So what actually happened? Three things actually:

1. **Interpretation:** Kurtosis first ran your Starlark to build [a plan](https://docs.kurtosis.com/reference/plan) for what you wanted done (in this case, starting a Postgres instance)
1. **Validation:** Kurtosis then ran several validations against your plan, including validating that the Postgres image exists
1. **Execution:** Kurtosis finally executed the validated plan inside the enclave to start a Postgres container

Note that Kurtosis did not execute anything until _after_ Interpretation and Validation completed. You can think of Interpretation and Validation like Kurtosis' "compilation" step for your distributed application: you can catch many errors before any containers run, which shortens the dev loop and reduces the resource burden on your machine.

We call this approach [multi-phase runs][multi-phase-runs-reference]. While multi-phase runs has powerful benefits over traditional scripting, it also means _you cannot reference Execution values like IP address in Starlark_ because they simply don't exist at Interpretation time. We'll explore how Kurtosis gracefully handles values generated during the Execution phase at the Interpretation phase later on in the quickstart.

**This section introduced Kurtosis' ability to validate that definitions work as intended, _before_ they are run - helping developers catch errors sooner & save resources when configuring multi-container test environments.**

Add some data
-------------
A database without data is a fancy heater, so let's add some. 

Your two options for seeding a Postgres database are:

1. Making a sequence of PSQL commands via the `psql` binary
1. Using `pg_restore` to load a package of data

Both are possible in Kurtosis, but for this tutorial we'll use `pg_restore` to seed your database with a TAR of DVD rental information, [courtesy of postgresqltutorial.com](https://www.postgresqltutorial.com/postgresql-getting-started/postgresql-sample-database/). 

#### Without Kurtosis
Normally going this route (using `pg_restore`) requires downloading the seed data to your local machine, starting Postgres, writing a pile of Bash to copy the seed data to the Postgres server, and then finally running the `pg_restore` command. If you forget to check if the database is available, you may get flakes when you try to use the seeding logic in a test. 

Alternatively, you could use Docker Compose to volume-mount the data TAR into the Postgres server, but you'd still need to handle Postgres availability and sequencing the `pg_restore` afterwards.

#### With Kurtosis
By contrast, Kurtosis Starlark scripts can use data as a first-class primitive and sequence tasks such as `pg_restore` into the plan. 

Let's see it in action, and we'll explain what's happening afterwards.

First, in your working directory (`kurtosis-quickstart`), next to your `main.star` file, create a file called `kurtosis.yml` with the following contents:

```bash
name: "github.com/john-snow/kurtosis-quickstart"
```

Then update your `main.star` with the following:

```python
data_package_module = import_module("github.com/kurtosis-tech/awesome-kurtosis/data-package/main.star")

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
        name = "postgres",
        config = ServiceConfig(
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

Now, run the following to see what happens:

```bash
kurtosis clean -a && kurtosis run --enclave quickstart .
```

(Notice you are using `.` instead of `main.star`)

The output should also look more interesting as your plan has grown bigger:

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
INFO[2023-03-15T04:34:10-03:00] Compressing package 'github.com/john-snow/kurtosis-quickstart' at '.' for upload
INFO[2023-03-15T04:34:10-03:00] Uploading and executing package 'github.com/john-snow/kurtosis-quickstart'

> upload_files src="github.com/kurtosis-tech/awesome-kurtosis/data-package/dvd-rental-data.tar"
Files with artifact name 'howling-thunder' uploaded with artifact UUID '32810fc8c131414882c52b044318b2fd'

> add_service name="postgres" config=ServiceConfig(image="postgres:15.2-alpine", ports={"postgres": PortSpec(number=5432, application_protocol="postgresql")}, files={"/seed-data": "howling-thunder"}, env_vars={"POSTGRES_DB": "app_db", "POSTGRES_PASSWORD": "password", "POSTGRES_USER": "app_user"})
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
Name:                                 quickstart
UUID:                                 995fe0ca69fe
Status:                               RUNNING
Creation Time:                        Wed, 15 Mar 2023 04:34:06 -03

========================================= Files Artifacts =========================================
UUID           Name
32810fc8c131   howling-thunder

========================================== User Services ==========================================
UUID           Name       Ports                                                Status
f1d9cab2ca34   postgres   postgres: 5432/tcp -> postgresql://127.0.0.1:62914   RUNNING
```

Does your Postgres have data now? Let's find out by opening a shell on the Postgres container and logging into the database:

```bash
kurtosis service shell quickstart postgres
```

From there, listing the tables in the Postgres can be done with:

``` bash
psql -U app_user -d app_db -c '\dt'
```

...which will reveal that many new tables now exist:

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

### Review: Add some data
So what just happened?

#### We created a Kurtosis package

By creating a [`kurtosis.yml`][kurtosis-yml-reference] file in your working directory, you turned your working directory into a [Kurtosis package][packages-reference] (specifically, a [runnable package][runnable-packages-reference]). After you did this, your newly created Kurtosis package could now declare dependencies on external packages using [Kurtosis’ built-in packaging/dependency system][how-do-imports-work-explanation].

To see this in action, the first line in your local `main.star` file was used to import, and therefore declare a dependency on, an external package called `data-package` using a [locator][locators-reference]:

```python
data_package_module = import_module("github.com/kurtosis-tech/awesome-kurtosis/data-package/main.star")
```
... which we then ran locally:
```python
data_package_module_result = data_package_module.run(plan, struct())
```

This external Kurtosis package, named ["data-package"][data-package-example] contains the seed data for your Postgres instance that we [referenced earlier](#add-some-data) as a `.tar` file.

:::info
Special note here that we used a locator to import an external package from your [`awesome-kurtosis` repository][awesome-kurtosis-repo] and _not_ a regular URL. Learn more about how they differ [here][locators-reference].
:::

#### You imported seed data into your Kurtosis package
The [`main.star` file][data-package-example-main.star] in that external "data-package" contained Starlark instructions to store the `.tar` data as a [files artifact][files-artifacts-reference] using the [`files_upload` Starlark instruction][kurtosis-files-upload-reference]:

```python
TAR_FILENAME = "dvd-rental-data.tar"
def run(plan, args):
    dvd_rental_data = plan.upload_files("github.com/kurtosis-tech/awesome-kurtosis/data-package/" + TAR_FILENAME)

    result =  struct(
        files_artifact = dvd_rental_data, # Needed to mount the data on a service
        tar_filename = TAR_FILENAME,      # Useful to reference the data TAR contained in the files artifact
    )

    return result
```

A [files artifact][files-artifacts-reference] is Kurtosis' first-class data primitive and is a TGZ of arbitrary files living inside an enclave. So long as a files artifact exists, Kurtosis knows how to mount its contents on a service.  

#### You mounted and seeded the data into your Postgres instance
Next, you mounted the seed data, stored in your enclave now as a files artifact, into your Postgres instance using the `ServiceConfig.files` option:

```python
postgres = plan.add_service(
    name = "postgres",
    config = ServiceConfig(
        # ...omitted...
        files = {
            SEED_DATA_DIRPATH: data_package_module_result.files_artifact,
        }
    ),
)
```

Then to seed the data, you used the [`exec` Starlark instruction][exec-reference]:
```python
plan.exec(
    service_name = "postgres",
    recipe = ExecRecipe(command = ["pg_restore"] + postgres_flags + [
        "--no-owner",
        "--role=" + POSTGRES_USER,
        SEED_DATA_DIRPATH + "/" + data_package_module_result.tar_filename,
        ]
    ),
```
**Here, you saw one of Kurtosis' most loved features: the ability to modularize and share your distributed application logic using only a Github repository.** We won't dive into all the usecases now, but [the examples here][awesome-kurtosis-repo] can serve as a good source of inspiration.

Add an API
----------
Databases don't come alone, however. In this section we'll add a [PostgREST API][postgrest] in front of the database and see how Kurtosis handles inter-service dependencies.

Replace the contents of your `main.star` with this:

```python
data_package_module = import_module("github.com/kurtosis-tech/awesome-kurtosis/data-package/main.star")

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
        name = "postgres",
        config = ServiceConfig(
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
    api = plan.add_service(
        name = "api", # Naming our PostgREST service "api"
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
        service_name = "api",
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

Now, run the same dev loop command as before (and don't worry about the result, we'll explain that later):

```bash
kurtosis clean -a && kurtosis run --enclave quickstart .
```

We just got a failure, just like we might when building a real system!

```text
> wait recipe=GetHttpRequestRecipe(port_id="http", endpoint="/actor", extract="") field="code" assertion="==" target_value=200 timeout="5s"
There was an error executing Starlark code
An error occurred executing instruction (number 6) at github.com/ME/kurtosis-quickstart[77:14]:
wait(recipe=GetHttpRequestRecipe(port_id="http", endpoint="/actor", extract=""), field="code", assertion="==", target_value=200, timeout="5s", service_name="api")
Caused by: Wait timed-out waiting for the assertion to become valid. Waited for '8.183602629s'. Last assertion error was: 
<nil>

Error encountered running Starlark code.
```

Here, Kurtosis is telling us that the `wait` instruction on line `77` of your `main.star` (the one for ensuring PostgREST is up) is timing out.

:::info
Fun fact: this failure was encountered at the last step in Kurtosis' [multi-phase run approach][multi-phase-runs-reference], which is also called the Execution step that we mentioned earlier [when we got Postgres up and running](#review-run-postgres).
:::


#### Investigating the issue
The enclave state is usually a good place to start. If you look at the bottom of your output you'll see the following state of the enclave:

```text

Name:                                 quickstart
UUID:                                 5b360f940bcc
Status:                               RUNNING
Creation Time:                        Tue, 14 Mar 2023 22:15:19 -03

========================================= Files Artifacts =========================================
UUID           Name
323c9a71ebbf   crimson-haze

========================================== User Services ==========================================
UUID           Name        Ports                                                Status
45b355fc810b   postgres    postgres: 5432/tcp -> postgresql://127.0.0.1:59821   RUNNING
80987420176f   api         http: 3000/tcp                                       STOPPED
```

From the above, the problem is clear now: the PostgREST service (named: `api`) status is `STOPPED`, rather than `RUNNING`. 

When we grab the PostgREST logs...

```bash
kurtosis service logs quickstart api
```

...we can see that the PostgREST is dying:

```text
15/Mar/2023:01:15:30 +0000: Attempting to connect to the database...
15/Mar/2023:01:15:30 +0000: {"code":"PGRST000","details":"FATAL:  password authentication failed for user \"postgres\"\n","hint":null,"message":"Database connection error. Retrying the connection."}
15/Mar/2023:01:15:30 +0000: FATAL:  password authentication failed for user "postgres"

postgrest: thread killed
```

Looking back to your Starlark code, you can see the problem: it's creating the Postgres database with a user called `app_user`, but it's telling PostgREST to try and connect through a user called `postgres`:

```python
POSTGRES_USER = "app_user"

# ...

def run(plan, args):
    # ...

    # Add a Postgres server
    postgres = plan.add_service(
        name = "postgres",
        config = ServiceConfig(
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
        "postgres",   # <---------- THE PROBLEM IS HERE
        POSTGRES_PASSWORD,
        postgres.ip_address,
        postgres.ports[POSTGRES_PORT_ID].number,
        POSTGRES_DB,
    )
```

In the line declaring the `postgres_url` variable in your `main.star` file, replace the `"postgres",` string with `POSTGRES_USER,` to use the correct username we specified at the beginning of our file. Then re-run your dev loop:

```bash
kurtosis clean -a && kurtosis run --enclave quickstart .
```

Now at the bottom of the output we can see that the PostgREST service is `RUNNING` correctly:

```text
Name:                         quickstart
UUID:                         11c0ac047299
Status:                       RUNNING
Creation Time:                Tue, 14 Mar 2023 22:30:02 -03

========================================= Files Artifacts =========================================
UUID           Name
323c9a71ebbf   crimson-haze

========================================== User Services ==========================================
UUID           Name        Ports                                                Status
ce90b471a982   postgres    postgres: 5432/tcp -> postgresql://127.0.0.1:59883   RUNNING
98094b33cd9a   api         http: 3000/tcp -> http://127.0.0.1:59887             RUNNING
```

### Review: Add an API
In this section, you spun up a new PostgREST service (that we named `api` for readability) with a dependency on the Postgres service. Normally, PostgREST needs to know the IP address or hostname of the Postgres service, and we said earlier that Starlark (the Interpretation phase) can never know Execution values. 

So how did the services get connected?

Answer: Execution-time values are represented at Interpretation time as [future references][future-references-reference] that Kurtosis will replace at Execution time with the actual value. In this case, the `postgres_url` variable here...

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
api = plan.add_service(
    name = "api", # Naming our PostgREST service "api"
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

**What you've just seen is Kurtosis' powerful ability to data generated at runtime to set up service dependencies in multi-container test environments. You also saw how seamless it was to to run on-box CLI commands on a container.**

Modifying data
--------------
Now that you have an API, you should be able to interact with the data.

[Inspect][kurtosis-enclave-inspect-reference] your enclave:

```bash
kurtosis enclave inspect quickstart
```

Notice how Kurtosis automatically exposed the PostgREST container's `http` port to your machine:

```text
28a923400e50   api         http: 3000/tcp -> http://127.0.0.1:59992             RUNNING
```

:::info
In this output the `http` port is exposed as URL `http://127.0.0.1:59992`, but your port number will be different.
:::

You can paste the URL from your output into your browser (or Cmd+click if you're using [iTerm][iterm]) to verify that you are indeed talking to the PostgREST inside your `quickstart` enclave:

```json
{"swagger":"2.0","info":{"description":"","title":"standard public schema","version":"10.2.0.20230209 (pre-release) (a1e2fe3)"},"host":"0.0.0.0:3000","basePath":"/","schemes":["http"],"consumes":["application/json","application/vnd.pgrst.object+json","text/csv"],"produces":["application/json","application/vnd.pgrst.object+json","text/csv"],"paths":{"/":{"get":{"tags":["Introspection"],"summary":"OpenAPI description (this document)","produces":["application/openapi+json","application/json"],"responses":{"200":{"description":"OK"}}}},"/actor":{"get":{"tags":["actor"],"parameters":[{"$ref":"#/parameters/rowFilter.actor.actor_id"},{"$ref":"#/parameters/rowFilter.actor.first_name"},{"$ref":"#/parameters/rowFilter.actor.last_name"},{"$ref":"#/parameters/rowFilter.actor.last_update"},{"$ref":"#/parameters/select"},{"$ref":"#/parameters/order"},{"$ref":"#/parameters/range"},{"$ref":"#/parameters/rangeUnit"},{"$ref":"#/parameters/offset"},{"$ref":"#/parameters/limit"},{"$ref":"#/parameters/preferCount"}], ...
```

Now make a request to insert a row into the database (replacing `$YOUR_PORT` with the `http` port from your `enclave inspect` output for the PostgREST service that we named `api`)...

```bash
curl -XPOST -H "content-type: application/json" http://127.0.0.1:$YOUR_PORT/actor --data '{"first_name": "Kevin", "last_name": "Bacon"}'
```

...and then query for it (again replacing `$YOUR_PORT` with your port)...

```bash
curl -XGET "http://127.0.0.1:$YOUR_PORT/actor?first_name=eq.Kevin&last_name=eq.Bacon"
```

...to get it back:

```text
[{"actor_id":201,"first_name":"Kevin","last_name":"Bacon","last_update":"2023-03-15T02:08:14.315732"}]
```

Of course, it'd be much nicer to formalize this in Kurtosis. Replace your `main.star` with the following:

```python
data_package_module = import_module("github.com/kurtosis-tech/awesome-kurtosis/data-package/main.star")

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
        name = "postgres",
        config = ServiceConfig(
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
    api = plan.add_service(
        name = "api",
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
        service_name = "api",
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
    if "actors" in args:
        insert_data(plan, args["actors"])

def insert_data(plan, data):
    plan.request(
        service_name = "api",
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
kurtosis clean -a && kurtosis run --enclave quickstart . '{"actors": [{"first_name":"Kevin", "last_name": "Bacon"}, {"first_name":"Steve", "last_name":"Buscemi"}]}'
```

Using the new `http` URL on the `api` service in the output, query for the rows you just added (replacing `$YOUR_PORT` with your correct PostgREST `http` port number)...

```bash
curl -XGET "http://127.0.0.1:$YOUR_PORT/actor?or=(last_name.eq.Buscemi,last_name.eq.Bacon)"
```

...to yield:

```text
[{"actor_id":201,"first_name":"Kevin","last_name":"Bacon","last_update":"2023-03-15T02:29:53.454697"},
 {"actor_id":202,"first_name":"Steve","last_name":"Buscemi","last_update":"2023-03-15T02:29:53.454697"}]
```

### Review
How did this work?

Mechanically, we first created a JSON string of data using Starlark's `json.encode` builtin. Then we used [the `request` Starlark instruction][request-reference] to shove the string at PostgREST, which writes it to the database:

```python
plan.request(
    service_name = "api",
    recipe = PostHttpRequestRecipe(
        port_id = POSTGREST_PORT_ID,
        endpoint = "/actor",
        content_type = "application/json",
        body = json.encode(data),
    )
)
```

At a higher level, Kurtosis automatically deserialized the `{"actors": [{"first_name":"Kevin", "last_name": "Bacon"}, {"first_name":"Steve", "last_name":"Buscemi"}]}` string passed as a parameter to `kurtosis run`, and put the deserialized object in the `args` parameter to the `run` function in `main.star`:

```python
def run(plan, args):
```

**This section showed how to interact with your test environment, and also how to parametrize it for others to easily modify and re-use.**


<!-- 
// NOTE(ktoday): We commented this out because the Git publishing aspect was giving people trouble (needed to have Git set up, 'git init' was hard, etc.
// I still think that publishing/shareability is a large part of Kurtosis, but there's probably a better way to highlight this

Publishing
----------
Congratulations - you've written your very first distributed application in Kurtosis! Now it's time to share it with the world.

The Kurtosis packaging system uses Github as its package repository, just like Go modules. Also like Go modules, Kurtosis packages need their name to match their location on Github.

Update the `name` key of the `kurtosis.yml` file to replace `john-snow` with your Github username:

```yaml
# You'll need to update this
name: "github.com/john-snow/kurtosis-quickstart"
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
kurtosis clean -a && kurtosis run --enclave quickstart github.com/YOUR-GITHUB-USERNAME/kurtosis-quickstart
```

(Parameterization will still work, of course.)

### Review
Publishing a Kurtosis package is as simple as verifying the `name` key in `kurtosis.yml` matches your repo and pushing it to Github. That package will then be available to every `kurtosis run`, as well as every Starlark script via the `import_module` composition flow.
-->

<!-- TODO TODO TDOO
Testing
=======
- Use the package in some Starlark tests
- Show how we can use the `insert_data` function independently, to operate on an existing environment
- Show the `enclave dump` usecase - very useful for test logs!!!
-->

Conclusion
----------
And that's it - you've written your very first distributed application in Kurtosis!

Let's review. In this tutorial you have:

- Started a Postgres database in an ephemeral, isolated test environment
- Seeded your database by importing an external Starlark package from the internet
- Set up an API server for your database and gracefully handled dynamically generated dependency data
- Inserted & queried data via the API
- Parameterized data insertion for future use

This was still just an introduction to Kurtosis. To dig deeper, visit other sections of our docs where you can read about [what Kurtosis is][homepage], understand the [architecture][architecture-explanation], and hear our [inspiration for starting Kurtosis][why-we-built-kurtosis-explanation]. 

To learn more about how Kurtosis is used, we encourage you to check out our [`awesome-kurtosis` repository][awesome-kurtosis-repo], where you will find real-world examples of Kurtosis in action, including:
- The [Ethereum package][ethereum-package], used by the Ethereum Foundation, which can be used to set up local testnets 
- A parameterized package for standing up an [n-node Cassandra cluster with Grafana and Prometheus][cassandra-package-example] out-of-the-box
- The [NEAR package][near-package] for local dApp development in the NEAR ecosystem

Finally, we'd love to hear from you. Please don't hesitate to share with us what went well, and what didn't, using `kurtosis feedback` to file an issue in our [Github](https://github.com/kurtosis-tech/kurtosis/issues/new/choose), to [email us](mailto:feedback@kurtosistech.com), or to [chat with our cofounder, Kevin](https://calendly.com/d/zgt-f2c-66p/kurtosis-onboarding).

Lastly, feel free to [star us on Github](https://github.com/kurtosis-tech/kurtosis), [join the community in our Discord](https://discord.com/channels/783719264308953108/783719264308953111), and [follow us on Twitter](https://twitter.com/KurtosisTech)!

Thank you for trying our quickstart. We hope you enjoyed it. 

<!-- !!!!!!!!!!!!!!!!!!!!!!!!!!! ONLY LINKS BELOW HERE !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! -->

<!--------------------------- Guides ------------------------------------>
[installing-kurtosis-guide]: ./guides/installing-the-cli.md#ii-install-the-cli
[installing-docker-guide]: ./guides/installing-the-cli.md#i-install--start-docker
[upgrading-kurtosis-guide]: ./guides/upgrading-the-cli.md

<!--------------------------- Explanations ------------------------------------>
[architecture-explanation]: ./explanations/architecture.md
[enclaves-reference]: ./reference/enclaves.md
[services-explanation]: ./explanations/architecture.md#services
[reusable-environment-definitions-explanation]: ./explanations/reusable-environment-definitions.md
[why-we-built-kurtosis-explanation]: ./explanations/why-we-built-kurtosis.md
[how-do-imports-work-explanation]: ./explanations/how-do-kurtosis-imports-work.md
[why-multi-phase-runs-explanation]: ./explanations/why-multi-phase-runs.md

<!--------------------------- Reference ------------------------------------>
<!-- CLI Commands Reference -->
[cli-reference]: ./reference/cli/cli.md
[kurtosis-run-reference]: ./reference/cli/run-starlark.md
[kurtosis-clean-reference]: ./reference/cli/clean.md
[kurtosis-enclave-inspect-reference]: ./reference/cli/enclave-inspect.md
[kurtosis-files-upload-reference]: ./reference/cli/files-upload.md
[kurtosis-feedback-reference]: ./reference/cli/feedback.md
[kurtosis-twitter]: ./reference/cli/twitter.md
[starlark-reference]: ./reference/starlark.md

<!-- SL Instructions Reference-->
[starlark-instructions-reference]: ./reference/starlark-instructions.md
[upload-files-reference]: ./reference/starlark-instructions.md#upload_files
[request-reference]: ./reference/starlark-instructions.md#request
[exec-reference]: ./reference/starlark-instructions.md#exec

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
[awesome-kurtosis-repo]: https://github.com/kurtosis-tech/awesome-kurtosis
[data-package-example]: https://github.com/kurtosis-tech/awesome-kurtosis/tree/main/data-package
[data-package-example-main.star]: https://github.com/kurtosis-tech/awesome-kurtosis/blob/main/data-package/main.star
[data-package-example-seed-tar]: https://github.com/kurtosis-tech/awesome-kurtosis/blob/main/data-package/dvd-rental-data.tar
[cassandra-package-example]: https://github.com/kurtosis-tech/cassandra-package

<!-- Misc -->
[homepage]: home.md
[kevin-linked]: https://www.linkedin.com/in/kevintoday/
[kurtosis-managed-packages]: https://github.com/kurtosis-tech?q=in%3Aname+package&type=all&language=&sort=
[wild-kurtosis-packages]: https://github.com/search?q=filename%3Akurtosis.yml&type=code
[bazel-github]: https://github.com/bazelbuild/bazel/
[starlark-github-repo]: https://github.com/bazelbuild/starlark
[postgrest]: https://postgrest.org/en/stable/
[ethereum-package]: https://github.com/kurtosis-tech/eth2-package
[waku-package]: https://github.com/logos-co/wakurtosis
[near-package]: https://github.com/kurtosis-tech/near-package
[iterm]: https://iterm2.com/
