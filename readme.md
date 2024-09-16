# optionalv2

`optionalv2` is a Go package that provides a generic `Option[T any]` type, inspired by the Option type in languages like Rust. It allows you to represent an optional value: every `Option` is either `Some(value)` or `None`. This can be particularly useful when you need to distinguish between a missing value and a zero value in your applications, especially when dealing with JSON serialization/deserialization.

## Features

- **Generic Option Type**: Supports any type `T`, thanks to Go's generics.
- **Explicit Null Values**: Ability to represent explicit `null` values when serializing to JSON.
- **JSON Marshalling/Unmarshalling**: Seamless integration with Go's `encoding/json` package.
- **Convenient Methods**: Provides a set of methods for working with optional values, such as `Unwrap()`, `IsSome()`, `IsNone()`, `TakeOr()`, etc.

## Installation

```bash
go get github.com/tapp-ai/go-optional-v2
```

## Usage

### Importing the Package

```go
import "github.com/tapp-ai/go-optional-v2"
```

### Creating Option Values

#### Some

To create an `Option[T]` with a value:

```go
opt := optionalv2.Some(42)
```

**Note**: If you pass the zero value of the type `T` to `Some`, it will be treated as an explicit `null` when marshalled to JSON. See the [Edge Cases and Special Behaviors](#edge-cases-and-special-behaviors) section for more details.

#### None

To create an `Option[T]` without a value:

```go
opt := optionalv2.None[int]()
```

### Checking if an Option Has a Value

```go
if opt.IsSome() {
    // Option has a value
} else if opt.IsNone() {
    // Option is None (no value)
}
```

### Unwrapping the Value

#### Unwrap

Retrieves the value inside the `Option`. If the `Option` is `None`, it returns the zero value of type `T`. Otherwise, it returns the value.

```go
value := opt.Unwrap()
```

#### UnwrapAsPtr

Retrieves the value as a pointer. If the `Option` is `None`, it returns `nil`. Otherwise, it returns a pointer to the value.

```go
valuePtr := opt.UnwrapAsPtr()
```

### Taking the Value with Error Handling

```go
value, err := opt.Take()
if err != nil {
    // Handle the error (e.g., Option is None)
}
```

### Fallback Values

#### TakeOr

Returns the value if present, otherwise returns the provided fallback value.

```go
value := opt.TakeOr(100) // Returns 100 if opt is None
```

#### TakeOrElse

Returns the value if present, otherwise executes the provided function and returns its result.

```go
value := opt.TakeOrElse(func() int {
    // Compute fallback value
    return 200
})
```

### Chaining Options

#### Or

Returns the current `Option` if it has a value, otherwise returns the provided fallback `Option`.

```go
opt := opt.Or(optionalv2.Some(300))
```

### Filtering Options

#### Filter

Returns the `Option` if it satisfies the predicate, otherwise returns `None`.

```go
opt = opt.Filter(func(v int) bool {
    return v > 10
})
```

### Conditional Execution

#### IfSome

Executes a function if the `Option` is `Some`.

```go
opt.IfSome(func(v int) {
    fmt.Println("Value is:", v)
})
```

#### IfNone

Executes a function if the `Option` is `None`.

```go
opt.IfNone(func() {
    fmt.Println("No value")
})
```

### String Representation

```go
fmt.Println(opt.String()) // Outputs: Some[42] or None[]
```

## JSON Marshalling/Unmarshalling

The `Option` type implements `json.Marshaler` and `json.Unmarshaler`, allowing it to be seamlessly serialized and deserialized using the standard `encoding/json` package.

**Important**: The `Option` type should always be used with an `omitempty` tag in struct fields to ensure correct behavior when marshalling to JSON. Marshalling an `Option` without `omitempty` may result in unexpected behavior.

TLDR: In this package, the JSON `null` is treated as the GoLang zero value (and vice versa). JSON absent fields are treated as `None`.

### Marshalling Behavior

- If the `Option` is `Some` and contains a non-zero value, it is marshalled as the value.
- If the `Option` is `Some` and contains the zero value of type `T`, it is marshalled as `null`.
- If the `Option` is `None`, it is omitted when marshalling (assuming `omitempty` is set in struct tags).

### Unmarshalling Behavior

- If the JSON field is absent, `UnmarshalJSON` is not called, and the `Option` remains `None`.
- If the JSON field is present and `null`, the `Option` becomes `Some` with the zero value of type `T` (representing an explicit `null`).
- If the JSON field has a value, the `Option` becomes `Some` with that value.

### Example

```go
type MyStruct struct {
    Name     optionalv2.Option[string] `json:"name,omitempty"`
    Age      optionalv2.Option[int]    `json:"age,omitempty"`
    Birthday optionalv2.Option[time.Time] `json:"birthday,omitempty"`
}
```

#### Marshalling

```go
s := MyStruct{
    Name: optionalv2.Some("Alice"),
    Age: optionalv2.None[int](),
    Birthday: optionalv2.Some(time.Time{}), // Zero value, will be marshalled as null
}

data, err := json.Marshal(s)
// data will be: { "name": "Alice", "birthday": null }
```

#### Unmarshalling

```go
jsonData := []byte(`{ "name": "Bob", "age": null }`)

var s MyStruct
err := json.Unmarshal(jsonData, &s)

// s.Name is Some("Bob")
// s.Age is Some(0) (explicit null)
// s.Birthday is None (field absent)
```

## Edge Cases and Special Behaviors

- **Zero Values**: When you pass the zero value of type `T` to `Some`, it is treated as an explicit `null` when marshalling to JSON. This allows you to distinguish between an absent field (`None`) and a field explicitly set to `null`.
- **Omitted Fields**: If an `Option` field in a struct is `None` and has the `omitempty` tag, it will be omitted from the JSON output.

## Examples

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/tapp-ai/go-optional-v2"
)

func main() {
    opt := optionalv2.Some(42)

    if opt.IsSome() {
        fmt.Println("Value is:", opt.Unwrap())
    } else {
        fmt.Println("No value")
    }

    optNone := optionalv2.None[int]()
    fmt.Println("Has value?", optNone.IsSome())
}
```

### Working with JSON

```go
package main

import (
    "encoding/json"
    "fmt"
    "github.com/tapp-ai/go-optional-v2"
)

type User struct {
    Name optionalv2.Option[string] `json:"name,omitempty"`
    Age  optionalv2.Option[int]    `json:"age,omitempty"`
}

func main() {
    user := User{
        Name: optionalv2.Some("Charlie"),
        Age:  optionalv2.Some(0), // Explicit null when marshalled
    }

    data, _ := json.Marshal(user)
    fmt.Println(string(data)) // Output: { "name": "Charlie", "age": null }
}
```

### Conditional Execution

```go
opt := optionalv2.Some(10)

opt.IfSome(func(v int) {
    fmt.Println("Value is:", v)
})

opt.IfNone(func() {
    fmt.Println("Option is None")
})
```

## Maintainers

This package is maintained by the engineering team @ [StyleAI](https://usestyle.ai/).

## License

[MIT License](LICENSE)
