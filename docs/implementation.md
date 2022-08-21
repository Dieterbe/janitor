# Implementation details

* Janitor scans filesystem paths. Scanning involves traversing all files and folders within these paths.
* All encountered Files are represented as "Prints" containing their size, content hash and basename. (note: ownership, permission bits, etc are ignored).
* All encountered directories are represented by a Print that includes all Prints of the files and directories contained inside of it, except the paths are adjusted to the full path within that directory (upon iteration)
* The same is true for all encountered zip files, which can be thought of as a "compressed directory".
* Every subfolder and zip file in the scanpath is represented by a DirPrint, even subfolders within zip files. While the user can't remove directories inside of zip files, it seems useful information, though this can be changed.
* Similarity between DirPrints consists of 2 values:
  - content similarity: `num_bytes_matching / (num_bytes_matching + num_bytes_non_matching)`
  - path similarity: average string similarity of path/filenames for matching content.
* `__MACOSX` folders don't seem to have any use and are completely ignored. (this could be turned into a preference if needed)


## Example
Imagine a directory containing these files.  These would result in the following Prints.
Note how the Prints for directories are simply the superset of all the Prints of the directories and filed contained inside them.

```
# File                       # Prints generated. FilePrints are {Path, Hash}.

a (directory)                [{1,0xa0}, {2, 0x3f}]
a/1                          {1, 0xa0}
a/2                          {2, 0x3f
b (directory)                [{1/x, 0xbb}, {1/y, 0xaf}, {2, 0xff}]
b/1 (directory)              [{x, 0xbb}, {y, 0xaf}]
b/1/x                        {x, 0xbb}
b/1/y                        {y, 0xaf}
b/2                          {2, 0xff}
b/3 (empty directory)        []
c (empty directory)          []
```
The root directory itself would be
```
[{a/1,0xa0}, {a/2, 0x3f}, {b/1/x, 0xbb}, {b/1/y, 0xaf}, {b/2, 0xff}]
```

Note that we don't redundantly store the FilePrints even though they seem repeated across their path in the hierarchy
DirPrints are dynamically generated.


### Note:

* `fs.WalkDir` uses lexical ordering, which makes things easier (consistent ordering of directory entries) and predictable (children are always walked after their parent).
This is true whether walking a real filesystem or a zip file.
* always log to the provided `log` file descriptor, never to stdout/stderr, as it messes with the TUI.
* if an error happens while walking a directory, that directory is omitted, but its parent (and other children) are still processed.  In a future version, we should also omit all parents (and grandparents) of the failing directory - this includes the root walking dir - as to only leave directories that have comprehensive (fully accurate) dirPrints. Since a directory's dirprint relies on accuracy of the dirprint of all its children.  For now, keep this into account: when errors happen, they will be logged, and take similarity reports for (grand)parents with a grain of salt.
