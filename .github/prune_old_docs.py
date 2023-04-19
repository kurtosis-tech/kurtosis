# This script is run by release-please "change-versions.yml"
# from the root of the docs folder

import json
import os
import shutil

from collections import namedtuple

VERSIONS_FILE_NAME = "versions.json"

Version = namedtuple("Version", "major minor patch")


def main():
    versions_to_keep = []
    versioned_docs_to_delete = []
    versioned_sidebars_to_delete = []
    with open(VERSIONS_FILE_NAME, "r") as available_versions:
        available_versions_json = json.loads(available_versions.read())
        current_version = str_to_version(available_versions_json[0])
        for version_str in available_versions_json:
            version = str_to_version(version_str)
            # if the version being processed is the same as the current major and minor version do nothing
            if version.major == current_version.major and version.minor == current_version.minor:
                versions_to_keep.append(version_str)
                continue

            # as we go down the list if the minor version changes then we are on the highest patch with the minor
            # version so we keep it
            if version.minor != current_version.minor:
                versions_to_keep.append(version_str)
                current_version  = version
                continue

            # if the minor version is the same as the previous one then we are on a lower patch
            if version.minor == current_version.minor:
                versioned_docs_to_delete.append(f"versioned_docs/version-{version_str}")
                versioned_sidebars_to_delete.append(f"versioned_sidebars/version-{version_str}-sidebars.json")

    for docs_folder_to_delete in versioned_docs_to_delete:
        shutil.rmtree(docs_folder_to_delete)

    for sidebar_to_remove in versioned_sidebars_to_delete:
        os.remove(sidebar_to_remove)

    with open(VERSIONS_FILE_NAME, 'w') as versions_file:
        versions_to_keep_json = json.dumps(versions_to_keep, indent=2)
        versions_file.write(versions_to_keep_json)
        versions_file.write("\n")


def str_to_version(version_str):
    split_str = version_str.split(".")
    return Version(*split_str)


if __name__ == "__main__":
    main()
