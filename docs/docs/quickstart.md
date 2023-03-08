---
title: Kurtosis Quickstart
sidebar_label: Quickstart
slug: /quickstart
---

These instructions will give you a brief quickstart of Kurtosis and should take 15 minutes. If you prefer, schedule a personalized onboarding session with us [here](https://calendly.com/d/zgt-f2c-66p/kurtosis-onboarding). 

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
INFO[2023-01-26T15:10:22Z] Creating new enclave...                      
INFO[2023-01-26T15:10:32Z] ==================================================== 
INFO[2023-01-26T15:10:32Z] ||          Created enclave: patient-sun          || 
INFO[2023-01-26T15:10:32Z] ==================================================== 
```

Now, type the following but don't press ENTER yet:

```bash
kurtosis enclave inspect
```

If [you have tab completion installed][installing-tab-complete], you can now press TAB to tab-complete your enclave's name (which will be different than `patient-sun`).

If you don't have tab completion installed, you'll need to paste the enclave ID from the `Created enclave:` line outputted above (yours will be different than `wandering-frog`).

:::tip
[All enclave ID arguments][cli-reference] can be tab-completed.
:::

Press ENTER, and you should receive an output like so:

```
UUID:                                 edf4c085912d
Enclave Name:                         patient-sun
Enclave Status:                       RUNNING
Creation Time:                        Thu, 26 Jan 2023 15:10:24 GMT
API Container Status:                 RUNNING
API Container Host GRPC Port:         127.0.0.1:55529
API Container Host GRPC Proxy Port:   127.0.0.1:55530

========================================== User Services ==========================================
UUID   Name   Ports   Status
```

`kurtosis enclave inspect` is the way to investigate an enclave.

If you ever forget your enclave name or don't feel like using tab completion, you can always run the following:

```bash
kurtosis enclave ls
```

This will print all the enclaves inside your Kurtosis cluster:

```
UUID           Name                                       Status    Creation Time
edf4c085912d   patient-sun                                RUNNING   Thu, 26 Jan 2023 15:10:24 GMT
```

Now run the following to store your enclave ID in a variable, replacing `YOUR_ENCLAVE_NAME_HERE` with your enclave's name.

```bash
ENCLAVE_IDENTIFIER="YOUR_ENCLAVE_NAME_HERE"
```

We'll use this variable so that you can continue to copy-and-paste code blocks in the next section.

Start A Service
---------------------------
Distributed applications are composed of [services][services-explanation]. Here we'll start a simple service, and see some of the options Kurtosis has for debugging.

Enter this command:

```bash
kurtosis service add "$ENCLAVE_IDENTIFIER" my-nginx nginx:latest --ports http=80
```

You should see output similar to the following:

```
Ports Bindings:
Name:      my-nginx
UUID:      ad28326bf014
   http:   80/tcp -> 127.0.0.1:55571
```

Now inspect your enclave again:

```bash
kurtosis enclave inspect "$ENCLAVE_IDENTIFIER"
```

You should see a new service with the service name `my-nginx` in your enclave:

```
UUID:                                 edf4c085912d
Enclave Name:                         patient-sun
Enclave Status:                       RUNNING
Creation Time:                        Thu, 26 Jan 2023 15:10:24 GMT
API Container Status:                 RUNNING
API Container Host GRPC Port:         127.0.0.1:55529
API Container Host GRPC Proxy Port:   127.0.0.1:55530

========================================== User Services ==========================================
UUID           Name       Ports                             Status
ad28326bf014   my-nginx   http: 80/tcp -> 127.0.0.1:55571   RUNNING
```

Kurtosis binds all service ports to ephemeral ports on your local machine. Copy the `127.0.0.1:XXXXX` address into your browser (yours will be different), and you should see a welcome message from your NginX service running inside the enclave you created.

Now paste the following but don't press ENTER yet:

```bash
kurtosis service shell "$ENCLAVE_IDENTIFIER"
```

If you have tab completion installed, press TAB. The service UUID & name of the NginX service will be completed (which in this case was `my-nginx`, but yours will be different).

If you don't have tab completion installed, paste in the service name of the NginX service from the `enclave inspect` output above (which was `my-nginx`, but yours will be different).

:::tip
Like enclave names, [all service UUID arguments][cli-reference] can be tab-completed.
:::

Press ENTER, and you'll be logged in to a shell on the container:

```
Found bash on container; creating bash shell...
root@89f8447e9f62:/#
```

Kurtosis will try to give you a `bash` shell, but will drop down to `sh` if `bash` doesn't exist on the container.

Feel free to explore, and enter `exit` or press Ctrl-D when you're done.

Now enter the following but don't press ENTER:

```bash
kurtosis service logs -f "$ENCLAVE_IDENTIFIER"
```

Once again, you can use tab completion to fill the service UUID if you have it enabled. If not, you'll need to copy-paste the service UUID as the last argument.

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
2023/01/26 15:14:42 [notice] 1#1: using the "epoll" event method
2023/01/26 15:14:42 [notice] 1#1: nginx/1.23.3
2023/01/26 15:14:42 [notice] 1#1: built by gcc 10.2.1 20210110 (Debian 10.2.1-6) 
2023/01/26 15:14:42 [notice] 1#1: OS: Linux 5.15.49-linuxkit
2023/01/26 15:14:42 [notice] 1#1: getrlimit(RLIMIT_NOFILE): 1048576:1048576
2023/01/26 15:14:42 [notice] 1#1: start worker processes
2023/01/26 15:14:42 [notice] 1#1: start worker process 29
2023/01/26 15:14:42 [notice] 1#1: start worker process 30
2023/01/26 15:14:42 [notice] 1#1: start worker process 31
2023/01/26 15:14:42 [notice] 1#1: start worker process 32
```

You can reload your browser window showing the NginX welcome page to see new log entries appear. When you're satisfied, press Ctrl-C to end the stream.

Lastly, run:

```bash
kurtosis enclave dump "$ENCLAVE_IDENTIFIER" enclave-output
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
        config = ServiceConfig(
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
INFO[2023-01-26T17:47:34Z] Creating a new enclave for Starlark to run inside... 
INFO[2023-01-26T17:47:37Z] Enclave 'winter-night' created successfully  

> add_service service_name="my-nginx" config=ServiceConfig(image="nginx:latest", ports={"http": PortSpec(number=80, transport_protocol="TCP", application_protocol="")})

Starlark code successfully run in dry-run mode. No output was returned.
INFO[2023-01-26T17:47:38Z] ===================================================== 
INFO[2023-01-26T17:47:38Z] ||          Created enclave: winter-night          || 
INFO[2023-01-26T17:47:38Z] =====================================================
```

Remove the `--dry-run` flag and execute the script:

```bash
kurtosis run main.star
```

The output will look similar to the dry run, but the `add_service` instruction now returns information about the service it started:

```
INFO[2023-01-26T17:47:01Z] Creating a new enclave for Starlark to run inside... 
INFO[2023-01-26T17:47:06Z] Enclave 'muddy-grass' created successfully   

> add_service service_name="my-nginx" config=ServiceConfig(image="nginx:latest", ports={"http": PortSpec(number=80, transport_protocol="TCP", application_protocol="")})
Service 'my-nginx' added with service UUID '490ac1e94356470a9da925e6c44722df'

Starlark code successfully run. No output was returned.
INFO[2023-01-26T17:47:08Z] ==================================================== 
INFO[2023-01-26T17:47:08Z] ||          Created enclave: muddy-grass          || 
INFO[2023-01-26T17:47:08Z] ==================================================== 
```

Inspecting the enclave will now print the service inside:

```
UUID:                                 edf501e234a1
Enclave Name:                         muddy-grass
Enclave Status:                       RUNNING
Creation Time:                        Thu, 26 Jan 2023 17:47:01 GMT
API Container Status:                 RUNNING
API Container Host GRPC Port:         127.0.0.1:57301
API Container Host GRPC Proxy Port:   127.0.0.1:57302

========================================== User Services ==========================================
UUID           Name       Ports                             Status
490ac1e94356   my-nginx   http: 80/tcp -> 127.0.0.1:57307   RUNNING
```

Just like the service added via the CLI, the same Kurtosis debugging tools are available for enclaves created via Starlark.

Create A Dependency
-------------------
Now that you've seen the basics, let's define a system where one service depends on another service.

Replace your `main.star` contents with the following:

```python
def run(plan, args):
    web_server = plan.add_service(
        "hello-world",
        config = ServiceConfig(
            image = "httpd",
            ports = {
                "http": PortSpec(number = 80),
            },
        ),
    )

    nginx_conf_data = {
        "HelloWorldIpAddress": web_server.ip_address,
        "HelloWorldPort": web_server.ports["http"].number,
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
          proxy_pass http://{{ .HelloWorldIpAddress }}:{{ .HelloWorldPort }}/;
        }
    }
    """

    nginx_config_file_artifact = plan.render_templates(
        name = "nginx-artifact",
        config = {
            "default.conf": struct(
                template = nginx_conf_template,
                data = nginx_conf_data,
            )
        },
    )

    plan.add_service(
        "my-nginx",
        config = ServiceConfig(
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
UUID:                                 edf351b63ba6
Enclave Name:                         shy-surf
Enclave Status:                       RUNNING
Creation Time:                        Thu, 26 Jan 2023 17:49:59 GMT
API Container Status:                 RUNNING
API Container Host GRPC Port:         127.0.0.1:57336
API Container Host GRPC Proxy Port:   127.0.0.1:57337

========================================== User Services ==========================================
UUID           Name          Ports                             Status
049a95f55a93   hello-world   http: 80/tcp -> 127.0.0.1:57357   RUNNING
a1b43c21f0e7   my-nginx      http: 80/tcp -> 127.0.0.1:57361   RUNNING
```

Now in your browser open the `my-nginx` endpoint with the `/sample` URL path (e.g. `http://127.0.0.1:63749/sample`, though your URL will be different). You'll see the `hello-world` service responding through the NginX proxy that we've configured:

```
It works!
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
UUID           Name              Status    Creation Time
edffa85eb3b1   twilight-shadow   RUNNING   Thu, 26 Jan 2023 15:33:10 GMT
edf78299bbbd   patient-field     RUNNING   Thu, 26 Jan 2023 16:50:12 GMT
edf1abc0bfc4   dry-thunder       RUNNING   Thu, 26 Jan 2023 16:50:22 GMT
edf501e234a1   muddy-grass       RUNNING   Thu, 26 Jan 2023 17:47:01 GMT
edf705c22be4   winter-night      RUNNING   Thu, 26 Jan 2023 17:47:34 GMT
edf351b63ba6   shy-surf          RUNNING   Thu, 26 Jan 2023 17:49:59 GMT
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
        config = ServiceConfig(
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
        name = "nginx-artifact",
        config = {
            "default.conf": struct(
                template = nginx_conf_template,
                data = nginx_conf_data,
            )
        },
    )

    plan.add_service(
        "my-nginx",
        config = ServiceConfig(
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
        config = ServiceConfig(
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
        name = "nginx-artifact",
        config = {
            "default.conf": struct(
                template = nginx_conf_template,
                data = nginx_conf_data,
            )
        },
    )

    nginx_count = 1
    if hasattr(args, "nginx_count"):
        nginx_count = args.nginx_count

    for i in range(0, nginx_count):
        plan.add_service(
            "my-nginx-" + str(i),
            config = ServiceConfig(
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
UUID:                                 edf011afc46c
Enclave Name:                         purple-smoke
Enclave Status:                       RUNNING
Creation Time:                        Thu, 26 Jan 2023 17:53:28 GMT
API Container Status:                 RUNNING
API Container Host GRPC Port:         127.0.0.1:57429
API Container Host GRPC Proxy Port:   127.0.0.1:57430

========================================== User Services ==========================================
UUID           Name          Ports                               Status
0925dff98df7   hello-world   http: 5050/tcp -> 127.0.0.1:57437   RUNNING
3cf9f9fdd5e6   my-nginx-2    http: 80/tcp -> 127.0.0.1:57451     RUNNING
60ef364bb7c5   my-nginx-1    http: 80/tcp -> 127.0.0.1:57447     RUNNING
8ccb57e0e778   my-nginx-0    http: 80/tcp -> 127.0.0.1:57441     RUNNING
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

:::tip
If you want to run a non-master branch, tag or commit use the following syntax
`kurtosis run github.com/package-author/package-repo@tag-branch-commit`
:::

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

:::info
Get a personalized on-boarding session with us [here](https://calendly.com/d/zgt-f2c-66p/kurtosis-onboarding). 
:::

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
[cli-reference]: ./reference/cli/cli.md
[kurtosis-managed-packages]: https://github.com/kurtosis-tech?q=in%3Aname+package&type=all&language=&sort=
[wild-kurtosis-packages]: https://github.com/search?q=filename%3Akurtosis.yml&type=code
[architecture-explanation]: ./explanations/architecture.md
