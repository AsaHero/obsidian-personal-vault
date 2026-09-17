Как вообще можно сигнализировать об ошибке? Вариантов на самом деле не так много:

- Флаг (true/false, статусы)
- Optional
- Вернуть error из функции
- Exception (исключения)

Го выбрал третий путь, но чтобы понять, почему именно так — стоит сначала посмотреть, что не так с остальными.

## Флаги (bool)

Самый простой вариант — просто сказать "получилось / не получилось".

```go
package main

import "fmt"

// вариант 1: возвращаем bool вторым значением
func divideV1(lhs, rhs int) (int, bool) {
	if rhs == 0 {
		return 0, false
	}

	return lhs / rhs, true
}

// вариант 2: пишем статус через указатель (out-параметр)
func divideV2(lhs, rhs int, status *bool) int {
	*status = false
	if rhs == 0 {
		return 0
	}

	*status = true
	return lhs / rhs
}

func main() {
	x := 100
	y := 0

	value, ok := divideV1(x, y)
	fmt.Println(value, ok) // 0 false

	value = divideV2(x, y, &ok)
	fmt.Println(value, ok) // 0 false
}
```

**Минус:** флаг говорит только "что-то пошло не так", но не говорит **что именно**. Если у функции пять разных способов сломаться — bool их все склеивает в одну картину.

## Статусы (коды ошибок)

Логичный шаг вперёд — вместо bool вернуть код, который различает причины.

```go
package main

import "fmt"

const (
	OkStatus = iota
	EvenNumberErr
	ZeroNumberErr
)

func divideV1(lhs, rhs int) (int, int) {
	if rhs == 0 {
		return 0, ZeroNumberErr
	} else if lhs%2 == 0 || rhs%2 == 0 {
		return 0, EvenNumberErr
	}

	return lhs / rhs, OkStatus
}

func divideV2(lhs, rhs int, status *int) int {
	if rhs == 0 {
		*status = ZeroNumberErr
		return 0
	} else if lhs%2 == 0 || rhs%2 == 0 {
		*status = EvenNumberErr
		return 0
	}

	*status = OkStatus
	return lhs / rhs
}

func main() {
	x := 100
	y := 0

	value, status := divideV1(x, y)
	fmt.Println(value, status)

	value = divideV2(x, y, &status)
	fmt.Println(value, status)
}
```

**Плюс:** уже можно различать причины ошибки. **Минус:** код — это просто число (`iota`), само по себе оно ничего не рассказывает. Плюс легко забыть проверить статус — компилятор тебя не заставит.

## Optional (привет из C++)

Идея — обернуть значение в контейнер, у которого можно спросить "а есть ли там что-то". Реализация ниже — обычный [[Generics|generic]] тип с параметром `[T any]`.

```go
package main

import "fmt"

var NullOptional = Optional[int]{}

type Optional[T any] struct {
	value   T
	present bool
}

func NewOptional[T any](value T) Optional[T] {
	return Optional[T]{
		value:   value,
		present: true,
	}
}

func (o *Optional[T]) HasValue() bool {
	return o.present
}

func (o *Optional[T]) Value() T {
	return o.value
}

func divide(lhs, rhs int) Optional[int] {
	if rhs == 0 {
		return NullOptional
	}

	result := lhs / rhs
	return NewOptional(result)
}

func main() {
	x := 100
	y := 0

	optional := divide(x, y)
	fmt.Println(optional)
}
```

⚠️ **Подводный камень:** ничто не мешает вызвать `Value()` не проверив `HasValue()` — получишь zero value типа и не поймёшь, что это "пустота", а не реальный 0. Optional красиво выглядит, но всё ещё не рассказывает **почему** ничего нет.

## Так в чём проблема у всех трёх подходов?

Все они либо:

- засоряют сигнатуру функции (лишний return-параметр, лишний out-параметр), либо
- вообще не несут информации о причине ошибки.

Есть ещё два способа не захламлять сигнатуру:

**Errno (подход из C).** Завести одну глобальную переменную, которая говорит "была ли ошибка". Плюс — сигнатура функции чистая. Минус — глобальное мутируемое состояние: не thread-safe из коробки, легко забыть сбросить/проверить, ошибка теряет привязку к конкретному вызову.

**Custom Errors через контракт.** Завести общий интерфейс-контракт, чтобы каждый мог сделать свой тип ошибки по этому контракту. Это уже прямой путь к тому, как сделано в Go.

## Ошибки в Go: главная идея

> **ОШИБКА ДОЛЖНА РАССКАЗЫВАТЬ ИСТОРИЮ**

Из этого следуют три простых правила:

1. Сообщение об ошибке должно быть максимально подробным
2. Сообщение должно содержать весь контекст для расследования причин
3. Сообщение должно однозначно указывать место возникновения ошибки

В Go `error` — это просто [[Interface|интерфейс]] (`type error interface { Error() string }`), и весь фреймворк ошибок строится вокруг этой простой идеи + пары стандартных функций (`errors.Is`, `errors.As`, `errors.Unwrap`).

