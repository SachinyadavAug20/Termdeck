---
format: 0.1
title: Git Under The Hood: The Architecture Beyond Commands
author: Sachin
align: left
theme: dracula
---

::align center
# Git Under The Hood

### Demystifying the Core Architecture for CS Students

A terminal presentation powered by **Termdeck**

---

# The Problem: Commands vs Mental Model

- You already know the basic commands:

- `git add .`

- `git commit -m "fixed stuff"`

- `git push origin main`

- But what actually happens when you run them?

- What is a commit? Where does it live?

- Why does **"Detached HEAD"** sound terrifying?

- *Core truth:* If you understand Git's underlying data structures, all 150+ Git commands become intuitive.

::notes
Hook the audience immediately. Acknowledge that most college tutorials teach Git backwards: they teach recipes to memorize rather than the 3 data structures that make everything click.

---

# Myth: Git Stores File Diffs

- **Old VCS (SVN, CVS)**:

- Stored delta changes (diffs):

- File A (v1) $\to$ +3 lines (v2) $\to$ -1 line (v3).

- **Git does NOT store diffs!**

- Git stores **Snapshots** of your entire project:

- Every commit is a full picture of your repository at that point in time.

- If a file didn't change in a commit, Git doesn't duplicate it—it just stores a pointer to the existing file.

- Git is fundamentally a **Content-Addressable Key-Value Store** with a Directed Acyclic Graph (DAG) on top.

::notes
Emphasize this distinction! Thinking Git stores diffs is the #1 reason students get confused by branching and rebasing.

---

# Content-Addressable Storage

- In a normal filesystem: You find data by **path** (`/docs/notes.txt`).

- In Git: You find data by its **Cryptographic Hash (SHA-1 / SHA-256)**.

::code lang=bash
  # Calculate Git's hash of any text content:
  echo "hello world" | git hash-object --stdin
  # Output: 3b18e512dba79e4c8300dd08aeb37f8e728b8dad

- **Properties of Content Addressing**:
  - Same content always generates the exact same 40-character hash.
  - Change even a single space or dot, and the hash changes completely.
  - Built-in data integrity: Corrupted files are instantly detected.

::notes
Ask the students: what is the time complexity of checking if two files are identical if you have their SHA hash? O(1) string comparison instead of O(N) byte comparison!

---

# The 4 Fundamental Git Objects

Everything inside your `.git/objects` folder is one of these 4 objects:

- **1. Blob (Binary Large Object)**:

- Just raw file bytes. No filename, no timestamps, no permissions.

- **2. Tree**:

- Represents a directory. Maps file names & permissions to Blob hashes or sub-Tree hashes.

- **3. Commit**:

- Contains author, committer, timestamp, commit message, pointer to root Tree, and parent Commit hash.

- **4. Tag / Annotated Tag**:

- A permanent pointer to a specific commit with a message.

---

# How Git Stores a Project: The Tree Graph

::code lang=
  [ Commit: "Initial commit" ]
               |
               v
       [ Root Tree (/) ]
         /           \
        v             v
 [ Blob: main.go ]  [ Subtree: /utils ]
                           |
                           v
                    [ Blob: math.go ]

- If you change `main.go` and commit:

- Git creates a new Blob for `main.go`.

- Reuses the existing Blob for `utils/math.go` (zero copy!).

- Creates a new Root Tree and a new Commit pointing back to Parent 1.

::notes
Walk through the diagram. Point out that unchanged files cost zero extra disk space because the new Tree simply points to the existing Blob hash.

---

# Looking Inside .git with Plumbing Commands

Most commands you use are **Porcelain** (user-friendly UI).

The engine runs on **Plumbing** commands:

::code lang=bash
  # 1. View the type of any object hash
  git cat-file -t 3b18e51
  # => blob

  # 2. Pretty-print the content of any hash
  git cat-file -p 3b18e51
  # => hello world

  # 3. Inspect a commit's internal metadata
  git cat-file -p HEAD
  # => tree 4b825dc642cb6eb9a060e54bf8d69288fbee4904
  # => parent 96da3b24f5a...
  # => author Sachin <sachin@...> 1726678200 +0530

---

# The Three Areas of Git

When you work on your computer, code moves through 3 distinct zones:

::code lang=
 Working Directory    ---- git add ---->       Staging (Index)
 (Files you edit)                             (Draft Snapshot)
        ^                                            |
        |                                       git commit
        |                                            |
        +------------- git checkout ---------------- v
                                              Git Repository (.git)
                                              (Permanent History)

- **Working Directory**: Your regular filesystem folder.

- **Staging Area (`.git/index`)**: A binary file holding the exact list of blobs ready for the next commit.

- **Repository**: Immutable historical commits linked in a graph.

::notes
Explain why the staging area exists: It lets you craft clean, atomic commits. You might edit 5 files to fix a bug and add a feature, but stage only 2 files for the bugfix commit.

---

# What is a Branch? (It's a Sticky Note!)

- Many beginners think a branch duplicates their entire project.

- In Git, a branch is literally a **41-byte plain text file**!

::code lang=bash
  # Look inside your local branch folder:
  cat .git/refs/heads/main
  # Output:
  # 96da3b24f5a89e4c8300dd08aeb37f8e728b8dad

- That's it! A branch is just a named pointer holding **one 40-character commit hash**.
- Creating a branch (`git branch feature`) doesn't copy files—it writes 41 bytes to disk.

::notes
This is why branching in Git is instantaneous (O(1)), whereas in SVN it used to take minutes to copy entire directories.

---

# Understanding HEAD and "Detached HEAD"

- **What is HEAD?**

- A text file in `.git/HEAD` that tracks where you currently are.

- Normally it points to a branch:

`cat .git/HEAD` $\to$ `ref: refs/heads/main`

- **What is "Detached HEAD"?**

- When you checkout a specific commit instead of a branch name:

`git checkout 7a3b4c`

- Now `HEAD` points directly to a commit hash instead of a branch pointer.

- **Is it broken?**

- No! You can inspect, compile, test, and run code safely.

- If you want to keep changes made in detached HEAD, just make a branch:

`git switch -c my-new-branch`

---

# Git as a Directed Acyclic Graph (DAG)

Commits form an immutable Directed Acyclic Graph pointing backwards in time:

::code lang=
  (C1) <--- (C2) <--- (C3) <--- [main]
               ^
                \
                 (C4) <--- (C5) <--- [feature]

- **Fast-Forward Merge**:

- If `main` has no new commits, merging `feature` simply moves the `[main]` pointer to `(C5)`. Zero new objects created!

- **3-Way Merge**:

- If both `main` and `feature` progressed, Git creates a new merge commit `(C6)` with **two parent pointers**: `(C3)` and `(C5)`.

::notes
Explain the DAG acronym: Directed (arrows point to parent), Acyclic (you can never loop back to create a circular history), Graph (nodes and edges).

---

# Git vs GitHub: Don't Confuse Them

- **Git** (The Engine):

- Created by Linus Torvalds in 2005 for the Linux kernel.

- A local, command-line distributed version control system.

- Works 100% offline with zero internet connection.

- **GitHub** (The Cloud Host & Social Layer):

- A cloud platform (owned by Microsoft) that hosts remote bare Git repositories.

- Adds collaboration features on top:

- Pull Requests (code review workflow)

- GitHub Actions (CI/CD automated testing)

- Issues, Discussions, Project boards

- Alternatives: GitLab, Bitbucket, Gitea, Sourcehut.

::notes
College students often think Git == GitHub. Emphasize that Git was around before GitHub and that they can push to any server, local network, or USB stick.

---

# 5 Golden Rules for Every CS Student

- **1. Make Atomic Commits**:

- One logical change per commit. Don't bundle 10 unrelated fixes into "updated files".

- **2. Write Useful Commit Messages**:

- Good: `feat(auth): add JWT expiration refresh check`

- Bad: `fixed bug`, `asdf`, `wip`

- **3. Never Force-Push to Shared Branches**:

- `git push --force` rewrites history and breaks your teammates' local branches.

- **4. Master `.gitignore`**:

- Never commit build artifacts (`node_modules/`, `*.exe`, `.env`, binary outputs).

- **5. Inspect When in Doubt**:

- `git status` tells you where you are.

- `git log --graph --oneline --all` visualizes your DAG in the terminal!

---

::align center
# Master the Graph, Master Git

### Stop memorizing commands. Think in pointers and snapshots.

Presented live in the terminal using **Termdeck**

Press `q` to exit
