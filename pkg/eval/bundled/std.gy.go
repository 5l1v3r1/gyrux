package bundled

const stdgy = `use re

# Gyrux Standart Library (stdlib)
# Available:
#           std:pipesplit
#           std:eval
#           std:in
#           std:out
#           std:max
#           std:min
#           std:cond
#           std:partipal

fn pipesplit [l1 l2 l3]{
  pout = (pipe)
  perr = (pipe)
  run-parallel {
    $l1 > $pout 2> $perr
    pwclose $pout
    pwclose $perr
  } {
    $l2 < $pout
    prclose $pout
  } {
    $l3 < $perr
    prclose $perr
  }
}

fn out [text]{
  put $text
}

fn eval [str]{
  tmpf = (mktemp)
  echo $str > $tmpf
  -source $tmpf
  rm -f $tmpf
}

in~ = { print (head -n1) }

use builtin
if (has-key $builtin: read-upto~) {
  in~ = { print (read-upto "\n")[:-1] }
}

fn max [a @rest &with=[v]{put $v}]{
  res = $a
  val = ($with $a)
  each [n]{
    nval = ($with $n)
    if (> $nval $val) {
      res = $n
      val = $nval
    }
  } $rest
  print $res
}

fn min [a @rest &with=[v]{put $v}]{
  res = $a
  val = ($with $a)
  each [n]{
    nval = ($with $n)
    if (< $nval $val) {
      res = $n
      val = $nval
    }
  } $rest
  print $res
}

fn cond [clauses]{
  range &step=2 (count $clauses) | each [i]{
    exp = $clauses[$i]
    if (eq (kind-of $exp) fn) { exp = ($exp) }
    if $exp {
      print $clauses[(+ $i 1)]
      return
    }
  }
}

fn partial [f @p-args]{
  print [@args]{
    $f $@p-args $@args
  }
}
`