### Stack trace

Стандартная библиотека `errors` стек вызовов не даёт — она вообще довольно минималистична. Если нужен полноценный stack trace, обычно берут сторонний пакет (например `pkg/errors` или аналоги), который добавляет трассировку прямо в ошибку при её создании.

### Sentinel error (сигнальная / дозорная ошибка)

Это ошибка, объявленная как глобальная переменная — нужна, чтобы можно было однозначно проверить: "а не случилась ли именно **эта** конкретная ошибка".

```go
var ErrNotFound = errors.New("not found")
```

### Правило: обрабатывай ошибку один раз

Частая проблема — одна и та же ошибка "проходит" через несколько уровней, и на каждом уровне её логируют. В итоге в логах пять записей об одном и том же инциденте, и непонятно — это один сбой или пять разных. Ошибку нужно обработать (залогировать / показать / принять решение) один раз, в одном месте, а не на каждом уровне, через который она пролетает.

## fmt.Errorf: %v vs %w

Хочется добавить контекст к ошибке, не обрабатывая её (не логируя дважды). Первая мысль — `Errorf`:

```go
package main

import (
	"errors"
	"fmt"
)

func main() {
	err := errors.New("source error")
	err = fmt.Errorf("additional error information: %v", err)

	fmt.Println(err.Error())
	fmt.Println(errors.Unwrap(err))
}
```

Вывод:

```
additional error information: source error
<nil>
```

**Почему не получилось достать source error обратно?** Потому что `%v` — это просто форматирование в строку. Новый `err` — это plain-строка, у него нет метода `Unwrap()`, он не знает, что "внутри" него раньше был другой error.

Начиная с Go 1.13 есть глагол `%w` — он оборачивает ошибку и создаёт у результата метод `Unwrap() error`. Именно эта обёртка делает ошибку доступной для `errors.Is()` и `errors.As()`.

```go
package main

import (
	"errors"
	"fmt"
)

func main() {
	err := errors.New("source error")
	err = fmt.Errorf("additional error information: %w", err)
	err = fmt.Errorf("internal error: %w", err)

	fmt.Println(err.Error())
	fmt.Println(errors.Unwrap(err))
	fmt.Println(errors.Unwrap(errors.Unwrap(err)))
	fmt.Println(errors.Unwrap(errors.Unwrap(errors.Unwrap(err))))
}
```

Получаем цепочку (chain) ошибок — каждый `Unwrap()` снимает один слой, как с луковицы.

### ⚠️ Bad wrapping — грабли, на которые легко наступить

```go
package main

import (
	"errors"
	"fmt"
)

func main() {
	err1 := errors.New("source error 1")
	err2 := errors.New("source error 2")
	err := fmt.Errorf("additional error information: %w and %w", err1, err2)

	fmt.Println(err.Error())
	fmt.Println(errors.Unwrap(err))

	err = fmt.Errorf("additional error information: %w", "error")

	fmt.Println(err.Error())
	fmt.Println(errors.Unwrap(err))
}
```

Тут два разных грабля:

- **Два `%w` в одном `Errorf`.** С Go 1.20 это разрешено (можно оборачивать сразу несколько ошибок), но получившийся error реализует не `Unwrap() error`, а `Unwrap() []error`. А вот `errors.Unwrap()` (единственного числа) ждёт именно сигнатуру `Unwrap() error` — и для такого multi-wrap возвращает `nil`. При этом `errors.Is`/`errors.As` продолжают работать корректно, потому что они умеют работать с обеими сигнатурами `Unwrap`. Вывод: не полагайся на `errors.Unwrap()` для multi-wrap ошибок, используй `Is`/`As`.
- **`%w` с не-error значением** (строка `"error"`). `%w` обязан получать именно значение с методом `Error()`. Строку он так обернуть не может — получишь некорректный вывод в духе `%!w(string=error)` в тексте ошибки, а `Unwrap()` вернёт `nil`, потому что оборачивать было нечего.

## Оборачивание (wrapping) — зачем это вообще

Оборачивание — это упаковка ошибки в контейнер-обёртку, который делает исходную ошибку доступной "изнутри".

Используется для:

- добавления контекста к ошибке;
- маркировки ошибки (пометить, что она прошла через конкретное место).

Чтобы проверить, **относится ли** обёрнутая ошибка к определённому **типу** — используем `errors.As()`. Она рекурсивно разворачивает всю цепочку, пока не найдёт совпадение по типу.

```go
package main

import (
	"errors"
	"fmt"
)

type DatabaseError struct{}

func (d DatabaseError) Error() string {
	return "database error"
}

func GetDataFromDB() error {
	return fmt.Errorf("failed to get data: %w", DatabaseError{})
}

func main() {
	err := GetDataFromDB()
	if errors.As(err, &DatabaseError{}) {
		fmt.Println(err.Error())
	} else {
		fmt.Println("unknown error")
	}
}
```

