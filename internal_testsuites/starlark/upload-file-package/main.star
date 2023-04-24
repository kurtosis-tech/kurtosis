
# The test using this package will generate a 90MB file and place it at the root of this package
# named large-file.bin. This file is generated on the spot to avoid checking it into GitHub
def run(plan, args):
    plan.print("Starting upload")
    large_file_artifact_id = plan.upload_files("github.com/sample/sample-kurtosis-package/large-file.bin")

    plan.print("Upload finished - Comparing file hash to parameter")

    plan.add_service(
        name="dummy",
        config=ServiceConfig(
            image="docker/getting-started",
            files={
                "/home/file/": large_file_artifact_id
            }
        )
    )
    result = plan.exec(
        service_name="dummy",
        recipe=ExecRecipe(
            command=["sh", "-c", "md5sum /home/file/large-file.bin | awk '{print $1}'"]
        )
    )

    expected_file_hash = args["file_hash"]
    plan.assert(
        value=result["output"],
        assertion="==",
        target_value=expected_file_hash,
    )
