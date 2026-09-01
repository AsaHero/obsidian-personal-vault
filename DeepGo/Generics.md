# Generics в Go

## Почему без них жилось плохо

До generics (до Go 1.18) приходилось либо копипастить один и тот же код под разные типы, либо прятать всё за `interface{}` и терять типобезопасность (плюс получать overhead на dynamic dispatch и type assertion).

Классический симптом — стандартная библиотека буквально генерировала отдельные константы под каждый размер числа: `MaxInt`, `MaxInt8`, `MaxInt16`, `MaxInt32`, `MaxInt64`... Это и есть тот самый **boilerplate**, о котором речь. Одна и та же логика ("максимальное значение для целого типа"), продублированная N раз только потому, что типы разные.

**Обобщённое программирование** (generic programming, оно же метапрограммирование в широком смысле) — это парадигма, которая позволяет писать код, работающий с разными типами данных без потери типобезопасности и без дублирования логики.

```go
package main

// constraint
type Signed interface {
	int | int8 | int16 | int32 | int64
}

// constraint
type Unsigned interface {
	uint | uint8 | uint16 | uint32 | uint64
}

// constraint
type Integer interface {
	Signed | Unsigned
}

func Max[T Integer](lhs, rhs T) T {
	if lhs >= rhs {
		return lhs
	}
	return rhs
}

func main() {
	var lhs1, rhs1 int = 10, 20
	_ = Max[int](lhs1, rhs1)

	var lhs2, rhs2 uint = 10, 20
	_ = Max[uint](lhs2, rhs2)
}
```

Один `Max`, который работает и с `int`, и с `uint`, и со всем остальным целочисленным зоопарком — без copy-paste и без `interface{}`.

## Constraint — что это такое

**Constraint (ограничение)** — это способ ограничить, какие типы можно подставлять в качестве type parameter. По сути это **интерфейс**, который может содержать:

- набор методов (как обычный интерфейс)
- и/или набор конкретных типов (что раньше в обычных интерфейсах было нельзя)

💡 Аналогия, которая реально помогает понять: связь между **constraint и type parameter** такая же, как связь между **типом и значением**. Тип ограничивает, какие значения допустимы для переменной — constraint ограничивает, какие типы допустимы для параметра типа.

## Инстанцирование (instantiation)

**Инстанцирование** — это передача конкретных type arguments в generic-функцию/тип. Происходит **на этапе компиляции**.

Из этого следует важное трио свойств:

- ✅ **типобезопасность сохраняется** — компилятор точно знает, с какими типами будет работать код
- ✅ **нет overhead-а в рантайме** — в отличие от dynamic dispatch через интерфейсы, здесь нет косвенных вызовов
- ⚠️ **но растёт время компиляции** — потому что под каждую использованную комбинацию типов компилятор генерирует отдельную версию функции/типа (подробнее — в разделе про размер бинарника ниже)

### Type parameters vs Type arguments

Не путать эти два термина:

```go
// T1, T2 — type parameters (объявляются в сигнатуре)
func Do[T1 any, T2 any](x T1, y T2) {
	// ...
}

// string, int — type arguments (передаются при вызове)
Do[string, int]("hello", 1)
```

**Типы-параметры** — это то, что объявлено в сигнатуре функции. **Типы-аргументы** — то, что реально подставляется при вызове (компилятор чаще всего сам их выводит через type inference, поэтому в `Do("hello", 1)` можно вообще не писать `[string, int]` явно).

## Constraint и defined types (`~`)

Вот тут кроется частый source of confusion — разница между `int` и `~int` в constraint:

