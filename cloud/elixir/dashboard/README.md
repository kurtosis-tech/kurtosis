# Dashboard

To start your Phoenix server:

  * Run `mix setup` to install and setup dependencies
  * Start Phoenix endpoint with `mix phx.server` or inside IEx with `iex -S mix phx.server`

Now you can visit [`localhost:4000`](http://localhost:4000) from your browser.

Ready to run in production? Please [check our deployment guides](https://hexdocs.pm/phoenix/deployment.html).

## Dev DB

For the first time:
```bash
initdb -D .pg_data/knowit_dev
pg_ctl -D .pg_data/knowit_dev -l .pg_data/logfile start
createuser postgres
createdb dashboard_dev
```

After restarting the process:
```bash
pg_ctl -D .pg_data/knowit_dev -l .pg_data/logfile start
```

## Generate API bindings

Install the `protoc` plug-in for Elixir
```bash
mix escript.install hex protobuf
export PATH=$PATH:/Users/edgar/.mix/escripts
```

Generate/Update the Elixir bindings:
```bash
protoc --elixir_out=plugins=grpc:./lib/api -I /code/kurtosis/api/protobuf/core/ api_container_service.proto
protoc --elixir_out=plugins=grpc:./lib/api -I /code/kurtosis/api/protobuf/engine engine_service.proto
```

## Learn more

  * Official website: https://www.phoenixframework.org/
  * Guides: https://hexdocs.pm/phoenix/overview.html
  * Docs: https://hexdocs.pm/phoenix
  * Forum: https://elixirforum.com/c/phoenix-forum
  * Source: https://github.com/phoenixframework/phoenix
