SERVICE_NAME = "service-example"

def run(plan, args={}):
    artifact_name_1 = plan.upload_files("./folder-1/file-1.txt")
    artifact_name_2 = plan.upload_files("./folder-2/file-2.txt")

    plan.add_service(
        SERVICE_NAME,
        config=ServiceConfig(
            image="alpine:latest",
            cmd=["/bin/sh", "-c", "sleep 999"],
            files={
                "/files": Directory(
                    artifact_names=[artifact_name_1,artifact_name_2]
                )
            },
        ),
    )

    plan.exec(
        service_name=SERVICE_NAME,
        recipe=ExecRecipe([
            "/bin/sh",
            "-c",
            "cat /files/file-1.txt",
        ])
    )

    plan.exec(
        service_name=SERVICE_NAME,
        recipe=ExecRecipe([
            "/bin/sh",
            "-c",
            "cat /files/file-2.txt",
        ])
    )