```go
package main

type Integer1 interface {
	int | int16 | int32 | int64
}

type Integer2 interface {
	~int | ~int16 | ~int32 | ~int64
}

func Process1[T1 Integer1](value T1) {
	// ..
}

func Process11[T2 interface{ int | int16 | int32 | int64 }](value T2) {
	// ..
}

func Process12[T2 int | int16 | int32 | int64](value T2) { // interface{...} можно опустить
	// ..
}

func Process2[T2 Integer2](value T2) {
	// ..
}

type MyInt1 int
type MyInt2 MyInt1

type MyIntAlias1 = int
type MyIntAlias2 = MyIntAlias1

func main() {
	Process1(int(100))
	Process1(MyInt1(100))         // ❌ не скомпилируется: MyInt1 — defined type, не int
	Process1(MyInt2(100))         // ❌ то же самое
	Process1(MyIntAlias1(100))    // ✅ ок, это просто alias для int
	Process1(MyIntAlias2(100))    // ✅ ок, alias алиаса — всё ещё int

	Process2(int(100))
	Process2(MyInt1(100))         // ✅ ок, ~int matчит любой тип с underlying type int
	Process2(MyInt2(100))         // ✅ ок
}
```

⚠️ **Ключевой gotcha:**

- `int` в constraint — matчит **только сам тип `int`** (ну и его type alias — `MyIntAlias1 = int` это не новый тип, а просто другое имя того же типа)
- `~int` — matчит **любой defined type, у которого underlying type — `int`** (например `type MyInt1 int`)

Это принципиально важно, если ты хочешь, чтобы твоя generic-функция работала не только с "чистым" `int`, но и с типами вроде `type UserID int`.

## Constraint с методами

Constraint может требовать не только конкретные типы, но и наличие методов — можно комбинировать:

```go
package main

type MyInt int

func (i MyInt) String() string {
	return "number"
}

type Constraint interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
	String() string
	any

	// Do()
	// interface{ Do() }
	// ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func Do[T Constraint](value T) {
	// ...
}

func main() {
	var value MyInt
	Do(value) // ок, MyInt matчит ~int и реализует String()
}
```

**Правило чтения constraint:**

- **Новая строка** внутри интерфейса — это **AND** (тип должен удовлетворять всем строкам сразу)
- **`|` на одной строке** — это **OR** (тип должен матчить хотя бы одну альтернативу из списка)

То есть выше `Constraint` требует: (тип с underlying `~int...`) **И** (реализует `String() string`) **И** (удовлетворяет `any`, что тривиально всегда true).

## Что можно и нельзя писать в constraint

Тут есть свои строгие правила, которые компилятор проверяет:

```go
package main

type _ interface {
	[]int | comparable // ❌ compilation error
}

type _ interface {
	string | error // ❌ compilation error
}

type _ interface {
	string | interface{ Do() } // ❌ compilation error
}

type _ interface {
	string | interface{ int } // ✅ ok
}

type _ interface {
	string | interface{} // ✅ ok
}
```

Плюс есть отдельное ограничение именно для дженериков насчёт смешивания `T` и `~T` в одном OR-списке:

```go
package main

// generics restriction
type _ interface {
	int | ~int // ❌ compilation error — int уже покрывается ~int, дублирование запрещено
}

type _ interface {
	interface{ int } | interface{ ~int } // ✅ ok — потому что это два разных вложенных интерфейса
}

type _ interface {
	int | interface{ ~int } // ✅ ok, по той же причине
}
```

💡 Тут не буду закапываться в формальную спецификацию (там свои тонкости про "overlapping type sets"), но практический вывод простой: если ловишь странную ошибку компиляции в constraint — скорее всего где-то пересекаются type sets, попробуй обернуть части в отдельные `interface{...}`.

## Generic constraint на основе другого generic-типа

Constraint тоже может быть параметризован и ссылаться на другой generic-тип:

```go
package main

// constraint
type Unsigned interface {
	uint | uint8 | uint16 | uint32 | uint64
}

type Slice[T any] []T

type SliceConstaint[E Unsigned] interface {
	Slice[E]
}

func Do[E Unsigned, T SliceConstaint[E]](values T) {
	// ...
}

func main() {
	var slice1 Slice[uint8]
	var slice2 Slice[uint16]

	Do(slice1)
	Do(slice2)
}
```

Это уже более продвинутый паттерн — пригождается, когда нужно ограничить не просто "любой срез", а "срез конкретно из unsigned-типов".

## Generics + структуры