## ❌ НИКОГДА не сравнивайте err.Error()

`err.Error()` — это строка **для человека**, а не для кода. Она предназначена для лога или экрана, а не для `if err.Error() == "..."`. Причина проста: текст сообщения — деталь реализации, она может поменяться в любой момент (даже случайно, при рефакторинге), и твоя проверка молча сломается. Для проверки значения ошибки есть `errors.Is()`.

Чтобы проверить, относится ли обёрнутая ошибка к конкретному **значению** (а не типу) — используем `errors.Is()`. Она тоже рекурсивно разворачивает цепочку.

```go
package main

import (
	"errors"
	"fmt"
)

var ErrDatabaseProblem = errors.New("database problem")

func GetDataFromDB() error {
	return fmt.Errorf("failed to get data: %w", ErrDatabaseProblem)
}

func main() {
	err := GetDataFromDB()
	if errors.Is(err, ErrDatabaseProblem) {
		fmt.Println(err.Error())
	} else {
		fmt.Println("unknown error")
	}
}
```

## Sentinel error vs Custom error type — когда что использовать

```go
if errors.Is(err, ErrNotFound) { ... }   // sentinel error
```

```go
if errors.As(err, &DBError{}) { ... }    // custom error type
```

|Критерий|Sentinel error (`errors.Is`)|Custom error type (`errors.As`)|
|---|---|---|
|Что сравниваем|конкретное **значение** ошибки|конкретный **тип** ошибки|
|Что получаем|факт "это именно та ошибка"|доступ к полям структуры с доп. контекстом|
|Гибкость|низкая — просто да/нет|высокая — можно нести любой контекст|
|Пример|`ErrNotFound`|`*PathError`|
|Когда использовать|когда важен сам факт "эта ли ошибка произошла"|когда нужен контекст (что, где, с чем случилось)|

Большое преимущество отдельных типов ошибок — они могут нести дополнительный контекст. Классический пример из стандартной библиотеки — `os.PathError`:

```go
type PathError struct {
	Op   string
	Path string
	Err  error
}
```

Тут сразу видно: какая операция (`Op`), над каким путём (`Path`) и с какой исходной ошибкой (`Err`) — это и есть "ошибка рассказывает историю" в чистом виде.

## Несколько ошибок из одной функции

Если функция может накопить сразу несколько ошибок (например, при валидации нескольких полей) — используют пакет `go-multierror`. Он умеет склеивать список ошибок в одну, сохраняя возможность потом разобрать их обратно.

## Go не поддерживает исключения (но почти)

Формально в Go нет `try/catch/throw`. Предпочтительный путь — явная обработка ошибок через `error`. Но у Go всё же есть неявный механизм, похожий на исключения — **panic / recover**.

**panic — это не то же самое, что throw**, и разница тут семантическая, а не синтаксическая:

- когда ты бросаешь exception — ты делаешь ошибку **проблемой вызывающей стороны**, предполагая, что она сможет её обработать;
- когда ты вызываешь `panic` — ты **никогда не предполагаешь**, что вызывающий код сможет решить проблему.

> Поэтому `panic` используют только в действительно исключительных обстоятельствах, когда продолжать работу приложения невозможно (например: нарушен инвариант программы, повреждено внутреннее состояние).

```go
package main

import (
	"fmt"
)

func process() {
	fmt.Println("first")
	panic("error from process")
	fmt.Println("second") // никогда не выполнится
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recover:", r)
		}
	}()

	process()
}
```

При `panic` Go приостанавливает выполнение текущей функции и начинает **раскручивать стек** (stack unwinding): по очереди выполняет `defer` (см. базовый механизм в [[Functions#Defer|Functions]]) каждой функции вверх по стеку, пока либо не найдёт `recover()`, либо стек не закончится — и тогда приложение падает.

> ⚠️ Важно: `recover()` работает только в рамках стека **той же горутины**, где случился `panic` — поймать панику из другой горутины нельзя, подробнее в [[DeepGo/Goroutines#Panic в одной горутине не убивает другую|Goroutines]].

### ⚠️ Не все defer'ы выполняются одинаково

|Способ завершения|`defer` выполняются?|
|---|---|
|Обычный `return` / `panic` + `recover()`|✅ Да|
|`runtime.Goexit()`|✅ Да — выполняются все defer'ы стека **этой горутины**|
|`os.Exit(1)`|❌ Нет|
|Stack overflow|❌ Нет|
|OOM (out of memory)|❌ Нет|

Отсюда важный практический вывод: если у тебя есть критичная логика в `defer` (закрыть файл, записать лог, отпустить lock) — она **не гарантированно** выполнится при `os.Exit`, переполнении стека или нехватке памяти. На такие сценарии `defer` полагаться нельзя. Тот же принцип всплывёт ещё раз, когда речь пойдёт про [[Garbage Collector#Finalizers|финализаторы]] в GC — они тоже не гарантированы при аварийном завершении.