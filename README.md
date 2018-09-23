# rvisu

rvisu is a redis topology visualizer. In other words, it is able to walk the topology graph for a given redis system, learning about nodes on the way and outputting the result in different formats.

## Example

Given the addresses of two redis sentinels, this finds all associated redis nodes - The master and it's slaves in this case.
```
rvisu -addr=10.121.201.240:26379,10.121.201.242:26379 | dot -Tpng > graph.png
```
![Example output](/examples/graph.png)

## Supported output formats

| Format                      | Description                                              |
| -------------               | -------------:                                           |
| -output=graphviz (default)  | https://www.graphviz.org Suitable for use by most tools. |
| -output=debug               | A debug format representing the internal structure.      |
