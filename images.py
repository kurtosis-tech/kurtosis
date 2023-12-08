MY_USER_NAME = "h4ck3kr3y"

import subprocess
import sys

image_tag = sys.argv[1]

def execute(command):
    return subprocess.check_output(command, shell=True, universal_newlines=True)

# login = execute(f"aws ecr get-login-password --region us-east-2 | docker login --username AWS --password-stdin {ECR_ADDRESS}")

returned_text = execute(f"docker image ls | grep {image_tag}")

for line in returned_text.split("\n"):
    if line == "":
        continue
    image, tag, image_id = line.split()[:3]
    author, repo = image.split("/")
    print(image, tag, image_id)

    if not image.startswith("kurtosistech"):
        continue

    execute(f"docker tag {image_id} {MY_USER_NAME}/{repo}:{tag}")
    execute(f"docker push {MY_USER_NAME}/{repo}:{tag}")


