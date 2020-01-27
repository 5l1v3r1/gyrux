package bundled

const dirgy = `use re

use builtin

before-chooser = []
after-chooser = []

max-stack-size = 100

-dirstack = [ $pwd ]
-cursor = (- (count $-dirstack) 1)

fn -trimstack {
  -dirstack = $-dirstack[0:(+ $-cursor 1)]
}

fn stack { print $@-dirstack }

fn size { count $-dirstack }

fn dhist {
  for index [(range 0 (size))] {
    if (== $index $-cursor) {
      echo (styled "* "$-dirstack[$index] green)
    } else {
      echo "  "$-dirstack[$index]
    }
  }
}

fn getcwd {
  if (> (size) 0) {
    print $-dirstack[$-cursor]
  } else {
    print ""
  }
}

fn push {
  if (or (== (size) 0) (!=s $pwd (getcwd))) {
    -dirstack = [ (explode $-dirstack[0:(+ $-cursor 1)]) $pwd ]
    if (> (size) $max-stack-size) {
      -dirstack = $-dirstack[(- $max-stack-size):]
    }
    -cursor = (- (size) 1)
  }
}

fn back {
  if (> $-cursor 0) {
    -cursor = (- $-cursor 1)
    builtin:cd $-dirstack[$-cursor]
  } else {
    echo "Beginning of directory history!" > /dev/tty
  }
}

fn forward {
  if (< $-cursor (- (size) 1)) {
    -cursor = (+ $-cursor 1)
    builtin:cd $-dirstack[$-cursor]
  } else {
    echo "End of directory history!" > /dev/tty
  }
}

fn pop {
  if (> $-cursor 0) {
    back
    -trimstack
  } else {
    echo "No previous directory to pop!" > /dev/tty
  }
}

fn chdir [@dir]{
  if (and (== (count $dir) 1) (eq $dir[0] "-")) {
    builtin:cd $-dirstack[(- $-cursor 1)]
  } else {
    builtin:cd $@dir
  }
}

fn cdb [p]{ cd (dirname $p) }

fn left-word-or-prev-dir {
  if (> (count $edit:current-command) 0) {
    edit:move-dot-left-word
  } else {
    back
  }
}

fn right-word-or-next-dir {
  if (> (count $edit:current-command) 0) {
    edit:move-dot-right-word
  } else {
    forward
  }
}

fn left-small-word-or-prev-dir {
  if (> (count $edit:current-command) 0) {
    edit:move-dot-left-small-word
  } else {
    back
  }
}

fn right-small-word-or-next-dir {
  if (> (count $edit:current-command) 0) {
    edit:move-dot-right-small-word
  } else {
    forward
  }
}

fn chis {
  for hook $before-chooser { $hook }
  index = 0
  candidates = [(each [arg]{
        print [
          &to-accept=$arg
          &to-show=$index" "$arg
          &to-filter=$index" "$arg
        ]
        index = (to-string (+ $index 1))
  } $-dirstack)]
  edit:listing:start-custom $candidates &caption="Dir history " &accept=[arg]{
    builtin:cd $arg
    for hook $after-chooser { $hook }
  }
}

fn init {
  after-chdir = [ $@after-chdir [dir]{ push } ]
}

init
`
