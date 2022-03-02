# gotypegraph

Generate definitions and references graph.

## Usage

```
❯ gotypegraph -h
Usage of gotypegraph:
  gotypegraph [flags] -type TYPE patterns...
Flags:
  -accept.name string
        Accept objects whose name matches this.
  -accept.pkg string
        Accept packages whose name matches this.
  -buffer int
        Size of search buffers. (default 1000)
  -deny.name string
        Deny objects whose name matches this.
  -deny.pkg string
        Deny packages whose name matches this.
  -fontsize,max int
        Max fontsize used for text in dot. (default 24)
  -fontsize.min int
        Min fontsize used for text in dot. (default 8)
  -foreign
        Search definitions in foreign packages.
  -log.regexp string
        Regexp to grep logs.
  -noselfloop
        Ignore self references.
  -penwidth.max int
        Max penwidth used to draw lines in dot. (default 1)
  -penwidth.min int
        Min penwidth used to draw lines in dot. (default 1)
  -private
        Search private definitions.
  -stat
        Generate stat graph when type is dot.
  -type string
        Output format. json or dot. (default "dot")
  -universe
        Search definitions in builtin packages.
  -v string
        Logging verbosity. error, warn, info, debug or verbose. (default "info")
  -weight.max int
        Max weight for dot. (default 100)
  -weight.min int
        Min weight for dot. (default 1)
  -worker int
        Number of search workers. (default 4)
```

## Example

Use graphviz.

``` shell
❯ gotypegraph ./... > /tmp/example.dot
❯ dot -Tsvg /tmp/example.dot -o /tmp/example.svg
```

A gray region is a package.

An arrow is a dependency, the arrow's tail is the reference and the arrow's head is the definition.  
The arrow's label is the count of the same dependency (no label means 1).

A square region in a package is a definition, func, var, etc.  
`Ref` is the number of times it refers the other definitions.  
`Def` is the number of times it is referred by the other definitions.  
`UniqRef` is the unique `Ref`, `UniqRef` is the unique `Def`.  
`RefDef` is `Ref + Def`.

Emphasized arrows and definitions have many dependencies.

On mouseover, an arrow and a definition displays additional information,  
an arrow displays the tail and the head, a definition displays dependencies in lines.

Generate graph with `-stat`:

``` shell
❯ gotypegraph -stat ./... > /tmp/example_stat.dot
❯ dot -Tsvg /tmp/example_stat.dot -o /tmp/example_stat.svg
```

The graph displays dependencies aggregated by package.
