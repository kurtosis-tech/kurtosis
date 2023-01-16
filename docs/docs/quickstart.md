---
title: Kurtosis Quickstart
sidebar_label: Quickstart
slug: /quickstart
---

These instructions will give you a brief quickstart of Kurtosis. They should take 15 minutes.

:::tip
All code blocks can be copied by hovering over the code block and clicking the clipboard icon that appears on the right.
:::

Set Up Prerequisites
------------------------------
### Install Docker
Verify that you have the Docker daemon installed and running on your local machine:

```
docker image ls
```

- If you don't have Docker installed, do so by following [the installation instructions](https://docs.docker.com/get-docker/)
- If Docker is installed but not running, start it

:::caution
[DockerHub restricts downloads from users who aren't logged in](https://www.docker.com/blog/what-you-need-to-know-about-upcoming-docker-hub-rate-limiting/) to 100 images downloaded per 6 hours, so if at any point in this tutorial you see the following error message:

```
Error response from daemon: toomanyrequests: You have reached your pull rate limit. You may increase the limit by authenticating and upgrading: https://www.docker.com/increase-rate-limit
```

you can fix it by creating a DockerHub account (if you don't have one already) and registering it with your local Docker engine like so:

```
docker login
```
:::

### Install the Kurtosis CLI
Follow the steps [on this installation page](./guides/installing-the-cli.md) to install the CLI, or upgrade it to latest if it's already installed.

:::tip
We strongly recommend [installing tab completion][installing-tab-complete]; you'll find it very useful!
:::

Create An Enclave
---------------------------
Kurtosis [enclaves][enclaves-explanation] are where your environments live; you can think of them as "environment containers". Here we'll create a fresh enclave.

Run the following:

```bash
kurtosis enclave add
```

:::info
This may take a few seconds as Kurtosis downloads its Docker images for the first time; subsequent runs will be much faster.
:::

:::tip
Kurtosis subcommands (e.g. `enclave` and `add` above) can be tab-completed as well!
:::

You'll see an output similar to the following:

```
INFO[2022-11-30T15:39:52-03:00] Creating new enclave...
INFO[2022-11-30T15:40:02-03:00] =======================================================
INFO[2022-11-30T15:40:02-03:00] ||          Created enclave: wandering-frog          ||
INFO[2022-11-30T15:40:02-03:00] =======================================================
```

Now, type the following but don't press ENTER yet:

```bash
kurtosis enclave inspect
```

If [you have tab completion installed][installing-tab-complete], you can now press TAB to tab-complete your enclave's ID (which will be different than `wandering-frog`).

If you don't have tab completion installed, you'll need to paste the enclave ID from the `Created enclave:` line outputted above (yours will be different than `wandering-frog`).

:::tip
[All enclave ID arguments][cli-reference] can be tab-completed.
:::

Press ENTER, and you should receive an output like so:

```
Enclave ID:                           wandering-frog
Enclave Status:                       RUNNING
Creation Time:                        Wed, 30 Nov 2022 15:39:52 -03
API Container Status:                 RUNNING
API Container Host GRPC Port:         127.0.0.1:63593
API Container Host GRPC Proxy Port:   127.0.0.1:63594

========================================== User Services ==========================================
GUID   ID   Ports   Status
```

`kurtosis enclave inspect` is the way to investigate an enclave.

If you ever forget your enclave ID or don't feel like using tab completion, you can always run the following:

```bash
kurtosis enclave ls
```

This will print all the enclaves inside your Kurtosis cluster:

```
EnclaveID        Status    Creation Time
wandering-frog   RUNNING   Wed, 30 Nov 2022 15:39:52 -03
```

Now run the following to store your enclave ID in a variable, replacing `YOUR_ENCLAVE_ID_HERE` with your enclave's ID.

```bash
ENCLAVE_ID="YOUR_ENCLAVE_ID_HERE"
```

We'll use this variable so that you can continue to copy-and-paste code blocks in the next section.

Start A Service
---------------------------
Distributed applications are composed of [services][services-explanation]. Here we'll start a simple service, and see some of the options Kurtosis has for debugging.

Enter this command:

```bash
kurtosis service add "$ENCLAVE_ID" my-nginx nginx:latest --ports http=80
```

You should see output similar to the following:

```
Service ID: my-nginx
Ports Bindings:
   http:   80/tcp -> 127.0.0.1:63614
```

Now inspect your enclave again:

```bash
kurtosis enclave inspect "$ENCLAVE_ID"
```

You should see a new service with the service ID `my-nginx` in your enclave:

```
Enclave ID:                           wandering-frog
Enclave Status:                       RUNNING
Creation Time:                        Wed, 30 Nov 2022 15:39:52 -03
API Container Status:                 RUNNING
API Container Host GRPC Port:         127.0.0.1:63593
API Container Host GRPC Proxy Port:   127.0.0.1:63594

========================================== User Services ==========================================
GUID                  ID         Ports                             Status
my-nginx-1669833787   my-nginx   http: 80/tcp -> 127.0.0.1:63614   RUNNING
```

Kurtosis binds all service ports to ephemeral ports on your local machine. Copy the `127.0.0.1:XXXXX` address into your browser (yours will be different), and you should see a welcome message from your NginX service running inside the enclave you created.

Now paste the following but don't press ENTER yet:

```bash
kurtosis service shell "$ENCLAVE_ID"
```

If you have tab completion installed, press TAB. The service GUID of the NginX service will be completed (which in this case was `my-nginx-1669833787`, but yours will be different).

If you don't have tab completion installed, paste in the service GUID of the NginX service from the `enclave inspect` output above (which was `my-nginx-1669833787`, but yours will be different).

:::tip
Like enclave IDs, [all service GUID arguments][cli-reference] can be tab-completed.
:::

Press ENTER, and you'll be logged in to a shell on the container:

```
Found bash on container; creating bash shell...
root@f046da8acc11:/#
```

Kurtosis will try to give you a `bash` shell, but will drop down to `sh` if `bash` doesn't exist on the container.

Feel free to explore, and enter `exit` or press Ctrl-D when you're done.

Now enter the following but don't press ENTER:

```bash
kurtosis service logs -f "$ENCLAVE_ID"
```

Once again, you can use tab completion to fill the service GUID if you have it enabled. If not, you'll need to copy-paste the service GUID as the last argument.

Press ENTER, and you'll see a live-updating stream of the service's logs:

```
/docker-entrypoint.sh: /docker-entrypoint.d/ is not empty, will attempt to perform configuration
/docker-entrypoint.sh: Looking for shell scripts in /docker-entrypoint.d/
/docker-entrypoint.sh: Launching /docker-entrypoint.d/10-listen-on-ipv6-by-default.sh
10-listen-on-ipv6-by-default.sh: info: Getting the checksum of /etc/nginx/conf.d/default.conf
10-listen-on-ipv6-by-default.sh: info: Enabled listen on IPv6 in /etc/nginx/conf.d/default.conf
/docker-entrypoint.sh: Launching /docker-entrypoint.d/20-envsubst-on-templates.sh
/docker-entrypoint.sh: Launching /docker-entrypoint.d/30-tune-worker-processes.sh
/docker-entrypoint.sh: Configuration complete; ready for start up
2022/11/30 18:43:09 [notice] 1#1: using the "epoll" event method
2022/11/30 18:43:09 [notice] 1#1: nginx/1.23.2
2022/11/30 18:43:09 [notice] 1#1: built by gcc 10.2.1 20210110 (Debian 10.2.1-6)
2022/11/30 18:43:09 [notice] 1#1: OS: Linux 5.10.104-linuxkit
2022/11/30 18:43:09 [notice] 1#1: getrlimit(RLIMIT_NOFILE): 1048576:1048576
2022/11/30 18:43:09 [notice] 1#1: start worker processes
2022/11/30 18:43:09 [notice] 1#1: start worker process 30
2022/11/30 18:43:09 [notice] 1#1: start worker process 31
2022/11/30 18:43:09 [notice] 1#1: start worker process 32
2022/11/30 18:43:09 [notice] 1#1: start worker process 33
2022/11/30 18:43:09 [notice] 1#1: start worker process 34
2022/11/30 18:43:09 [notice] 1#1: start worker process 35
172.17.0.1 - - [30/Nov/2022:18:43:53 +0000] "GET / HTTP/1.1" 200 615 "-" "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36" "-"
172.17.0.1 - - [30/Nov/2022:18:43:53 +0000] "GET /favicon.ico HTTP/1.1" 404 555 "http://127.0.0.1:63614/" "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36" "-"
2022/11/30 18:43:53 [error] 30#30: *1 open() "/usr/share/nginx/html/favicon.ico" failed (2: No such file or directory), client: 172.17.0.1, server: localhost, request: "GET /favicon.ico HTTP/1.1", host: "127.0.0.1:63614", referrer: "http://127.0.0.1:63614/"
```

You can reload your browser window showing the NginX welcome page to see new log entries appear. When you're satisfied, press Ctrl-C to end the stream.

Lastly, run:

```bash
kurtosis enclave dump "$ENCLAVE_ID" enclave-output
```

Kurtosis will dump a snapshot of the enclave's logs and container specs to the `enclave-output` directory. This can be useful for quickly sharing debugging information with your coworkers.

Write A Simple Starlark Script
-----------------------------------
We've used the CLI and some debugging tools, so let's start using [Kurtosis' Starlark environment definition language][starlark-explanation].

Create and `cd` into a new working directory:

```bash
mkdir my-kurtosis-package && cd my-kurtosis-package
```

Create a new Starlark file called `main.star` with the following contents:

```python
def run(plan, args):
    plan.add_service(
        "my-nginx",
        config = struct(
            image = "nginx:latest",
            ports = {
                "http": PortSpec(number = 80),
            },
        ),
    )
```

The commands in this file will do the same thing as the `service add` command you ran earlier, but they are now [infrastructure-as-code](https://en.wikipedia.org/wiki/Infrastructure_as_code).

Run the following:

```bash
kurtosis run main.star --dry-run
```

Because the `--dry-run` flag was specified, Kurtosis will read the file and show the instructions it would execute without executing them:

```
INFO[2022-11-30T15:48:18-03:00] Creating a new enclave for Starlark to run inside...
INFO[2022-11-30T15:48:20-03:00] Enclave 'long-water' created successfully

> add_service service_id="my-nginx"

Starlark code successfully executed. No output was returned
INFO[2022-11-30T15:48:20-03:00] ===================================================
INFO[2022-11-30T15:48:20-03:00] ||          Created enclave: long-water          ||
INFO[2022-11-30T15:48:20-03:00] ===================================================
```

Remove the `--dry-run` flag and execute the script:

```bash
kurtosis run main.star
```

The output will look similar to the dry run, but the `add_service` instruction now returns information about the service it started:

```
INFO[2022-11-30T15:48:59-03:00] Creating a new enclave for Starlark to run inside...
INFO[2022-11-30T15:49:02-03:00] Enclave 'cool-meadow' created successfully

> add_service service_id="my-nginx"
Service 'my-nginx' added with internal ID 'my-nginx-1669834142'

Starlark code successfully executed. No output was returned
INFO[2022-11-30T15:49:04-03:00] ====================================================
INFO[2022-11-30T15:49:04-03:00] ||          Created enclave: cool-meadow          ||
INFO[2022-11-30T15:49:04-03:00] ====================================================
```

Inspecting the enclave will now print the service inside:

```
Enclave ID:                           cool-meadow
Enclave Status:                       RUNNING
Creation Time:                        Wed, 30 Nov 2022 15:48:59 -03
API Container Status:                 RUNNING
API Container Host GRPC Port:         127.0.0.1:63678
API Container Host GRPC Proxy Port:   127.0.0.1:63679

========================================== User Services ==========================================
GUID                  ID         Ports                             Status
my-nginx-1669834142   my-nginx   http: 80/tcp -> 127.0.0.1:63684   RUNNING
```

Just like the service added via the CLI, the same Kurtosis debugging tools are available for enclaves created via Starlark.

Create A Dependency
-------------------
Now that you've seen the basics, let's define a system where one service depends on another service.

Replace your `main.star` contents with the following:

```python
def run(plan, args):
    rest_service = plan.add_service(
        "hello-world",
        config = struct(
            image = "vad1mo/hello-world-rest",
            ports = {
                "http": PortSpec(number = 5050),
            },
        ),
    )

    nginx_conf_data = {
        "HelloWorldIpAddress": rest_service.ip_address,
        "HelloWorldPort": rest_service.ports["http"].number,
    }

    nginx_conf_template = """
    server {
        listen       80;
        listen  [::]:80;
        server_name  localhost;

        location / {
            root   /usr/share/nginx/html;
            index  index.html index.htm;
        }

        # redirect server error pages to the static page /50x.html
        #
        error_page   500 502 503 504  /50x.html;
        location = /50x.html {
            root   /usr/share/nginx/html;
        }

        # Reverse proxy configuration (note the template values!)
        location /sample{
          proxy_pass http://{{ .HelloWorldIpAddress }}:{{ .HelloWorldPort }}/sample;
        }
    }
    """

    nginx_config_file_artifact = plan.render_templates(
        name = "nginx-artifact"
        config = {
            "default.conf": struct(
                template = nginx_conf_template,
                data = nginx_conf_data,
            )
        }
    )

    plan.add_service(
        "my-nginx",
        config = struct(
            image = "nginx:latest",
            ports = {
                "http": PortSpec(number = 80),
            },
            files = {
                "/etc/nginx/conf.d": nginx_config_file_artifact,
            }
        ),
    )
```

Run the Starlark script again:

```bash
kurtosis run main.star
```

Now inspect the enclave that got created. You'll see that two services, `my-nginx` and `hello-world`, have been added now:

```
Enclave ID:                           late-bird
Enclave Status:                       RUNNING
Creation Time:                        Wed, 30 Nov 2022 15:54:38 -03
API Container Status:                 RUNNING
API Container Host GRPC Port:         127.0.0.1:63729
API Container Host GRPC Proxy Port:   127.0.0.1:63730

========================================== User Services ==========================================
GUID                     ID            Ports                               Status
hello-world-1669834480   hello-world   http: 5050/tcp -> 127.0.0.1:63735   RUNNING
my-nginx-1669834483      my-nginx      http: 80/tcp -> 127.0.0.1:63749     RUNNING
```

Now in your browser open the `my-nginx` endpoint with the `/sample` URL path (e.g. `http://127.0.0.1:63749/sample`, though your URL will be different). You'll see the `hello-world` service responding through the NginX proxy that we've configured:

```
/ - Hello sample! Host:078807a8a776/172.17.0.11
```

Your Starlark script defined a set of instructions - a [plan][plan-reference] - for building the environment. This plan was:

1. Start a `hello-world` service, listening on port `5050`
1. Render a NginX config file using a template and the IP address and port of the `hello-world` service
1. Start the `my-nginx` service with the NginX config file mounted at `/etc/nginx/conf.d/default.conf`

Kurtosis read this plan, [ran pre-flight validation on it][multi-phase-runs-reference] to catch common errors (e.g. referencing container images or services or ports that don't exist), and started the environment you specified.

These instructions are just the beginning, however - there are [many more instructions available][starlark-instructions-reference].

Interlude
---------
We've started a few enclaves at this point, and `kurtosis enclave ls` will display something like the following:

```
EnclaveID        Status    Creation Time
wandering-frog   RUNNING   Wed, 30 Nov 2022 15:39:52 -03
long-water       RUNNING   Wed, 30 Nov 2022 15:48:18 -03
cool-meadow      RUNNING   Wed, 30 Nov 2022 15:48:59 -03
late-bird        RUNNING   Wed, 30 Nov 2022 15:54:38 -03
```

Enclaves themselves have very little overhead and are cheap to create, but the services inside the enclaves will naturally consume resources. Clean up your Kurtosis cluster now:

```
kurtosis clean -a
```

The `-a` flag indicates that even running enclaves should be removed. If you prefer to manage enclaves individually, this can be done with the `kurtosis enclave stop` and `kurtosis enclave rm` commands.

Before we continue, let's review what we've learned so far. We've:

1. Seen how the Kurtosis CLI can manage enclaves and services
1. Played with various debugging tools that the Kurtosis engine provides
1. Used Starlark to define an infrastructure-as-code environment
1. Defined a simple app that contained service dependencies and template-rendering

Use Resources In Starlark
-----------------------------------
It would be very cumbersome if your entire environment definition needed to fit in a single Starlark file, and we already see how the NginX config template makes the Starlark harder to read. Let's fix this.

First, create a file called `kurtosis.yml` next to your `main.star` file with the following contents, replacing `YOUR-GITHUB-USERNAME` with your GitHub username:

```yaml
name: "github.com/YOUR-GITHUB-USERNAME/my-kurtosis-package"
```

This is [a Kurtosis package manifest][kurtosis-yml-reference], which transforms your directory into [a Kurtosis package][packages-reference] and allows all Starlark scripts inside to use external dependencies.

Now create a file called `default.conf.tmpl` next to your `main.star` with the NginX config file contents (copied from the Starlark script):

```
server {
    listen       80;
    listen  [::]:80;
    server_name  localhost;

    location / {
        root   /usr/share/nginx/html;
        index  index.html index.htm;
    }

    # redirect server error pages to the static page /50x.html
    #
    error_page   500 502 503 504  /50x.html;
    location = /50x.html {
        root   /usr/share/nginx/html;
    }

    # Reverse proxy configuration (note the template values!)
    location /sample{
      proxy_pass http://{{ .HelloWorldIpAddress }}:{{ .HelloWorldPort }}/sample;
    }
}
```

Finally, replace your `main.star` with the following, replacing `YOUR-GITHUB-USERNAME` in the first line with your GitHub username:

```python
nginx_conf_template = read_file("github.com/YOUR-GITHUB-USERNAME/my-kurtosis-package/default.conf.tmpl")

def run(plan, args):
    rest_service = plan.add_service(
        "hello-world",
        config = struct(
            image = "vad1mo/hello-world-rest",
            ports = {
                "http": PortSpec(number = 5050),
            },
        ),
    )

    nginx_conf_data = {
        "HelloWorldIpAddress": rest_service.ip_address,
        "HelloWorldPort": rest_service.ports["http"].number,
    }

    nginx_config_file_artifact = plan.render_templates(
        config = {
            "default.conf": struct(
                template = nginx_conf_template,
                data = nginx_conf_data,
            )
        }
    )

    plan.add_service(
        "my-nginx",
        config = struct(
            image = "nginx:latest",
            ports = {
                "http": PortSpec(number = 80),
            },
            files = {
                "/etc/nginx/conf.d": nginx_config_file_artifact,
            }
        ),
    )
```

Take note that:

- The template contents are now being imported using the `read_file` [Starlark instruction][starlark-instructions-reference].
- The template file is referenced using a URL-like syntax; [this is called a "locator"][locators-reference] and is how Starlark files include external resources.
- Because our directory is now a Kurtosis package due to the `kurtosis.yml` file and the package has a `main.star` file with a `run` function, our directory is now a [runnable Kurtosis package][runnable-packages-reference].

Because our directory is a runnable Kurtosis package, we now run the package by specifying the package directory (the directory with the `kurtosis.yml`):

```bash
kurtosis run .
```

This will create the same `hello-world` and `my-nginx` services, but using external resources.

Parameterize Your Package
-------------------------------------
Notice that the `run` function in the `main.star` has an `args` argument. This allows you to parameterize your Kurtosis package.

Replace your `main.star` with the following, replacing `YOUR-GITHUB-USERNAME` in the first line with your GitHub username:

```python
nginx_conf_template = read_file("github.com/YOUR-GITHUB-USERNAME/my-kurtosis-package/default.conf.tmpl")

def run(plan, args):
    rest_service = plan.add_service(
        "hello-world",
        config = struct(
            image = "vad1mo/hello-world-rest",
            ports = {
                "http": PortSpec(number = 5050),
            },
        ),
    )

    nginx_conf_data = {
        "HelloWorldIpAddress": rest_service.ip_address,
        "HelloWorldPort": rest_service.ports["http"].number,
    }

    nginx_config_file_artifact = plan.render_templates(
        config = {
            "default.conf": struct(
                template = nginx_conf_template,
                data = nginx_conf_data,
            )
        }
    )

    nginx_count = 1
    if hasattr(args, "nginx_count"):
        nginx_count = args.nginx_count

    for i in range(0, nginx_count):
        plan.add_service(
            "my-nginx-" + str(i),
            config = struct(
                image = "nginx:latest",
                ports = {
                    "http": PortSpec(number = 80),
                },
                files = {
                    "/etc/nginx/conf.d": nginx_config_file_artifact,
                }
            ),
        )
```

Now run the package again, passing in a JSON object for args:

```bash
kurtosis run . '{"nginx_count": 3}'
```

After execution, inspecting the enclave will reveal that three NginX services have been started:

```
Enclave ID:                           quiet-morning
Enclave Status:                       RUNNING
Creation Time:                        Wed, 30 Nov 2022 16:02:41 -03
API Container Status:                 RUNNING
API Container Host GRPC Port:         127.0.0.1:63802
API Container Host GRPC Proxy Port:   127.0.0.1:63803

========================================== User Services ==========================================
GUID                     ID            Ports                               Status
hello-world-1669834964   hello-world   http: 5050/tcp -> 127.0.0.1:63808   RUNNING
my-nginx-0-1669834966    my-nginx-0    http: 80/tcp -> 127.0.0.1:63812     RUNNING
my-nginx-1-1669834970    my-nginx-1    http: 80/tcp -> 127.0.0.1:63816     RUNNING
my-nginx-2-1669834975    my-nginx-2    http: 80/tcp -> 127.0.0.1:63820     RUNNING
```

Each one of these NginX services works identically.

Note that we used the `hasattr` Starlark builtin to check if `args.nginx_count` exists, so the package will continue to work if you omit the `args` argument to `kurtosis run`.

Publish & Consume Your Package
------------------------------------------
[Kurtosis packages][packages-reference] are designed to be trivial to share and consume, so let's do so now.

First, create a repo in your personal GitHub called `my-kurtosis-package`.

Second, run the following in your Kurtosis `my-kurtosis-package` directory (the directory with `kurtosis.yml`), replacing `YOUR-GITHUB-USERNAME` with your GitHub username:

```bash
git init && git remote add origin https://github.com/YOUR-GITHUB-USERNAME/my-kurtosis-package.git
```

This makes your directory a Git repo associated with the new GitHub repo you just created.

Now commit and push your changes:

```bash
git add . && git commit -m "Initial commit" && git push origin master
```

Your package is now published, and available to anyone using Kurtosis. To use it, anyone can run the following, replacing `YOUR-GITHUB-USERNAME` with your GitHub username:


```bash
kurtosis run github.com/YOUR-GITHUB-USERNAME/my-kurtosis-package
```

Starlark code is composable, meaning you can import Starlark inside other Starlark (in keeping with [the properties of a reusable environment definition][reusable-environment-definitions-reference]). 

To see this in action, create a new `new-kurtosis-package` directory outside of your `my-kurtosis-package` directory:

```bash
cd ~ && mkdir new-kurtosis-package && cd new-kurtosis-package
```

<!-- TODO refactor this when dependencies are specified in the kurtosis.yml file -->
Add a `kurtosis.yml` manifest make the directory [a Kurtosis package][packages-reference]:

```yaml
name: "github.com/test/test"
```

Next to the `kurtosis.yml`, add a `main.star`, replacing `YOUR-GITHUB-USERNAME` in the first line with your GitHub username:

```python
my_package = import_module("github.com/YOUR-GITHUB-USERNAME/my-kurtosis-package/main.star")

def run(plan, args):
    my_package.run(plan, struct(nginx_count = 3))
```

Finally, run it by referencing the directory containing the new `kurtosis.yml`:

```
kurtosis run .
```

Kurtosis will handle the importing of your already-published package, allowing anyone to use your environment definition.

Conclusion
----------
In this tutorial we've seen:

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

<!-- TODO add link to how-to guide on how to use Kurtosis in Go/TS tests -->



<!------------------------------- ONLY LINKS BELOW HERE ------------------------------------->
[installing-tab-complete]: ./guides/adding-tab-completion.md
[enclaves-explanation]: ./explanations/architecture.md#enclaves
[services-explanation]: ./explanations/architecture.md#services
[starlark-explanation]: ./explanations/starlark.md
[starlark-instructions-reference]: ./reference/starlark-instructions.md
[multi-phase-runs-reference]: ./reference/multi-phase-runs.md
[kurtosis-yml-reference]: ./reference/kurtosis-yml.md
[packages-reference]: ./reference/packages.md
[runnable-packages-reference]: ./reference/packages.md#runnable-packages
[locators-reference]: ./reference/locators.md
[plan-reference]: ./reference/plan.md
[reusable-environment-definitions-reference]: ./explanations/reusable-environment-definitions.md
[cli-reference]: ./reference/cli.md
[kurtosis-managed-packages]: https://github.com/kurtosis-tech?q=in%3Aname+package&type=all&language=&sort=
[wild-kurtosis-packages]: https://github.com/search?q=filename%3Akurtosis.yml&type=code
[architecture-explanation]: ./explanations/architecture.md