Дженерики отлично работают и со структурами, не только с функциями — классический пример, generic `Set`:

```go
package main

import "fmt"

type Set[K comparable] struct {
	data map[K]struct{}
}

func NewSet[K comparable]() Set[K] {
	return Set[K]{
		data: make(map[K]struct{}),
	}
}

func (s *Set[K]) Insert(key K) {
	s.data[key] = struct{}{}
}

func (s *Set[K]) Erase(key K) {
	delete(s.data, key)
}

func (s *Set[K]) Contains(key K) bool {
	_, found := s.data[key]
	return found
}

// _ вместо K — если параметр типа не используется в теле метода,
// точно так же, как можно писать _ для неиспользуемого аргумента функции
func (s *Set[_]) Print() {
	fmt.Println(s.data)
}

func main() {
	set := NewSet[string]()
	set.Insert("key")
	set.Erase("key")
}
```

Обрати внимание: у методов на generic-типе **нельзя** объявить новые type parameters — они наследуются от типа (`K` тут пришёл из `Set[K]`), а не объявляются заново в сигнатуре метода.

## `comparable` — встроенный constraint

`comparable` — специальный предопределённый constraint, matчит любые типы, которые можно сравнивать через `==`/`!=` (числа, строки, указатели, массивы сравнимых типов, структуры из сравнимых полей). Именно поэтому в `Set[K comparable]` выше `K` может быть ключом мапы — мапе для ключей и нужен `comparable`.

## Как узнать реальный тип за дженериком

Внутри generic-функции `T` — это абстрактный type parameter, а не `any`. Поэтому **напрямую** делать type assertion или type switch на `T` **нельзя** — надо явно скастить в `any`:

```go
package main

import "fmt"

func IsInt1[T any](x T) bool {
	_, ok := any(x).(int) // ✅ явный каст в any — работает
	return ok
}

func Is1[T int | string](x T) {
	switch any(x).(type) { // ✅ тоже через any
	case int:
		fmt.Println("int")
	case string:
		fmt.Println("string")
	}
}

func IsInt2[T any](x T) bool {
	_, ok := x.(int) // ❌ cannot use type assertion on type parameter value
	return ok
}

func Is2[T int | string](x T) {
	switch x.(type) { // ❌ cannot use type switch on type parameter value
	case int:
		fmt.Println("int")
	case string:
		fmt.Println("string")
	}
}

func main() {
	fmt.Println("IsInt(100):", IsInt1(100))
	fmt.Println(`IsInt("100"):`, IsInt1("100"))

	Is1("100")
	Is1(100)
}
```

⚠️ **Gotcha:** ошибка "cannot use type assertion on type parameter value" вылезает довольно неожиданно для новичков в generics — интуитивно кажется, что раз `T` может быть разными типами, то assertion должен работать так же, как с `interface{}`. Но нет, `T` — это не интерфейс в этом смысле, компилятор требует явного `any(...)`.

## Нельзя встраивать generic type parameter как embedded field

```go
package main

type Base1[Derived any] struct {
	Derived // ❌ compilation error
}

type Base2[Derived any] struct {
	d Derived // ✅ ok — просто именованное поле
}
```

Это ограничение спецификации языка: embedding работает только с конкретными (не параметризованными) типами. Так что паттерн "generic mixin через embedding" в Go не работает — приходится просто явно называть поле.

## Когда вообще использовать generics

Простое практическое правило: **если ты видишь, что пишешь один и тот же код несколько раз, и единственная разница между копиями — используемые типы — вот тут и стоит подумать про generics.**

Типичные случаи:

- **Структуры данных** — когда пишешь обобщённые структуры (heap, binary tree, linked list, set, stack...) — они не должны быть завязаны на конкретный тип данных.
- **Функции** — когда логика функции работает одинаково для slice/map/channel любого типа (например, функция merge для каналов любого типа, или `Map`/`Filter`/`Reduce` для срезов).

**Главная цель generics — избежать дублирования кода**, то есть повысить возможность его переиспользования.

