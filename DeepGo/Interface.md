## Что такое интерфейс

Интерфейс в Go — это набор методов (method set), который описывает поведение, а не структуру данных. Если тип реализует все методы интерфейса — он автоматически считается его реализацией. Никаких `implements` как в Java не надо, это и называется **утиной типизацией** (duck typing): "если оно крякает как утка — значит это утка".

Zero value для интерфейса — это `nil`. Причём именно "полный" nil (когда и type, и value внутри пустые) — ниже разберём подводный камень с этим отдельно, там не всё так просто.

До версии 1.18 все типы интерфейсов можно было использовать как обычные типы значений (переменные, поля структур, аргументы функций). Начиная с 1.18 (введение generics) появились интерфейсы, которые можно использовать **только** как constraints для type parameters — например интерфейсы с union types типа `~int | ~int64`. Такие интерфейсы нельзя объявить как тип переменной.

```go
type Number interface {
	~int | ~int64 | ~float64
}

// var n Number = 5 // так нельзя, это constraint-only interface

func Sum[T Number](values []T) T {
	var result T
	for _, v := range values {
		result += v
	}
	return result
}
```

## Полиморфизм и утиная типизация

Полиморфизм через интерфейсы в Go работает так: одна и та же переменная интерфейсного типа может в рантайме содержать разные конкретные типы, и вызов метода приведёт к разному поведению в зависимости от того, что там реально лежит.

Утиная типизация — это как раз механизм, который делает этот полиморфизм неявным: тебе не нужно явно указывать, что тип реализует интерфейс, компилятор сам это проверяет по набору методов.

## Статический и динамический тип

У переменной интерфейсного типа есть два "типа":

- **Статический тип** — это тип самой переменной интерфейса, известный на этапе компиляции (например, `io.Reader`).
- **Динамический тип** — это конкретный тип, который реально хранится внутри интерфейса в рантайме (например, `*os.File`).

Это ключевая штука для понимания type assertion и dispatch — см. ниже.

## Свойства хороших интерфейсов (по мнению community)

- **Минималистичность** — не стоит пихать в интерфейс кучу методов "на всякий случай". Чем меньше методов, тем проще реализовать интерфейс и тем он более переиспользуемый. Классика — `io.Reader` с одним методом `Read`.
- **Независимость от реализации** — интерфейс не должен ничего знать о конкретных типах, которые его реализуют. Он просто описывает контракт.

**Правило:** возвращайте из функций конкретные типы, а не интерфейсы. Это даёт вызывающему коду доступ ко всем методам и полям конкретного типа, а не только к тому, что описано в интерфейсе. Абстракцию (интерфейс) лучше делать на стороне потребителя (см. раздел про расположение интерфейсов ниже).

## Пустой интерфейс (`interface{}` / `any`)

Пустого интерфейса стоит избегать, потому что мы теряем главное преимущество Go как статически типизированного языка — проверку типов на этапе компиляции. Как только что-то стало `interface{}`, компилятор больше не может подсказать тебе, что ты сделал не так с типами — все ошибки вылезут только в рантайме (см. пример с паникой ниже).

Но есть законные случаи, когда без него не обойтись:

- `QueryContext` (или похожие функции, которые принимают произвольные аргументы `...interface{}`)
- `json.Marshal` / `Unmarshal` — когда структура данных заранее неизвестна

### Как узнать реальный тип за пустым интерфейсом? — Type Assertion

```go
package main

import "fmt"

func main() {
	var value int = 100
	var i interface{} = value

	// "safe" форма - с ok, паники не будет
	converted1, ok1 := i.(int)
	if ok1 {
		fmt.Println("converted1 int:", converted1)
	}

	converted2, ok2 := i.(float32)
	if ok2 {
		fmt.Println("converted2 float32:", converted2)
	}

	// "unsafe" форма - без ok
	converted3 := i.(int)
	fmt.Println("converted3 int:", converted3)

	converted4 := i.(float32) // 💥 паника, i хранит int, а не float32
	fmt.Println("converted4 float32:", converted4)
}
```

Вывод:

