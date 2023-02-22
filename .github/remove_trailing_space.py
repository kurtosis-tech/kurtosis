# This is used by the pr-description-update action
# to trim trailing newlines and spaces in pr descriptions
# as doing so in bash proved tough

import sys

def remove():
    line = sys.argv[1]
    return line.strip()

if __name__ == "__main__":
    print(remove())
