## Интроспекция vs Рефлексия

**Интроспекция** — способность программы исследовать тип или свойства объекта во время выполнения (runtime). Это just "посмотреть".

**Рефлексия (reflection)** — более широкое понятие: способность программы **изучать И модифицировать** свою собственную структуру и поведение (значения, метаданные, свойства, функции) во время выполнения. То есть интроспекция — это "посмотреть", а рефлексия — это "посмотреть и поменять". Рефлексия — отдельная форма метапрограммирования.

## #1 Reflection распространяется от интерфейса до reflection-объекта

На базовом уровне пакет `reflect` — это просто механизм для изучения пары **(тип, значение)**, которая хранится внутри интерфейсной переменной. Помнишь из темы про интерфейсы, что интерфейс под капотом — это пара `(itab/type, data)`? Reflection буквально работает с этой парой напрямую.

ⓘ Чтобы начать работать с reflection, нужно знать два ключевых типа:

- `reflect.Type` — информация о типе
- `reflect.Value` — информация о значении

Они получаются через две простые функции — `reflect.TypeOf()` и `reflect.ValueOf()` соответственно. По сути они "достают" тип и значение из интерфейсной переменной (не забывай, что при вызове `TypeOf(i any)` / `ValueOf(i any)` твоё значение уже неявно завёрнуто в `interface{}`, потому что таковы сигнатуры функций).

```go
package main

import (
	"fmt"
	"reflect"
)

func main() {
	var value float64 = 3.4

	// неявное преобразование в пустой интерфейс при передаче в TypeOf/ValueOf
	fmt.Println("type:", reflect.TypeOf(value))  // func TypeOf(i any) reflect.Type
	fmt.Println("type:", reflect.ValueOf(value)) // func ValueOf(i any) reflect.Value

	equal := reflect.ValueOf(value).Type() == reflect.TypeOf(value)
	fmt.Println("equal:", equal) // true — Value.Type() и TypeOf дают одно и то же

	kind := reflect.ValueOf(value).Kind()
	fmt.Println("equal reflect.Float64:", reflect.Float64 == kind)

	value = reflect.ValueOf(value).Float()
	_ = value
}
```

### Почему API простой: getter/setter работают с "самым большим" типом

Чтобы `reflect.Value` не пришлось городить отдельные методы под каждый конкретный размер числа, "getter"/"setter" методы работают с **самым большим типом той же категории**: например для всех целых чисел со знаком (`int8`, `int16`, `int32`, `int64`, `int`) используется `int64`, а сам метод называется `Int()`. Для беззнаковых — `Uint()` (возвращает `uint64`), для float — `Float()` (возвращает `float64`).

```go
package main

import (
	"fmt"
	"reflect"
)

func main() {
	var x uint8 = 10
	v := reflect.ValueOf(x)

	fmt.Println("type:", v.Type()) // uint8
	fmt.Println("kind:", v.Kind()) // uint8

	x = uint8(v.Uint()) // ⚠️ Uint() возвращает uint64, надо кастовать обратно вручную
	_ = x
}
```

⚠️ **Подводный камень:** `v.Uint()` всегда вернёт `uint64`, даже если реальный тип был `uint8`. Тебе нужно вручную привести результат обратно к нужному размеру (`uint8(v.Uint())`). Это упрощает API пакета `reflect` (не нужно 5 разных методов под 5 разных размеров), но легко забыть про это преобразование.

### `Kind()` — что это

`Kind()` возвращает **базовую категорию типа** (её ещё называют "kind" — `int`, `struct`, `slice`, `map`, `float64` и т.д.), в отличие от `Type()`, который возвращает конкретный (возможно, именованный) тип.

Разница важна, когда у тебя есть defined type: например `type MyInt int` — `Type()` вернёт `MyInt`, а `Kind()` вернёт `reflect.Int`. То есть `Kind()` говорит "как это устроено под капотом", а `Type()` — "как это называется".

## #2 Reflection распространяется от reflection-объекта обратно до интерфейса

Имея `reflect.Value`, можно восстановить обратно интерфейсное значение через метод `Interface()`. Он упаковывает информацию о типе и значении обратно в `interface{}` и возвращает результат — тебе останется только сделать type assertion до нужного конкретного типа.

```go
package main

import (
	"fmt"
	"reflect"
)

func main() {
	var value float32 = 3.14
	v := reflect.ValueOf(value)
	iValue := v.Interface()      // interface{}, упаковали обратно
	fmt.Println(iValue.(float32)) // достаём конкретный тип через assertion
}
```

⚠️ **Важно помнить про copy semantics** — везде, где происходит "каст" туда-обратно через reflection (`ValueOf`, `Interface()`), под капотом **копируется значение**. Это касается и обычных non-pointer типов при передаче в `interface{}` в принципе — reflection тут не исключение, просто это особенно легко упустить из виду, когда работаешь с абстрактным `reflect.Value`.