ⓘ Как бонус, в некоторых случаях generics дают:

- более чистый и простой API (не нужно городить `interface{}` + type assertion наружу)
- лучшую производительность в рантайме (нет overhead-а на dynamic dispatch, всё резолвится на этапе компиляции)

## Generics — не бесплатный синтаксический сахар

⚠️ Важно понимать: generics **не** free lunch. Они влияют на **размер скомпилированного бинарника**.

Причина: для каждой уникальной комбинации type arguments компилятор генерирует **отдельную специализацию** функции/типа (похоже на C++ templates, хоть механизм под капотом и отличается):

```go
type Data[T any] struct {
	Value T
}

d1 := Data[int]{}
d2 := Data[uint]{}

// компилятор генерирует что-то вроде...
type DataInt struct {
	Value int
}

type DataUint struct {
	Value uint
}
```

То есть чем больше разных типов ты используешь с одним generic-типом/функцией — тем больше кода реально попадёт в бинарник. Для большинства проектов это не проблема, но если у тебя очень много специализаций (например, generic-код, используемый с десятками разных типов) — стоит держать это в голове при профилировании размера бинарника и времени компиляции.

## Указатель как type parameter — разные способы

Есть несколько способов сказать "параметр типа — это указатель", и не все из них рабочие:

```go
package main

func process1[T int32 | int64](value *T) {
	*value = *value + 1 // ✅ ok — T это int32/int64, *T это указатель на него
}

func process2[T *int32 | *int64](value T) {
	*value = *value + 1 // ❌ compilation error
}

func process3[T *Int, Int int32 | int64](value T) {
	*value = *value + 1 // ✅ ok
}
```

⚠️ **Gotcha:** `process2` не компилируется, хотя выглядит вроде бы логично — constraint из указательных типов `*int32 | *int64` не даёт компилятору достаточно информации, чтобы разыменовать `value` и присвоить туда значение (там нет гарантии, что `*T` — это именно то, что можно разыменовывать таким образом в этом контексте). А вот `process3` работает, потому что там явно два параметра типа: `Int` — базовый числовой тип, а `T` constraint-ится как указатель именно на этот `Int`.

Практический вывод: если нужен generic-код, работающий с указателями и меняющий значение через разыменование — используй схему как в `process1`/`process3` (T — сам тип, а указатель добавляется отдельно в сигнатуре параметра), а не "constraint из самих указательных типов".

## Map и slice как constraint

Ещё один неочевидный кейс — как индексация ведёт себя в зависимости от того, из чего состоит constraint:

```go
package main

func process1[T []byte | [2]byte | string](value T) {
	_ = value[0] // ✅ ok — у всех вариантов в OR-списке есть индексация по int
}

func process2[T map[int]string](value T) {
	_ = value[0] // ✅ ok — тут T это конкретно один map-тип
}

func process3[T map[int]string | []byte](value T) {
	_ = value[0] // ❌ compilation error
}
```

⚠️ **Gotcha:** `process3` не компилируется, хотя по отдельности и map, и slice поддерживают `value[0]`. Причина в том, что у map и slice **разная семантика индексации** (у мапы `[0]` — это lookup по ключу с возможным `ok`, у среза — это доступ по индексу с паникой при выходе за границы), и компилятор не может унифицировать операцию `[...]` для смешанного constraint из принципиально разных категорий типов.

И отдельный момент про **мутацию** через generic с `~string | ~[]byte`:

```go
package main

func process[T ~string | ~[]byte](value T, index int) {
	_ = value[index]        // ✅ ok — чтение работает для обоих вариантов
	value[index] = byte('a') // ❌ error — а вот запись уже нет
}
```

⚠️ Тут ошибка возникает потому, что **строки в Go неизменяемы** (immutable) — байты строки нельзя перезаписать через индекс, а вот байты среза (`[]byte`) — можно. Раз `T` может быть и тем, и другим, компилятор обязан запретить операцию, которая невалидна хотя бы для одного из вариантов в type set — иначе бы для строк это была runtime-паника вместо честной compile-time ошибки.