```
converted1 int: 100
converted3 int: 100
panic: interface conversion: interface {} is int, not float32
```

⚠️ **Подводный камень**: если не использовать форму с `ok`, при несовпадении типа программа паникует. Почти всегда лучше писать `v, ok := i.(T)`, а не `v := i.(T)`, если ты не на 100% уверен в типе.

### Как на самом деле работает `x.(T)`

`x.(T)` проверяет, что `x != nil`, и:

- если `T` **не интерфейс** — проверяет, что динамический тип `x` совпадает с `T`
- если `T` **интерфейс** — проверяет, что динамический тип `x` реализует `T`

```go
package main

import "fmt"

type fooer interface{ foo() }
type barer interface{ bar() }
type foobarer interface {
	foo()
	bar()
}

type thing struct{}

func (t *thing) foo() {}
func (t *thing) bar() {}

func main() {
	var i foobarer = &thing{}
	_, ok := i.(fooer) // проверяем, реализует ли динамический тип i интерфейс fooer
	fmt.Println("result:", ok)
}
```

Вывод:

```
result: true
```

А вот пример, который наглядно показывает, что проверяется именно "динамический тип реализует интерфейс" — причём интерфейс можно вообще объявить анонимно, прямо на месте assertion:

```go
package main

import "fmt"

type fooer interface{ foo() }

type thing struct{}

func (t *thing) foo() {}
func (t *thing) bar() {}

func main() {
	var i fooer = &thing{}

	// динамически проверяем наличие метода bar(),
	// хотя fooer про bar() ничего не знает
	_, ok := i.(interface{ bar() })
	fmt.Println("result:", ok)
}
```

Это удобно, когда нужно проверить наличие опционального метода (классический пример — проверка, реализует ли `io.Writer` ещё и `io.Closer`, без объявления отдельного именованного интерфейса).

## Type Switch

Когда вариантов конкретных типов много, вместо цепочки `if`-ов с `type assertion` удобнее использовать `type switch`:

```go
func describe(i interface{}) string {
	switch v := i.(type) {
	case int:
		return fmt.Sprintf("int: %d", v)
	case string:
		return fmt.Sprintf("string: %s", v)
	case nil:
		return "nil value"
	default:
		return fmt.Sprintf("unknown type: %T", v)
	}
}
```

Плюс: читается линейно и понятно. Минус: если забыл `default`, компилятор не подскажет — надо самому не забывать покрывать неизвестные случаи.

## Как жили без generics

До появления generics (1.18) полиморфные структуры данных писали через интерфейсы. Например бинарное дерево поиска:

```go
type BSTItem interface {
	Less(BSTItem) bool
}

type BSTNode struct {
	item  BSTItem
	left  *BSTNode
	right *BSTNode
}
```

Минус такого подхода — каждый `Less` вызов это динамическая диспетчеризация (см. ниже про перформанс), плюс приходится делать type assertion, если нужен доступ к полям конкретного типа. Generics это решают: код становится и типобезопасным, и быстрым (никакого dispatch через interface).

## Интерфейс под капотом

Это то самое место, где становится понятно, почему интерфейсы ведут себя так, а не иначе.

### Интерфейс с методами (`iface`)

```go
type iface struct {
	tab  *itab
	data unsafe.Pointer
}

type itab struct {
	inter *interfacetype
	_type *_type
	hash  uint32
	_     [4]byte
	fun   [1]uintptr // указатели на методы конкретного типа
}
```

То есть интерфейс — это на самом деле пара из двух указателей: **itab** (таблица методов + информация о типе) и **data** (указатель на реальные данные).

### ITABLE

`itab` уникальна для **каждой конкретной пары (интерфейс, статический тип)**. Просчитывать все возможные пары на этапе компиляции (early binding) было бы нерационально — комбинаций слишком много, и большинство из них никогда не будет использовано.

ℹ️ Вместо этого компилятор генерирует метаданные для каждого статического типа (список его методов) и отдельно метаданные для каждого интерфейса (список требуемых методов). А сама `itable` для конкретной пары вычисляется **во время выполнения программы** (late binding), причём результат кешируется — то есть просчёт для одной и той же пары происходит только один раз.