## #3 Чтобы изменить объект через reflection — значение должно быть settable

### Устанавливаемость (settability)

**Устанавливаемость** — это свойство, при котором `reflect.Value` может **изменить исходное хранимое значение**, а не просто прочитать его копию.

💡 Устанавливаемость немного похожа на **адресуемость (addressability)**, но она строже. Определяется тем, содержит ли reflection-объект **сам исходный элемент**, или только **его копию**.

```go
package main

import (
	"fmt"
	"reflect"
)

func main() {
	var value float64 = 3.14
	v := reflect.ValueOf(value) // это копия value!
	//v.SetFloat(2.7)           // 💥 паника, если раскомментировать: value is not addressable
	//v.Addr()                 // 💥 то же самое

	fmt.Println("settability of v:", v.CanSet())
}
```

Вывод:

```
settability of v: false
```

⚠️ Это не должно казаться странным, если вспомнить, как работают обычные функции в Go: когда ты передаёшь значение в функцию (не указатель), функция получает **копию**, и логично, что она не может поменять оригинал. `reflect.ValueOf(value)` устроен точно так же — как будто ты "передал" `value` в функцию `ValueOf` по значению.

**Практический вывод:** если хочешь менять значение через reflection — нужно передавать **указатель** на это значение, точно так же, как ты бы сделал это с обычной функцией.

```go
package main

import (
	"fmt"
	"reflect"
)

func main() {
	var value float64 = 3.14
	v := reflect.ValueOf(&value) // передаём указатель

	fmt.Println("type of v:", v.Type())               // *float64
	fmt.Println("settability of v:", v.CanSet())       // false — сам v (указатель) не settable
	fmt.Println("addresability of v:", v.CanAddr())    // false

	fmt.Println("type of v.Elem():", v.Elem().Type())             // float64 — разыменовали
	fmt.Println("settability of v.Elem():", v.Elem().CanSet())    // true!
	fmt.Println("addresability of v.Elem():", v.Elem().CanAddr()) // true!

	v.Elem().SetFloat(2.7) // меняем реальное значение через reflection

	fmt.Println(v.Elem().Addr())
	fmt.Println("value:", value) // 2.7 — оригинал реально изменился
	fmt.Println("address:", &value)
}
```

Вывод:

```
type of v: *float64
settability of v: false
addresability of v: false
type of v.Elem(): float64
settability of v.Elem(): true
addresability of v.Elem(): true
0x2a84c183c048
value: 2.7
address: 0x2a84c183c048
```

⚠️ **Ключевой gotcha тут:** сам `v` (reflect.Value указателя) — **не settable**, потому что ты не можешь "переставить" сам указатель через reflection в этом смысле. А вот `v.Elem()` (разыменованное значение, на которое указывает указатель) — **settable**, потому что это как раз доступ к оригинальным данным по адресу. `.Elem()` — это методический аналог оператора `*` для reflect.Value.

## Ограничения reflection (по состоянию на Go 1.22-1.23)

Стоит держать в голове несколько принципиальных ограничений пакета `reflect`:

1. **Нельзя создавать типы интерфейсов через reflection.** Можно работать с существующими интерфейсными типами, но объявить новый interface type в рантайме через `reflect` — нельзя.
    
2. **Создание struct type с embedded (anonymous) полями — не полностью надёжно.** Через reflection можно создать структуру с анонимными встроенными полями, но получит ли эта структура методы встроенных типов — не гарантировано, и в некоторых случаях создание такой структуры может вообще запаниковать в рантайме. Поведение частично зависит от реализации компилятора.
    
3. **Нельзя объявлять новые именованные типы через reflection.** Можно строить составные типы (`reflect.SliceOf`, `reflect.PointerTo`, `reflect.StructOf` и т.д.), но объявить полноценный defined type с собственным именем (как `type MyInt int`) — нельзя.
    

Эти ограничения — не баг, а следствие того, что `reflect` в первую очередь работает с уже существующей информацией о типах, сгенерированной компилятором, а не подменяет собой сам компилятор.

## Работа с полями структуры через `reflect.Type`

```go
package main

import (
	"fmt"
	"reflect"
)

type Data struct {
	Count int
	Title string
}

func (d Data) Do() {}

func main() {
	data := Data{}

	tData := reflect.TypeOf(data)
	fmt.Println("Kind:", tData.Kind())       // struct
	fmt.Println("PkgPath:", tData.PkgPath()) // package path, где объявлен тип

	fmt.Println("NumField:", tData.NumField()) // 2
	fmt.Println("Field(0):", tData.Field(0).Name) // "Count"
	fmt.Println("Field(1):", tData.Field(1).Name) // "Title"

	fmt.Println("NumMethod:", tData.NumMethod())            // 1
	fmt.Println("Method(0):", tData.Method(0).Name)         // "Do"
	fmt.Println("Method(0).NumIn:", tData.Method(0).Type.NumIn())  // кол-во аргументов (receiver считается первым!)
	fmt.Println("Method(0).NumOut:", tData.Method(0).Type.NumOut()) // кол-во возвращаемых значений
}
```

