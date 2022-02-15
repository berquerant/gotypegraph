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
  -deny.name string
        Deny objects whose name matches this.
  -deny.pkg string
        Deny packages whose name matches this.
  -foreign
        Search definitions in foreign packages.
  -private
        Search private definitions.
  -stat
        Generate stat graph when type is dot.
  -type string
        Output format. string, json or dot. (default "dot")
  -universe
        Search definitions in builtin packages.
```

## Example

Use graphviz.

``` shell
❯ gotypegraph ./... > /tmp/example.dot
❯ dot -Tsvg /tmp/example.dot -o img/example.svg
```

A gray region is a package.

An arrow is a dependency, the arrow's tail is the reference and the arrow's head is the definition.  
The arrow's label is the count of the same dependency (no label means 1).

A square region in a package is a definition, func, var, etc.  
`In` is the number of times it is referred by the other definitions.  
`Out` is the number of times it refers the other definitions.  
`UniqIn` is the unique `In`, `UniqOut` is the unique `Out`.  
`IO` is `In + Out`.

Emphasized arrows and definitions have many dependencies.

On mouseover, an arrow and a definition displays additional information,  
an arrow displays the tail and the head, a definition displays dependencies in lines.

Generate graph with `-stat`:

``` shell
❯ gotypegraph -stat ./... > /tmp/example_stat.dot
❯ dot -Tsvg /tmp/example_stat.dot -o img/example_stat.svg
```

The graph displays dependencies aggregated by package.
