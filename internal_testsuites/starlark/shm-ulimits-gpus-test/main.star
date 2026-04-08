def run(plan):
    plan.add_service(
        name = "shm-test",
        config = ServiceConfig(
            image = "busybox:1.36",
            cmd = ["sh", "-c", "df -k /dev/shm | tail -1"],
            shm_size = 128,
        ),
    )
    plan.print("shm-test: shm_size=128 service started")

    plan.add_service(
        name = "ulimits-memlock",
        config = ServiceConfig(
            image = "busybox:1.36",
            cmd = ["sh", "-c", "ulimit -l"],
            ulimits = {"memlock": -1},
        ),
    )
    plan.print("ulimits-memlock: memlock=unlimited service started")

    plan.add_service(
        name = "ulimits-nofile",
        config = ServiceConfig(
            image = "busybox:1.36",
            cmd = ["sh", "-c", "ulimit -n"],
            ulimits = {"nofile": 65536},
        ),
    )
    plan.print("ulimits-nofile: nofile=65536 service started")

    plan.add_service(
        name = "ulimits-combined",
        config = ServiceConfig(
            image = "busybox:1.36",
            cmd = ["sh", "-c", "ulimit -l && ulimit -n"],
            ulimits = {"memlock": -1, "nofile": 65536},
        ),
    )
    plan.print("ulimits-combined: memlock=unlimited + nofile=65536 service started")
