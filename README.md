**Wrap middleware handlers**

*Chain Wrappers for http Handlers or middle ware chaining*

* Chain creates an ordered chain of handlers from an argument list
* ChainLinkWrap wraps each handler in the argument list of handlers
  calls in a single env

* As an example wrapping middleware can be done with a recover wrapper
  [R] can wrap the chain of handlers

```
    // The handlers call chain A->B->C => R(A->B->C)
    handler := R(Chain(A,B,C))
```

* Or recover could wrap each of the individual handlers in the chain

```
    // The handlers call chain A->B->C => R(A)->R(B)->R(C)
    handler := ChainLinkWrap(R,A,B,C)
```

* Or recover could wrap one handler

```
    // The handlers call chain A->B->C =>  A->B->R(C)
    handler := Chain(A,B,R(C))
```

* Example buffered handlers using a buffer pool bytes.Buffer might be
  used like the following.

* Simple buffer

```
    handler := HttpScopedHandlerWriter(ChainLinkWrap(R,A,B,C))
```

* Buffer pool
```
    handler := HttpScopedBPHandlerWriter(ChainLinkWrap(R,A,B,C))
```

