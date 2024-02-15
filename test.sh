set -euo pipefail
source core/server/scripts/_constants.env
dockerfile_filepath='core/server/Dockerfile'
version_build="$(./scripts/get-docker-tag.sh)"
version_to_publish="$(cat version.txt)"
echo "Version that was built: ${version_build}"
echo "Version that will be published: ${version_to_publish}"
image_name_with_version="${IMAGE_ORG_AND_REPO}:${version_build}"
image_name_to_publish_semver="${IMAGE_ORG_AND_REPO}:${version_to_publish}"
image_name_to_publish_latest="${IMAGE_ORG_AND_REPO}:latest"
push_to_dockerhub=true
scripts/docker-image-builder.sh "${push_to_dockerhub}" "${dockerfile_filepath}" "${image_name_with_version}" "${image_name_to_publish_semver}" "${image_name_to_publish_latest}"
