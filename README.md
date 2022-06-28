# homedirclean

Hacky experimental program to clean your home directory, or any file system really.

Tools like fdupes, fslint (kindof outdated), rmlint are neat, but they deal with individual files.
I have often entire directories that are either exact copies, or one is a superset of the other (e.g. has all the same files plus some more)
So I want to work on the level of directories, not files.  Similarly, I may have zip files lingering around for which the extracted content is in a different folder somewhere, or even split over multiple directories. I want to find such stale zip files, get an overview of where their content lives, and clean them (remove them). And I often run into idiosyncracies like `__MACOSX` folders which don't seem to have any use and can be ignored or removed.

I often don't even know in advance just what kind of a mess I made, which means an interactive program where I can do previews or remember operations and preferences might be worthwhile, rather than a CLI which you have to frequently rerun with different arguments (or edit the resulting shellscript), which is what I'm usually a fan of.  

https://qarmin.github.io/czkawka/ has a bunch of interesting features, but is also file oriented, and written in Rust, which I'm not good at. I'm productive in Go so that's what I choose.

