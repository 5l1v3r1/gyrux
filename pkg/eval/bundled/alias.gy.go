package bundled

const stdgy = `use re

dir = ~/.gyrux/aliases

arg-replacer = '{}'

aliases = [&]

fn -load-alias [name file]{
  nop $aliases
  -source $file
  -tmpfile = (mktemp)
  echo 'aliases['$name'] = $'$name'~' > $-tmpfile
  -source $-tmpfile
  rm -f $-tmpfile
}

fn def [&verbose=false &use=[] name @cmd]{
  file = $dir/$name.elv
  use-statements = [(each [m]{ put "use "$m";" } $use)]
  echo "#alias:new" $name (if (not-eq $use []) { put "&use="(to-string $use) }) $@cmd > $file
  args-at-end = '$@_args'
  new-cmd = [
    (each [e]{
        if (eq $e $arg-replacer) {
          put '$@_args'
          args-at-end = ''
        } else {
          put $e
        }
    } $cmd)
  ]
  echo 'fn '$name' [@_args]{' $@use-statements $@new-cmd $args-at-end '}' >> $file
  if (not-eq $verbose false) {
    echo (styled "Defining alias "$name green)
  }
  -load-alias $name $file
}

fn new [&verbose=false &use=[] @arg]{ def &verbose=$verbose &use=$use $@arg }

fn bash-alias [@args]{
  line = $@args
  name cmd = (splits &max=2 '=' $line)
  def $name $cmd
}

fn export {
  result = [&]
  keys $aliases | each [k]{
    result[$k"~"] = $aliases[$k]
  }
  put $result
}

fn list {
  _ = ?(grep -h '^#alias:new ' $dir/*.elv | sed 's/^#//')
}

fn ls { list } # Alias for list

fn undef [name]{
  file = $dir/$name.elv
  if ?(test -f $file) {
    # Remove the definition file
    rm $file
    echo (styled "Alias "$name" removed (will take effect on new sessions, or when you run 'del "$name"~')." green)
  } else {
    echo (styled "Alias "$name" does not exist." red)
  }
}

fn rm [@arg]{ undef $@arg }

fn init {
  if (not ?(test -d $dir)) {
    mkdir -p $dir
  }

  for file [(_ = ?(put $dir/*.elv))] {
    content = (cat $file | slurp)
    if (or (re:match '^#alias:def ' $content) (re:match '\nalias\[' $content)) {
      m = (re:find '^#alias:(def|new) (\S+)\s+(.*)\n' $content)[groups]
      new $m[2][text] $m[3][text]
    } elif (re:match '^#alias:new ' $content) {
      name = (re:find '^#alias:new (\S+)\s+(.*)\n' $content)[groups][1][text]
      -load-alias $name $file
    }
  }
}

init
`