### Пустой интерфейс (`eface`)

```go
type eface struct {
	_type *_type
	data  unsafe.Pointer
}
```

У пустого интерфейса нет `itab`, потому что нет методов, которые нужно диспетчеризовать — только тип и данные.

### Опасный трюк с `unsafe` (работает до 1.23)

⚠️ Это НЕ то, что стоит делать в проде — но отлично иллюстрирует внутреннее устройство `eface`:

```go
package main

import (
	"fmt"
	"unsafe"
)

type eface struct {
	typ unsafe.Pointer
	val unsafe.Pointer
}

func main() {
	var value int = 100
	var i interface{} = value
	fmt.Println("before:", i)

	obj := (*eface)(unsafe.Pointer(&i))
	*(*int)(obj.val) = 200 // напрямую лезем в память интерфейса
	fmt.Println("after:", i)
	fmt.Println("after value:", value)
}
```

Вывод:

```
before: 100
after: 200
after value: 200
```

Интересно, что меняется даже `value` — потому что тут за интерфейсом на самом деле спрятан указатель на копию значения (детали см. ниже, в разделе про аллокации/copy semantics — в общем случае поведение зависит от того, попало значение в кучу или нет).

## Статическая vs Динамическая диспетчеризация

**Статическая диспетчеризация** — когда тип экземпляра, у которого вызывается метод, известен на этапе компиляции. Компилятор точно знает, какую функцию звать, и может её даже заинлайнить.

**Динамическая диспетчеризация** — когда тип экземпляра неизвестен заранее (мы работаем через интерфейс), и вызов метода нужно диспетчеризовать через `itab.fun` в рантайме.

```go
package main

import "fmt"

type Square struct{}

func (s *Square) Area() float64      { return 0.0 }
func (s *Square) Perimeter() float64 { return 0.0 }

const (
	AreaMethod = iota
	PerimeterMethod
)

// Interface тут — просто наглядная симуляция того, как
// выглядит itable "изнутри": массив указателей на методы
type Interface struct {
	methods [2]func() float64
}

func NewInterface(square *Square) Interface {
	return Interface{
		methods: [2]func() float64{
			AreaMethod:      square.Area,
			PerimeterMethod: square.Perimeter,
		},
	}
}

func (i *Interface) Area() float64      { return i.methods[AreaMethod]() }
func (i *Interface) Perimeter() float64 { return i.methods[PerimeterMethod]() }

func main() {
	// static dispatch — компилятор точно знает тип square
	square := &Square{}
	fmt.Println(square.Area())
	fmt.Println(square.Perimeter())

	// dynamic dispatch — вызов идёт через массив указателей,
	// как это происходит через itab.fun под капотом
	iface := NewInterface(square)
	fmt.Println(iface.Area())
	fmt.Println(iface.Perimeter())
}
```

### Бенчмарк: насколько дороже dynamic dispatch

```go
package main

import "testing"

// go test -bench=. -benchmem performance_test.go

type ValueInterface interface {
	Add() int
}

type Value struct{ number int }

//go:noinline
func (v Value) Add() int { return v.number + v.number }

type Value2 struct{ number int }

//go:noinline
func (v Value2) Add() int { return v.number + v.number }

var Result int

func Get(index int) ValueInterface {
	if index == 0 {
		return Value{}
	}
	return Value2{}
}

func BenchmarkDirect(b *testing.B) {
	var value Value
	for i := 0; i < b.N; i++ {
		Result = value.Add() // static dispatch
	}
}

func BenchmarkWithInterface1(b *testing.B) {
	var iface ValueInterface = Get(0)
	for i := 0; i < b.N; i++ {
		Result = iface.Add() // dynamic dispatch
	}
}

func BenchmarkWithInterface2(b *testing.B) {
	var iface ValueInterface = Get(1)
	for i := 0; i < b.N; i++ {
		Result = iface.(Value2).Add() // type assertion + static dispatch
	}
}
```

