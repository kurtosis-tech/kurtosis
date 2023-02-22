# this is used by the breaking-change-description-title-update.py script to get
# the first conventional commit
import sys

def title():
    line = sys.argv[1]
    split_line = line.split("\n")
    if split_line:
        return split_line[-1]
    return ""

if __name__ == "__main__":
    print(title())
