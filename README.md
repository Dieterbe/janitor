# Janitor

Experimental work in progress, to make a program that cleans up a filesystem (e.g. /home), by
1) working on the level of directories and zip archives, rather than individual files.
2) focusing on interactivity.

## Working on the level of directories, not files

Tools like fdupes, fslint (deprecated) and rmlint are good at finding duplicate files and letting you clean each individiual duplicate or otherwise "linty" file.

However, I have often entire directories that can be removed, because an identical directory may exist elsewhere, or one that is a superset (e.g. has all the same files plus some more), or the directory is full of lint. I rather not consider every single file, but rather the directory as a whole, when possible.
Similarly, I may have zip files for which the extracted content is in a different folder somewhere, or even split over multiple directories. I want to find such stale zip files, get an overview of where their content lives, and clean (remove) them if i'm assured all their content is safely stored somewhere.

## Interactivity

I often don't even know in advance 
1) what kind of a mess I made
2) what kind of files I will encounter in a directory ( if the directory can't be treated as a whole). They may have different purposes or types (`$HOME/downloads` is a good example)

This means an interactive program where I can see previews, and do very quick 1-by-1 processing where needed (remembering common actions/preferences) seems better,
rather than a CLI tool or shell scripts where you have to configure many special arguments or cases and hope that it'll apply correctly.
If i have to review files by one by anyway, I rather do it once and act upon them immediately.

## Current status

The basic file/directory/archive fingerprinting and similarity computation works.
However, all the work to actually __act upon__ this data and actually clean your data, is not implemented yet.


### Implementation details

* Janitor scans filesystem paths. Scanning involves traversing all files and folders within these paths.
* All encountered Files are represented as "Prints" containing their size, content hash and basename. (note: ownership, permission bits, etc are ignored).
* All encountered directories are represented by a Print that includes all Prints of the files and directories contained inside of it, except the paths are adjusted to the full path within that directory.
* The same is true for all encountered zip files, which can be thought of as a "compressed directory".
* Every, subfolder and zip file in the scanpath is represented by a DirPrint, even subfolders within zip files. While the user can't remove directories inside of zip files, it seems useful information, though this can be changed.
* Similarity between DirPrints consists of 2 values:
  - content similarity: num_bytes_matching / (num_bytes_matching + num_bytes_non_matching)
  - meta similarity: average string similarity for matching content
* `__MACOSX` folders don't seem to have any use and are completely ignored. (this could be turned into a preference if needed)


Example.
Imagine a directory containing these files.  These would result in the following Prints.
Imagine a directory containing these files.  These would result in the following object representations
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


Note:

* `fs.WalkDir` uses lexical ordering, which makes things easier (consistent ordering of directory entries) and predictable (children are always walked after their parent).
This is true whether walking a real filesystem or a zip file.