💡 Смысл бенчмарка: `BenchmarkDirect` быстрее всех, потому что компилятор может заинлайнить вызов. `BenchmarkWithInterface1/2` медленнее из-за косвенного вызова через `itab.fun` — процессору сложнее предсказать переход (branch prediction промахивается), плюс инлайнинг невозможен. На практике разница обычно небольшая, но в hot path это может быть заметно.

## Does interface always equal nil? — главный gotcha про интерфейсы

```go
package main

import "fmt"

type MyError struct{}

func (e *MyError) Error() string { return "error" }

func main() {
	var pointer *MyError = nil
	var err error = pointer
	fmt.Println("nil:", err == nil)
}
```

Вывод:

```
nil: false
```

⚠️ **Это САМЫЙ известный подводный камень с интерфейсами в Go.** Интерфейс — это пара `(itab, data)`. Когда мы кладём `nil`-указатель `*MyError` в интерфейс `error`:

- `data` — действительно `nil` (указатель на данные пустой)
- но `itab` — **не nil**! Компилятор уже знает, что в интерфейсе лежит именно `*MyError`, и `itab` для пары `(error, *MyError)` вполне себе существует

А интерфейс равен `nil` только тогда, когда **оба** поля — и `itab`, и `data` — пустые. Поэтому `err == nil` возвращает `false`, хотя внутри как бы "пустой" указатель.

**Практический вывод:** никогда не возвращайте типизированный nil-указатель как `error` из функции:

```go
// ПЛОХО
func doSomething() error {
	var err *MyError // nil, но типизированный
	if somethingBad {
		err = &MyError{}
	}
	return err // err == nil снаружи будет false, даже если somethingBad не случилось!
}

// ХОРОШО
func doSomething() error {
	if somethingBad {
		return &MyError{}
	}
	return nil // явный untyped nil
}
```

## Embedding интерфейсов

Интерфейсы можно встраивать друг в друга — это просто объединяет их method set:

```go
type Reader interface {
	Read(p []byte) (n int, err error)
}

type Writer interface {
	Write(p []byte) (n int, err error)
}

type ReadWriter interface {
	Reader
	Writer
}
```

Можно встраивать и структуры в интерфейс (в объявлении constraint-типа для generics) — но это уже не "классический" интерфейс, а именно constraint для type parameters.

### Можно ли приравнивать разные интерфейсы с одинаковыми методами?

Да, но только **если оба интерфейса объявлены в одном пакете**:

```go
package main

type A interface {
	Do()
}

type B interface {
	Do()
}

func main() {
	var a A
	var b B

	// Не сработает, если A и B объявлены
	// в разных пакетах!

	a = b
	b = a
}
```

Это работает благодаря **structural typing** — Go сравнивает не имя типа, а его "форму" (набор методов). Но с разными пакетами компилятор становится строже к типам (даже если method set идентичен) — так что не полагайся на это между пакетами.

## Интерфейс — это не всегда аллокация в куче

Частый миф: "как только я спрятал значение за интерфейсом — оно улетело в кучу (heap escape)". Это не всегда так! Компилятор может доказать через escape analysis, что значение не "убегает" за пределы функции, и оставить его на стеке.

```go
package main

// go build -gcflags '-l -m'

func printValue(v interface{}) {
	println(v)
	_, _ = v.(int)
}

func main() {
	var num1 int = 10
	var str1 string = "Hello"

	printValue(num1) // ничего не алоцируется в куче
	printValue(str1)

	var num2 int = 10
	var str2 string = "Hello"

	var i interface{}
	i = num2
	i = str2
	_ = i
}
```

Подробнее про то, когда переменная уходит в кучу, а когда остаётся на стеке — разберём отдельно в теме про аллокаторы и escape analysis.

## Компилятор не проверяет внутренние типы интерфейса на этапе компиляции

⚠️ Это значит, что можно легко получить панику в рантайме там, где статическая типизация никак не спасёт:

```go
package main

import "fmt"

func main() {
	var lhs interface{} = []int{1, 2, 3}
	var rhs interface{} = []int{1, 2, 3}
	fmt.Println(lhs == rhs) // 💥 panic: comparing uncomparable type []int
}
```