⚠️ **Подводный камень**: `Method(0).Type.NumIn()` для метода, полученного через `reflect.TypeOf`, будет включать **receiver как первый аргумент** — прямо как мы разбирали в теме про интерфейсы, что метод это просто функция с receiver-ом первым параметром. Так что для `Do(d Data)` `NumIn()` вернёт `1` (сам receiver `d`), а не `0`.

## Struct tags через reflection

Классический кейс, зачем reflection реально нужен в проде — чтение struct tags (то, на чём построены `encoding/json`, ORM-библиотеки и т.д.):

```go
package main

import (
	"fmt"
	"reflect"
)

type Data struct {
	X int  `json:"x" xml:"name"`
	Y bool `json:"y,omitempty"`
}

func main() {
	t := reflect.TypeOf(Data{})
	tXTag := t.Field(0).Tag
	tYTag := t.Field(1).Tag

	fmt.Println(reflect.TypeOf(tXTag), reflect.TypeOf(tYTag)) // reflect.StructTag

	value, present := tXTag.Lookup("json")
	fmt.Println(value, present) // "x" true

	value, present = tYTag.Lookup("json")
	fmt.Println(value, present) // "y,omitempty" true
}
```

💡 `StructTag` — это просто `string` под капотом (`type StructTag string`) со своим набором методов (`Get`, `Lookup`) для парсинга формата `key:"value"`. `Lookup` в отличие от `Get` возвращает ещё и `present bool` — полезно, чтобы отличить "тег есть, но пустой" от "тега вообще нет".

## Вызов методов и функций через reflection

Можно не только читать данные, но и **вызывать** методы и произвольные функции динамически, через `Call`:

```go
package main

import (
	"fmt"
	"reflect"
)

type Vector struct {
	X int
	Y int
}

func (v Vector) Add(factor int) int {
	return (v.X + v.Y) * factor
}

func main() {
	vector := Vector{X: 5, Y: 15}
	vVector := reflect.ValueOf(vector)

	vAdd := vVector.MethodByName("Add") // ищем метод по имени как строке
	vResults := vAdd.Call([]reflect.Value{reflect.ValueOf(2)}) // аргументы — тоже []reflect.Value

	fmt.Println(vResults[0].Int()) // (5+15)*2 = 40

	// точно так же работает и с обычными функциями (не только методами)
	negative := func(x int) int {
		return -x
	}

	vNegative := reflect.ValueOf(negative)
	vResults = vNegative.Call([]reflect.Value{reflect.ValueOf(100)})
	fmt.Println(vResults[0].Int()) // -100
}
```

⚠️ Обрати внимание: `Call` возвращает `[]reflect.Value` (даже если метод возвращает одно значение) — потому что через reflection нельзя заранее статически знать, сколько значений вернёт вызываемая функция/метод. Работа с результатом всегда идёт через индексацию плюс явные методы вроде `.Int()`, `.String()`, `.Interface()` и т.п.

`MethodByName` ищет метод по строковому имени в рантайме — это удобно для generic-фреймворков (например DI-контейнеров или ORM), но плата за это — полное отсутствие compile-time проверки: опечатался в имени метода — получишь панику в рантайме, а не ошибку компиляции.

## Работа с map через reflection

```go
package main

import (
	"fmt"
	"reflect"
)

func main() {
	data := map[string]int{"data1": 100, "data2": 200}

	vData := reflect.ValueOf(data)
	vData.SetMapIndex(reflect.ValueOf("data1"), reflect.ValueOf(1000))
	vData.SetMapIndex(reflect.ValueOf("data2"), reflect.ValueOf(2000))

	for it := vData.MapRange(); it.Next(); {
		fmt.Println(it.Key(), it.Value())
	}
}
```

💡 Обрати внимание, что `SetMapIndex` тут реально работает и меняет исходную мапу `data`, хотя мы вроде бы передали `data` "по значению" в `reflect.ValueOf(data)`. Это не противоречит правилу про settability выше — **мапа в Go сама по себе является ссылочным типом** (внутри — указатель на runtime hmap структуру), поэтому даже "копия" `reflect.Value` мапы всё ещё указывает на те же самые данные. То же самое было бы верно и без reflection, если бы ты просто передал `map` в обычную функцию и поменял там значение по ключу.

`MapRange()` — это способ проитерироваться по мапе через reflection (аналог обычного `for k, v := range data`, но когда конкретный тип мапы неизвестен на этапе компиляции).