#!/usr/bin/env python3
"""
Script to initially populate a (currently only) sqlite3 DB for specs-bot. The
DB will be filled with all label information for certain Issues/PRs that
already contain specific labels defined in LABELS. Alternatively, you can
leave LABEL as an empty list and it will download label information for all
Issues/PRs.
"""

from github import Github
import sqlite3

ACCESS_TOKEN = ""
REPO = ""
LABELS = [] # Only issues/PRs with these label names will be recorded in the DB. Empty for no filter
DB_PATH = "./specs-bot.db"

def main():
    global ACCESS_TOKEN
    global REPO
    global LABELS
    global DB_PATH

    # Login to Github
    github = Github(ACCESS_TOKEN)
    print("Connected to Github")

    # Get repo object
    repo = github.get_repo(REPO)

    # Connect to Sqlite3 DB
    conn = sqlite3.connect(DB_PATH)
    c = conn.cursor()

    # Create "proposal_state table if it does not exist"
    c.execute("""CREATE TABLE IF NOT EXISTS proposal_state (
        number INTEGER PRIMARY KEY,
        labels TEXT NOT NULL
    )""")

    # Get repo labels and map to our user-defined label names
    print("Downloading labels for %s" % REPO)
    label_objects = []
    for label in repo.get_labels():
        if label.name in LABELS:
            label_objects.append(label)

    # Get PRs with special label
    print("Downloading issues and PRs with labels: %s" % str([label.name for label in label_objects]))
    for issue in repo.get_issues(labels=label_objects):
        labels_str = ""
        for label in issue.labels:
            labels_str += label.name + ","

        # Remove final comma
        labels_str = labels_str[:-1]

        print("Inserting %s:%s" % (issue.number, labels_str))

        c.execute("INSERT OR REPLACE INTO proposal_state (number, labels) VALUES (?, ?)", (issue.number, labels_str))

    # Commit changes and close DB connection
    conn.commit()
    conn.close()

if __name__ == "__main__":
    main()