Слайсы (как и мапы, и функции) не поддерживают сравнение `==`. Но пока они спрятаны за `interface{}`, компилятор это никак не проверяет — ошибка вылезет только в рантайме, при попытке реально сравнить значения.

Аналогичная история с мапами, где ключ — интерфейс:

```go
package main

func main() {
	var data []int
	dictionary := make(map[interface{}]struct{})
	dictionary[data] = struct{}{} // 💥 panic: hash of unhashable type []int
}
```

Это ещё один аргумент в пользу того, чтобы избегать `interface{}` где только можно — компилятор буквально не может защитить тебя от такого рода ошибок.

## Interface guard pattern

Способ явно и на этапе компиляции проверить, что твой тип реализует нужный интерфейс — не дожидаясь рантайма или использования типа где-то ещё:

```go
package main

import "io"

type T struct {
	// ...
}

// Interface guard: если T перестанет реализовывать io.ReadWriter,
// компиляция упадёт прямо тут, а не там, где T реально используется
var _ io.ReadWriter = (*T)(nil)

func (t *T) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (t *T) Write(p []byte) (n int, err error) {
	return 0, nil
}
```

💡 Плюс: ошибку несоответствия интерфейсу видно сразу рядом с объявлением типа, а не где-то далеко в другом пакете, когда кто-то попытается присвоить `T` переменной интерфейсного типа.

## Методы — это просто функции с receiver-ом

Хороший пример, который наглядно это показывает — метод можно вызвать как обычную функцию, передав receiver первым аргументом:

```go
package main

import "fmt"

type Interface interface {
	process(int) bool
}

type String string

func (s String) process(size int) bool {
	return len(s) > size
}

func main() {
	var i String = String("inteface")

	fmt.Println(i.process(10))                                  // обычный вызов метода
	fmt.Println(Interface.process(i, 10))                        // метод интерфейса как функция
	fmt.Println(interface{ process(int) bool }.process(i, 10))   // даже с анонимным интерфейсом
}
```

Это работает, потому что под капотом метод `(s String) process(size int) bool` — это просто функция `process(s String, size int) bool`. `receiver` — синтаксический сахар для первого аргумента.

## Где объявлять интерфейс: producer side vs consumer side

Это холиварная тема (привет, Rob Pike и "accept interfaces, return structs"). Два варианта:

- **Producer side** — интерфейс объявлен в том же пакете, что и конкретная реализация.
- **Consumer side** — интерфейс объявлен во внешнем пакете, где он используется (потребитель сам решает, какой минимальный набор методов ему нужен).

|Критерий|Producer side|Consumer side|
|---|---|---|
|**Связность пакетов (coupling)**|❌ пакет `service` зависит от пакета `storage`|✅ `service` не зависит от `storage`|
|**Понятность кода**|❌ много лишних методов в интерфейсе, которые конкретному потребителю не нужны|✅ в интерфейсе только то, что реально нужно|
|**Гибкость (замена реализации)**|❌ сложно подменить реализацию, придётся писать заглушки под весь интерфейс|✅ легко подменить реализацию, заглушки не нужны|
|**Изменяемость (сигнатуры)**|✅ просто поменять сигнатуру в одном месте|❌ трудно менять сигнатуры сразу везде, где интерфейс переиспользуется|

**Вывод:** в Go принято объявлять интерфейсы **по месту использования** (consumer side), а из функций возвращать конкретные реализации, а не интерфейсы. Это даёт минимальные, легко тестируемые и слабо связанные абстракции.

### Но не всё так просто

В стандартной библиотеке кое-где интерфейсы всё-таки лежат на стороне производителя. Например, пакет `encoding` определяет интерфейсы (`encoding.TextMarshaler`, `encoding.BinaryMarshaler` и т.п.), которые реализуют другие субпакеты — `encoding/json`, `encoding/binary`.

Поэтому правило "интерфейсы на стороне потребителя" — это дефолт и хорошая практика, но не догма. Если ты **точно знаешь**, что абстракция будет полезна многим потребителям заранее (как в случае с `encoding`), имеет смысл объявить интерфейс на стороне производителя.