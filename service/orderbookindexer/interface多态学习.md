
# 给一个接口多态的例子，给出demo代码

````go
// ...existing code...

// 定义一个接口
type Notifier interface {
    Notify(msg string)
}

// 实现接口的结构体A
type EmailNotifier struct {
    Email string
}

func (e *EmailNotifier) Notify(msg string) {
    fmt.Printf("Send Email to %s: %s\n", e.Email, msg)
}

// 实现接口的结构体B
type SMSNotifier struct {
    Phone string
}

func (s *SMSNotifier) Notify(msg string) {
    fmt.Printf("Send SMS to %s: %s\n", s.Phone, msg)
}

// 使用接口实现多态
func SendNotification(n Notifier, msg string) {
    n.Notify(msg)
}

func main() {
    email := &EmailNotifier{Email: "user@example.com"}
    sms := &SMSNotifier{Phone: "123456789"}

    SendNotification(email, "订单已创建")
    SendNotification(sms, "订单已创建")
}

// ...existing code...
````

**说明：**
- `Notifier` 是接口，`EmailNotifier` 和 `SMSNotifier` 都实现了该接口。
- `SendNotification` 函数参数为接口类型，可以接收任意实现了该接口的对象，实现多态调用。


# SendNotification的参数为Notifier对象，为什么在main函数使用传参的时候给的是指针？

因为 `Notifier` 接口的方法（如 `Notify`）是在指针接收者（如 `*EmailNotifier`、`*SMSNotifier`）上实现的。  
Go 只有当传递指针类型时，才认为这些类型实现了接口。如果传递结构体值（如 `EmailNotifier`），则不会满足接口实现，编译会报错。

简言之：  
**接口的方法集与接收者类型有关，指针接收者实现的接口，必须用指针传递。**

# 那如果EmailNotifier实现了一个传递结构体值的方法呢？

如果 `EmailNotifier` 实现了用结构体值（非指针）作为接收者的方法，例如：

````go
func (e EmailNotifier) Notify(msg string) {
    fmt.Printf("Send Email to %s: %s\n", e.Email, msg)
}
````

那么，无论你传递 `EmailNotifier` 还是 `*EmailNotifier`，都可以赋值给 `Notifier` 接口参数：

- 传递结构体值：`SendNotification(email, "msg")`，其中 `email := EmailNotifier{...}`
- 传递指针：`SendNotification(&email, "msg")`

Go 会自动根据方法集规则，将指针和结构体值都视为实现了接口。

**总结：**
- 如果方法接收者是结构体值，可以用值或指针赋值给接口。
- 如果方法接收者是指针，只能用指针赋值给接口。

# 给出这两种情况的详细代码

````go
// ...existing code...

import "fmt"

// 定义接口
type Notifier interface {
    Notify(msg string)
}

// 情况1：方法接收者为指针
type EmailNotifier struct {
    Email string
}

func (e *EmailNotifier) Notify(msg string) {
    fmt.Printf("指针接收者: Send Email to %s: %s\n", e.Email, msg)
}

// 情况2：方法接收者为结构体值
type SMSNotifier struct {
    Phone string
}

func (s SMSNotifier) Notify(msg string) {
    fmt.Printf("值接收者: Send SMS to %s: %s\n", s.Phone, msg)
}

// 多态调用
func SendNotification(n Notifier, msg string) {
    n.Notify(msg)
}

func main() {
    // 指针接收者，只能传指针
    email := &EmailNotifier{Email: "user@example.com"}
    SendNotification(email, "订单已创建") // 正确

    // 下面这行如果取消注释会编译报错，因为EmailNotifier值没有实现接口
    // emailVal := EmailNotifier{Email: "user@example.com"}
    // SendNotification(emailVal, "订单已创建") // 编译错误

    // 值接收者，既可以传值也可以传指针
    sms := SMSNotifier{Phone: "123456789"}
    SendNotification(sms, "订单已创建") // 正确

    smsPtr := &SMSNotifier{Phone: "987654321"}
    SendNotification(smsPtr, "订单已创建") // 也正确
}

// ...existing code...
````

**说明：**
- 指针接收者（`*EmailNotifier`）只能用指针赋值给接口。
- 值接收者（`SMSNotifier`）既可以用结构体值，也可以用指针赋值给接口。


# 值接收者（SMSNotifier）既可以用`结构体值`，也可以用`指针`赋值给接口 的底层原理是什么？

Go 的方法集规则决定了这一行为：

- **对于结构体类型（如 SMSNotifier）**，其方法集包含所有用值接收者和指针接收者定义的方法。
- **对于结构体指针类型（如 *SMSNotifier）**，其方法集也包含所有用值接收者和指针接收者定义的方法。

但**只有值接收者方法**时，`结构体值`和`结构体指针`都能调用这些方法，也都实现了接口。这是因为 Go 编译器会自动帮你做“取址”操作（即 `(&s).Method()`），所以指针和值都能满足接口。

**底层原理：**
- 当接口需要的方法是值接收者实现的，Go 允许用结构体值或指针赋值给接口变量。
- 当接口需要的方法是指针接收者实现的，只能用指针赋值给接口变量。

**简要总结：**  
值接收者方法既属于类型本身，也属于其指针类型的方法集，因此两者都能赋值给接口。