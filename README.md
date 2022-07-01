# homedirclean

Hacky experimental program to clean your home directory, or any file system really.

Tools like fdupes, fslint (kindof outdated), rmlint are neat, but they deal with individual files.
I have often entire directories that are either exact copies, or one is a superset of the other (e.g. has all the same files plus some more)
So I want to work on the level of directories, not files.  Similarly, I may have zip files lingering around for which the extracted content is in a different folder somewhere, or even split over multiple directories. I want to find such stale zip files, get an overview of where their content lives, and clean them (remove them).

I often don't even know in advance just what kind of a mess I made, which means an interactive program where I can do previews or remember operations and preferences might be worthwhile, rather than a CLI which you have to frequently rerun with different arguments (or edit the resulting shellscript), can be great but also tedious.

https://qarmin.github.io/czkawka/ has a bunch of interesting features, but is also file oriented, and written in Rust, which I'm not good at. I'm productive in Go so that's what I choose for this project.

## How it works

* HDC scans filesystem paths. Scanning involves traversing all files and folders within these paths.
* All encountered Files are represented as "Prints" containing their content hash and basename. (note: ownership, permission bits, etc are ignored).
* All encountered directories are represented by a Print that includes all Prints of the files and directories contained inside of it, except the paths are adjusted to the full path within that directory.
* The same is true for all encountered zip files, which can be thought of as a "compressed directory".
* Every, subfolder and zip file in the scanpath is represented by a DirPrint, but not subfolders within zip files, since the user can't do deletes inside of zip files.  We only represent "deletable entries" as Prints // TBD. might make sense to show subfolders within zipfiles, to see zips which have been partially extracted already.
* Similarity between DirPrints consists of 2 values:
  - content similarity: num_bytes_matching / (num_bytes_matching + num_bytes_non_matching)
  - meta similarity: average string similarity for matching content
* `__MACOSX` folders don't seem to have any use and are completely ignored.


Example.
Imagine a directory containing these files.  These would result in the following Prints.
The entries are already shown in lexical order, how they will be walked by the code.
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


TODO: test that fingerprint of a file and a zip file are identical, regardless of order

## Implementation details

fs.WalkDir uses lexicographical ordering, which makes things easier (consistent ordering of directory entries) and predictable (children are always walked after their parent).
This is true whether walking a real filesystem or a zip file.

