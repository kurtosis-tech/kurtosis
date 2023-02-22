# this is used by the breaking-change-description-title-update.py script to get
# all but the first conventional commit
import sys

def description():
    line = sys.argv[1]
    split_line = line.split("\n")
    if len(split_line) > 2:
        return split_line[1]
    return ""


if __name__ == "__main__":
    print(description())
