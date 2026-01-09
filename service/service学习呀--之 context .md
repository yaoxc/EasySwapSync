# User: 这里定义的全局上下文ctx为什么不是指针类型？

GitHub Copilot: 在 Go 语言中，`context.Context` 本身就是一个接口类型，设计上是不可变的（immutable），且通常以值传递（非指针）方式使用。这样做的原因有：

1. `context.Context` 的实现是轻量级的接口，值传递不会带来性能损失。
2. 语义上，`Context` 只需要传递引用，不需要修改原始对象，所以不需要用指针。
3. 官方推荐用值类型传递 `context.Context`，避免指针带来的混淆和误用。

因此，`ctx context.Context` 用值类型是 Go 的最佳实践，也是官方推荐的用法。

# User: 值类型ctx 传递到子goroutine中，ctx复制为新的对象，怎么控制整个goroutine链的取消

GitHub Copilot: `context.Context` 作为接口类型，值传递时只是复制接口的引用，不会复制底层数据。无论传递到多少 goroutine，都是指向同一个 context 实例。

取消 context（如 `context.WithCancel`、`context.WithTimeout`）时，所有持有该 context 的 goroutine 都会收到取消信号。只要 goroutine 用的是同一个 context（或其派生子 context），就能正确响应取消操作。

总结：ctx 传递到子 goroutine 时不会复制为新对象，取消操作会同步通知所有相关 goroutine。

补充说明： context 传递到子 goroutine时的取消机制，澄清不会复制为新对象，所有 goroutine 都能同步响应取消。

示例：

```go
ctx, cancel := context.WithCancel(parentCtx)
go func(ctx context.Context) { <-ctx.Done(); ... }(ctx)
cancel() // 所有持有 ctx 的 goroutine 都会收到取消通知
```