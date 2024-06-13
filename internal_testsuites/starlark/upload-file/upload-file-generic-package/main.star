SERVICE_NAME = "service-example"


def run(plan, args={}):
    url = "github.com/kurtosis-tech/minimal-grpc-server/golang/scripts/build.sh"
    artifact_name_1 = plan.upload_files(url)

    service = plan.add_service(
        SERVICE_NAME,
        config=ServiceConfig(
            image="alpine:latest",
            cmd=["/bin/sh", "-c", "ls /files && sleep 9999"],
            files={"/files": Directory(artifact_names=[artifact_name_1])},
        ),
    )
