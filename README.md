# Janitor

Experimental work in progress, to make a program that cleans up a filesystem (e.g. /home), by
1) working primarily on the level of directories and zip archives, rather than just individual files. (in progress)
2) focusing on interactivity: fast iteration cycles of previewing files/directories and making decisions quickly. (TODO)

## Working on the level of directories, not files

Tools like fdupes, fslint (deprecated) and rmlint are good at finding duplicate files and letting you clean each individiual duplicate or otherwise "linty" file.

However, I have often entire directories that can be removed, because an identical directory may exist elsewhere, or one that is a superset (e.g. has all the same files plus some more), or the directory is full of lint. I rather not consider every single file, but rather the directory as a whole, when possible.
Similarly, I may have zip files for which the extracted content is in a different folder somewhere, or even split over multiple directories. I want to find such stale zip files, get an overview of where their content lives, and clean (remove) them if i'm assured all their content is safely stored somewhere.

## Interactivity

I often don't even know in advance 
1) what kind of a mess I made
2) what kind of files I will encounter in a directory (if the directory can't be treated as a whole). They may have different purposes or types (`$HOME/downloads` is a good example)

This means an interactive program where I can see previews, and do very quick 1-by-1 processing where needed (remembering common actions/preferences) seems better,
rather than a CLI tool or shell scripts where you have to configure many special arguments or cases and hope that it'll apply correctly.
If i have to review files by one by anyway, I rather do it once and act upon them immediately.

## Current status

The basic file/directory/archive fingerprinting and similarity computation works.
However, all the work to actually __act upon__ this data and actually clean your data, is not implemented yet.
A cross-platform UI toolkit for Go would be ideal (to get good visual previews of data and have most extensive keybind options),
but in absence of that, [charm.sh/bubbletea](https://github.com/charmbracelet/bubbletea) looks like a neat TUI framework.

## Implementation details

see [./implementation.md](implementation details)


no symlink support at all. you can create loops or cause the same files to be rescanned if symlinks point to already scanned data, which may lead to false positives!
