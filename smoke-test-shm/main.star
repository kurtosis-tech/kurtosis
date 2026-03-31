def run(plan):
    # Start a service with 128 MB of /dev/shm and verify the size is correct.
    result = plan.add_service(
        name = "shm-test",
        config = ServiceConfig(
            image = "busybox:1.36",
            cmd = ["sh", "-c", "df -k /dev/shm | tail -1"],
            shm_size = 128,  # 128 MB
        ),
    )
    plan.print("shm-test service started")
