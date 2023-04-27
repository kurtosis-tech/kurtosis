# This script is run by release-please "change-versions.yml"
# from the root of the docs folder

import json
import os
import shutil

from collections import namedtuple

VERSIONS_FILE_NAME = "versions.json"

Version = namedtuple("Version", "major minor patch")

HISTORICAL_VERSIONS_TO_KEEP = 5
MOST_TO_KEEP_OF_LATEST_VERSION = 5

def main():
    versions_to_keep = []
    versioned_docs_to_delete = []
    versioned_sidebars_to_delete = []
    with open(VERSIONS_FILE_NAME, "r") as available_versions:
        available_versions_list = json.loads(available_versions.read())
        current_version = str_to_version(available_versions_list[0])
        latest_version = current_version
        historical_versions_kept = 0
        current_versions_kept = 0
        for version_str in available_versions_list:
            version = str_to_version(version_str)
            # if the version being processed is the same as the latest major and minor version do nothing
            if version.major == latest_version.major and version.minor == latest_version.minor:
                if MOST_TO_KEEP_OF_LATEST_VERSION > 0 and current_versions_kept >= MOST_TO_KEEP_OF_LATEST_VERSION:
                    continue
                current_versions_kept += 1
                versions_to_keep.append(version_str)
                continue

            if HISTORICAL_VERSIONS_TO_KEEP > 0 and historical_versions_kept >= HISTORICAL_VERSIONS_TO_KEEP:
                break

            # as we go down the list if the minor version changes then we are on the highest patch with the minor
            # version so, we keep it
            if version.minor != current_version.minor:
                versions_to_keep.append(version_str)
                historical_versions_kept += 1
                current_version = version
                continue

    assert len(versions_to_keep) <= MOST_TO_KEEP_OF_LATEST_VERSION + HISTORICAL_VERSIONS_TO_KEEP

    versions_to_delete = set(available_versions_list).difference(set(versions_to_keep))
    for version in versions_to_delete:
        versioned_docs_to_delete.append(f"versioned_docs/version-{version}")
        versioned_sidebars_to_delete.append(f"versioned_sidebars/version-{version}-sidebars.json")

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
