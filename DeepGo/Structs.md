## Что такое метод

Метод — это обычная [[Functions|функция]], но привязанная к какому-то типу (structure/type). ж в любом OOP-языке, только в Go нет классов — есть типы + методы, которые к ним "прикрепляются".

В Go есть приятный синтаксический sugar: не важно, работаешь ты со значением или с указателем — компилятор сам разыменует/возьмёт адрес там, где нужно, чтобы вызвать метод.

```Go
package main

type Data struct{}

func (d Data) Print() {}

func main() {
	var value1 = Data{}
	value1.Print() // вызов метода на значении

	// 1 способ получить указатель — через &
	var value2WithSugar = &Data{}
	value2WithSugar.Print() // Go сам разыменует указатель под капотом

	// 2 способ — через new(), тоже указатель
	var value2WithoutSugar = new(Data)
	(*value2WithoutSugar).Print() // а тут разыменовали руками, но это то же самое
}
```

Кстати, можно наплодить сколько угодно методов с именем `_` (blank identifier):

```Go
package main

type Data struct{}

func (d Data) _() {}
func (d Data) _() {}
func (d Data) _() {}
```

Смысла в этом ноль — вызвать такой метод нельзя, это чистой воды dead code. Просто забавная особенность языка, которую стоит знать, а не что-то, что реально пригодится в работе.

---

## Receiver: value vs pointer

Receiver — это аналог `this`/`self` из других языков: указатель (или значение) на тот объект, у которого вызвали метод.

В Go есть два вида receiver-а:

- **Value receiver** — метод получает **копию** структуры
- **Pointer receiver** — метод получает **указатель** на оригинал

```Go
package main

import "fmt"

type customer1 struct {
	balance int
}

// value receiver — работаем с копией
func (c customer1) add(value int) {
	c.balance += value
}

type customer2 struct {
	balance int
}

// pointer receiver — работаем с оригиналом
func (c *customer2) add(value int) {
	c.balance += value
}

func main() {
	c1 := customer1{}
	c1.add(100)

	c2 := customer2{}
	c2.add(100)

	fmt.Println(c1) // {0}   — изменения потерялись, работали с копией
	fmt.Println(c2) // {100} — изменения применились к оригиналу
}
```

Всё просто: value receiver — это как передать структуру по значению в обычную функцию, никакие изменения наружу не просочатся. Pointer receiver — это передача по ссылке.

### Когда что выбирать (шпаргалка от Go-комьюнити)

|Receiver **должен** быть pointer|Receiver **должен** быть value|
|---|---|
|Метод меняет состояние получателя|Нужно гарантировать неизменность получателя|
|В структуре есть поле, которое нельзя копировать (например, `sync.Mutex`)|Получатель — базовый тип (`int`, `float64`, `string`, ...)|

|Receiver **стоит** сделать pointer|Receiver **стоит** сделать value|
|---|---|
|Структура большая, копирование дорого (но это стоит проверять бенчмарком, а не гадать)|Изменяемые поля лежат не в самой структуре, а в другой структуре, на которую она ссылается (пример ниже)|

Пример на последний кейс — почему receiver можно оставить value, даже если внутри что-то меняется:

```Go
package main

import "fmt"

type account struct {
	balance int
}

type client struct {
	account *account // это указатель!
}

// value receiver — но это ок, потому что меняем не саму структуру client,
// а то, на что указывает её поле account
func (c client) add(value int) {
	c.account.balance += value
}

func main() {
	c := client{
		account: &account{},
	}

	c.add(100)
	fmt.Println(c.account.balance) // 100 — сработало!
}
```

**Почему это работает:** копируется сама структура `client`, но поле `account` внутри — это указатель, и копия указателя всё ещё смотрит на тот же `account` в памяти. Так что менять данные через него можно спокойно.

### ⚠️ Подводный камень: method value и binding

Вот тут кроется реально неочевидная штука. Смотри пример:

```Go
package main

import "fmt"

type Data struct {
	value int
}

func (d Data) Value1() int {
	return d.value
}

func (d *Data) Value2() int {
	return d.value
}

func main() {
	data := Data{100}
	pointer := &data

	// "берём метод" как отдельное значение (method value)
	value1ByData := data.Value1
	value1ByPointer := pointer.Value1

	value2ByData := data.Value2
	value2ByPointer := pointer.Value2

	data.value = 200 // меняем ПОСЛЕ того, как "забиндили" методы

	fmt.Println("value1ByData:", value1ByData())       // 100
	fmt.Println("value1ByPointer:", value1ByPointer())  // 100
	fmt.Println("value2ByData:", value2ByData())         // 200
	fmt.Println("value2ByPointer:", value2ByPointer())   // 200
}
```

**Почему так:** когда ты берёшь метод как значение (`data.Value1`), Go в момент этого присваивания уже "замораживает" receiver:

- Для **value receiver** метода — копирует структуру прямо сейчас (даже если брал через pointer — он его сначала разыменует и скопирует).
- Для **pointer receiver** метода — сохраняет just адрес, а не копию, поэтому все изменения после binding-а будут видны.

Отсюда и разница: `Value1` (value receiver) — навсегда запомнил `value=100`, а `Value2` (pointer receiver) — увидел актуальное `value=200` в момент вызова.

**Совет от комьюнити:** не смешивай типы receiver-а в рамках одной структуры. Как видишь, поведение получается непредсказуемым, если часть методов value, а часть pointer.

### Забавный кейс: вызов метода на nil-указателе

```Go
package main

import "fmt"

type Obect struct{}

func (o *Obect) Print() {
	if o == nil {
		fmt.Println("nil")
	} else {
		fmt.Println("not nil")
	}
}

func main() {
	var object *Obect // nil-указатель
	object.Print()    // и это не паника!
}
```

Output: `nil`

Вроде бы должна быть паника — ведь под капотом это `(*object).Print()`, разыменование nil? На самом деле нет: вызов метода **сам по себе** не разыменовывает указатель, паника случится только если внутри метода ты попытаешься обратиться к полям через `o.field`. А тут мы просто сравниваем `o == nil` — это безопасно.

### Как методы устроены под капотом

В ассемблере нет понятия "метод структуры" — там есть только [[Functions|функции]]. Метод — это sugar над функцией, где receiver — просто первый аргумент (по значению или по указателю, см. [[Functions#OS Stack on Function Call|передача аргументов через стек]]):

```Go
type data struct{}

func (d *data) methodWithSugar() {
	// implementation
}

// то же самое, но без сахара — receiver стал явным первым параметром
func methodWithoutSugar(d *data) {
	// implementation...
}
```

Это же соответствие receiver'а первому аргументу лежит в основе того, как в Go под капотом устроена [[Interface#Статическая vs Динамическая диспетчеризация|диспетчеризация методов через интерфейсы]].

И вот прямое доказательство — так называемое **method expression**, когда метод превращается в обычную функцию:

```Go
package main

import "fmt"

type Data struct{}

func (d Data) Print() {
	fmt.Println("data")
}

func main() {
	var data Data

	data.Print()  // обычный вызов
	(&data).Print()

	(Data).Print(data)   // method expression — метод как функция от значения
	(*Data).Print(&data) // method expression от указателя
}
```

---

## Union-подобные структуры через unsafe

Го не поддерживает union из коробки (как в C), но их можно эмулировать — использовать один и тот же участок памяти под разные типы данных, через пакет `unsafe`.

**Пример 1 — SBO (Small Buffer Optimization)** (тот же принцип лежит в основе **SSO** у строк, см. [[Strings#SSO — Small String Optimization|Strings]]). Идея: если данные маленькие — хранить их прямо внутри структуры, а если большие — хранить указатель + capacity в том же месте памяти:

```Go
package main

import "unsafe"

// SBO (Small Buffer Optimization)
type SBO struct {
	size  int64
	union [16]byte // либо маленький буфер, либо 8B[pointer] + 8B[capacity]
}

func main() {
	var small SBO
	small.size = 10
	small.union = [16]byte{} // маленькие данные хранятся прямо тут

	var big SBO
	big.size = 2048
	// интерпретируем те же самые 16 байт как указатель + размер
	pointer := (*[2048]byte)(unsafe.Pointer(&big.union))
	*pointer = [2048]byte{}
	capacity := (*int64)(unsafe.Add(unsafe.Pointer(&big.union), 8))
	*capacity = 2048
}
```

**Пример 2 — универсальный union на 8 байт**, куда можно писать то int64, то float64:

```Go
package main

import "unsafe"

type Union struct {
	value [8]byte
}

func (u *Union) SetInt64(value int64) {
	*(*int64)(unsafe.Pointer(&u.value)) = value
}

func (u *Union) GetInt64() int64 {
	return *(*int64)(unsafe.Pointer(&u.value))
}

func (u *Union) SetFloat64(value float64) {
	*(*float64)(unsafe.Pointer(&u.value)) = value
}

func (u *Union) GetFloat64() float64 {
	return *(*float64)(unsafe.Pointer(&u.value))
}
```

### ⚠️ Подводный камень

Это `unsafe` — тип safety тут полностью на твоей совести. Компилятор ничего не проверяет, легко словить memory corruption, если напутать с размерами/выравниванием. Используется в основном в перфоманс-критичном или low-level коде (стандартная библиотека, всякие allocators), а не в обычном бизнес-коде.

---

## Теги полей и анонимные структуры

В Go можно объявлять анонимные (inline) структуры прямо на месте, без объявления named type.

Важный момент про идентичность типов структур:

> **Два безымянных типа структур идентичны, только если у них одинаковая последовательность объявлений полей.** Поля идентичны, только если совпадают имя, тип **и тег** каждого поля.

То есть тег поля — это часть "подписи" типа структуры, не просто метаданные для reflection/json.

---

## Встраивание типов (Embedding)

Встроенные поля ещё называют **анонимными полями** — но на самом деле имя у них есть, просто неявное: имя встроенного типа и становится именем поля.

```Go
package main

type Bar struct {
	data int
}

type Foo struct {
	Bar // встроенный (анонимный) тип
}

func main() {
	var foo Foo

	foo.data = 100     // Go сам "продвигает" поля Bar наружу
	foo.Bar.data = 100 // это то же самое, компилятор превращает первое во второе
}
```

Это называется **field promotion** — поля (и методы!) встроенного типа "поднимаются" в содержащую структуру, и обращаться к ним можно и напрямую, и через полный путь.

Работает это и в глубину, через несколько уровней, и через указатели:

```Go
package main

type A struct {
	value int
}

func (a A) Print() {}

type B struct {
	A
}

type C struct {
	*B // встраивание через указатель — тоже работает!
}

func main() {
	var c C = C{B: &B{A: A{value: 10}}}

	// все три варианта эквивалентны благодаря promotion
	_ = c.B.A.value
	_ = c.A.value
	_ = c.value

	c.B.A.Print()
	c.B.Print()
	c.Print()
}
```

### ⚠️ Подводный камень: Ambiguous selector

Если два встроенных на одном уровне типа имеют поле/метод с одинаковым именем — обращение напрямую вызовет **ошибку компиляции**, потому что Go не понимает, какое из двух ты имеешь в виду:

```Go
package main

type Derived1 struct {
	values []int
}

func (d Derived1) Print() {}

type Derived2 struct {
	values []int
}

func (d Derived2) Print() {}

type Base struct {
	Derived1
	Derived2
}

func main() {
	var base Base

	// _ = base.values   // ❌ compile error: ambiguous selector
	// base.Print()      // ❌ compile error: ambiguous selector

	// нужно явно указывать путь
	base.Derived1.values = nil
	base.Derived2.values = nil

	base.Derived1.Print()
	base.Derived2.Print()
}
```

Важный нюанс: методы с одинаковой сигнатурой при встраивании **не переопределяют** друг друга (это не наследование в смысле OOP!) — они просто "конфликтуют"/перекрываются на одном уровне вложенности, и это заставляет тебя явно указывать, какой из них нужен.

---

## Выравнивание структур (memory alignment)

Важно понимать: **компилятор не может менять порядок полей**, который ты написал в коде структуры. Порядок влияет и на размер, и на скорость доступа.

### Кейс с массивом

```Go
package main

import (
	"fmt"
	"unsafe"
)

type data1 struct {
	aaa bool
	bbb [1023]byte
	ccc bool
}

func main() {
	d := data1{}
	fmt.Println(unsafe.Sizeof(d))  // 1025
	fmt.Println(unsafe.Alignof(d)) // 1
}
```

Почему выравнивание равно 1, а не больше? Потому что **выравнивание массива определяется выравниванием типа его элемента**. А у нас элемент — `byte`, у него выравнивание = 1. Отсюда и весь struct выравнивается по 1 байту — padding не нужен вообще.

Теперь другой пример, где вместо массива — `string`:

```Go
package main

import (
	"fmt"
	"unsafe"
)

type data1 struct {
	aaa bool
	bbb string // string весит 16 байт
	ccc bool
}

func main() {
	d := data1{}
	fmt.Println(unsafe.Sizeof(d))  // 32
	fmt.Println(unsafe.Alignof(d)) // 8
}
```

Тут выравнивание — 8, хотя сам `string` "весит" 16 байт. Почему? Потому что при вычислении alignment-а Go "распаковывает" всё до примитивных типов: `string` на самом деле — это структура `{ptr, len}`, где каждое поле по 8 байт. Вот эти 8 байт (размер самого широкого примитивного поля внутри) и определяют требуемое выравнивание всей структуры.

**Правило:** выравнивание массива всегда равно выравниванию типа его элемента.

### Как посчитать размер структуры руками (без unsafe/reflect)

```Go
package main

import (
	"fmt"
	"unsafe"
)

type data1 struct {
	aaa bool
	bbb int32
	ccc bool
}

type data2 struct {
	aaa int32
	bbb bool
	ccc bool
}

func main() {
	fmt.Println(unsafe.Sizeof(data1{})) // 12
	fmt.Println(unsafe.Sizeof(data2{})) // 8

	d := data1{aaa: true, bbb: 5, ccc: true}
	b := (*[12]byte)(unsafe.Pointer(&d))
	fmt.Printf("Bytes are %#v\n", b)
	// Bytes are &[12]uint8{0x1, 0x0, 0x0, 0x0, 0x5, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0}
}
```

Казалось бы, `bool + int32 + bool` — это 1+4+1 = 6 байт. А по факту — 12 и 8 байт в зависимости от порядка полей! И в дампе памяти видно кучу нулей между значениями — это **padding**.

### Требуемое выравнивание (required alignment)

Правило простое: **выравнивание структуры = размер самого большого поля в ней.**

- Если в структуре есть `int32` — выравнивание 4 байта.
- Если есть и `int32`, и `int64` — выравнивание 8 байт.

Каждое поле в памяти должно начинаться с адреса, кратного своему собственному выравниванию. Отсюда в `data1` (`bool, int32, bool`) после первого `bool` (1 байт) компилятор вставляет 3 байта padding, чтобы `int32` начинался с адреса, кратного 4. А в `data2` (`int32, bool, bool`) такой проблемы нет — оба `bool` идут подряд после `int32`, padding нужен только в конце, чтобы общий размер структуры был кратен выравниванию (4).

**Практический вывод:** если хочешь сэкономить память — сортируй поля от больших к маленьким (по убыванию размера). Это классический Go-совет по оптимизации памяти в структурах.

### Влияет ли количество методов на размер структуры?

Нет! Метод — это обычная функция, она никак не привязана физически к объекту структуры (см. секцию выше про "методы под капотом"). Сколько бы методов у типа ни было — на `unsafe.Sizeof` это не повлияет ни на байт.

### Размер пустой структуры — и неочевидный момент с адресом

Все знают, что `struct{}{}` весит 0 байт. Но раз она "ничего не весит", как у неё вообще может быть адрес в памяти?

```Go
package main

import (
	"fmt"
	"unsafe"
)

type empty struct{}

func main() {
	a := struct{}{}
	b := struct{}{}
	c := empty{}
	d := [0]byte{}

	fmt.Println(unsafe.Pointer(&a))
	fmt.Println(unsafe.Pointer(&b))
	fmt.Println(unsafe.Pointer(&c))
	fmt.Println(unsafe.Pointer(&d))
}
```

Output: у всех четырёх переменных **один и тот же адрес**. Это осознанное поведение рантайма Go: все нулевого размера объекты в куче указывают на один и тот же специальный адрес (`zerobase`), потому что реально выделять под них память бессмысленно.

Теперь смотрим, что происходит, если пустая структура — это **поле** внутри другой, не пустой структуры:

```Go
package main

import (
	"fmt"
	"unsafe"
)

func main() {
	type T1 struct {
		a struct{}
		x int64
	}

	var t1 T1
	fmt.Println("size:", unsafe.Sizeof(t1))         // 8
	fmt.Println("address a:", unsafe.Pointer(&t1.a)) // совпадает с x
	fmt.Println("address x:", unsafe.Pointer(&t1.x))

	type T2 struct {
		x int64
		a struct{} // теперь пустое поле — ПОСЛЕДНЕЕ
	}

	var t2 T2
	fmt.Println("size:", unsafe.Sizeof(t2))         // 16 (!)
	fmt.Println("address a:", unsafe.Pointer(&t2.a)) // НЕ совпадает с x
	fmt.Println("address x:", unsafe.Pointer(&t2.x))
}
```

Разница в размере (`T1` = 8 байт, `T2` = 16 байт) — из-за того, где стоит пустое поле:

- Если пустая структура **не последняя** — она просто не занимает места, "схлопывается" с соседним полем.
- Если пустая структура **последняя** в непустой структуре — компилятор добавляет padding после неё.

**Почему так специально сделано:** если бы padding не добавлялся, то `&t.a` (адрес последнего нулевого поля) указывал бы **за пределы** выделенного под структуру блока памяти. А это может случайно "зааллокировать" адрес, который принадлежит соседнему, вообще не связанному объекту в куче. Раз ты держишь указатель на этот адрес — GC не может собрать тот соседний объект → потенциальная утечка памяти. Чтобы этого избежать, компилятор Go подстраховывается и добавляет padding-байты именно после финального нулевого поля.

Похожая история и со стековыми переменными — пустая структура "занимает" адрес последней объявленной локальной переменной того же размера:

```Go
package main

import "time"

func fn() {
	var v3 int
	var e3 struct{}

	println("v3:", &v3)
	println("e3:", &e3)
}

func main() {
	var v1 int
	var e1 struct{}
	var v2 int
	var e2 struct{}

	println("v1:", &v1)
	println("e1:", &e1) // совпадает с v2!
	println("v2:", &v2)
	println("e2:", &e2) // совпадает с v2!

	fn()

	go func() {
		// у горутины отдельный стек — адреса будут другие
		var gv1 int
		var ge1 struct{}
		var gv2 int
		var ge2 struct{}
		println("gv2:", &gv2)
		println("ge2:", &ge2) // совпадает с gv2
	}()

	time.Sleep(time.Second)
}
```

Вывод: **у каждой горутины свой стек**, поэтому адреса на стеке горутины совсем другие, чем в `main`.

---

## [[Functions#Defer|Defer]] и структуры — пара нюансов

### Value vs pointer receiver при defer

```Go
package main

import "fmt"

type data1 struct {
	value int
}

func (d data1) print() { // value receiver
	fmt.Println("data1", d.value)
}

type data2 struct {
	value int
}

func (d *data2) print() { // pointer receiver
	fmt.Println("data2", d.value)
}

func main() {
	d1 := data1{}
	defer d1.print()

	d2 := data2{}
	defer d2.print()

	d1.value = 100
	d2.value = 200
}
```

Output:

```
data2 200
data1 0
```

**Почему так:** `defer` вычисляет **аргументы** отложенного вызова сразу, в момент выполнения строки `defer ...`, а не в момент, когда функция реально завершится (подробнее — [[Functions#Defer|Defer в Functions]]). А receiver — это, как мы уже разобрали, просто первый аргумент функции под капотом.

- Для `d1.print()` (value receiver) — Go копирует `d1` **сразу**, в момент `defer`, когда `d1.value` ещё `0`. Дальнейшее изменение `d1.value = 100` уже не видно.
- Для `d2.print()` (pointer receiver) — Go запоминает **адрес** `d2` сразу, но реально разыменовывает его только при выполнении, в конце функции — а к этому моменту `d2.value` уже `200`.

### Цепочка вызовов + defer

```Go
package main

type Data struct{}

func MakeData(pointer *int) Data {
	println("MakeData:", *pointer)
	return Data{}
}

func (Data) Print(pointer *int) {
	println("Print:", *pointer)
}

func main() {
	var value = 1
	var pointer = &value
	defer MakeData(pointer).Print(pointer)

	value = 2
	pointer = new(int)
	MakeData(pointer)
}
```

Output:

```
MakeData: 1
MakeData: 0
Print: 2
```

Тут та же логика, только более явно видна: в цепочке `defer A().B(x)` откладывается **только последний вызов** — `B`. А вот `A()` (получение receiver-а для `B`) и все аргументы (`x`) вычисляются немедленно, прямо в момент строки `defer`.

Поэтому: `MakeData(pointer)` внутри `defer` вызывается сразу же (`"MakeData: 1"`), а аргумент `pointer`, передаваемый в `Print`, тоже "замораживается" в этот момент — сохраняется именно тот адрес переменной `value`. Дальше в коде `pointer` переприсваивается на новый `new(int)`, но на отложенный вызов это уже не влияет — он держит старый адрес. А вот значение по старому адресу (`value`) успевает измениться на `2` до того, как реально выполнится отложенный `Print` — отсюда `"Print: 2"`.

**Вывод:** сколько бы вызовов ни было в цепочке через точку — деферится только самый последний, всё остальное отработает мгновенно.

---

## Сравнение структур: порядок полей влияет на скорость `==`

```Go
package main

import "testing"

// go test -bench=. comparison_test.go

type Data1 struct {
	size   int32
	values [10 << 20]byte // 10 MB
}

type Data2 struct {
	values [10 << 20]byte
	size   int32
}

func BenchmarkComparisonData1(b *testing.B) {
	data1 := Data1{size: 100}
	data2 := Data1{size: 101}
	for i := 0; i < b.N; i++ {
		_ = data1 == data2
	}
}

func BenchmarkComparisonData2(b *testing.B) {
	data1 := Data2{size: 100}
	data2 := Data2{size: 101}
	for i := 0; i < b.N; i++ {
		_ = data1 == data2
	}
}
```

`BenchmarkComparisonData1` отрабатывает **намного** быстрее `BenchmarkComparisonData2`. Причина: Go сравнивает поля структуры **по порядку объявления** и останавливается на первом же несовпадении (short-circuit). В `Data1` первое поле — `size`, оно отличается (100 vs 101) → сравнение завершается мгновенно, до 10-мегабайтного массива дело не доходит. В `Data2` порядок обратный: сначала сравнивается весь 10 MB массив (пусть даже он одинаковый в обоих случаях!), и только потом — `size`.

**Практический вывод:** если структуру часто сравнивают через `==` и в ней есть тяжёлые поля (большие массивы) — ставь "дешёвые" и часто различающиеся поля в начало структуры.

---

## Паттерны

### DOD (Data-Oriented Design) — cache-friendly подход

Обычный ООП-стиль хранит массив структур (**AoS — Array of Structs**): каждый элемент — вся структура целиком, лежит рядом в памяти. DOD переворачивает это: хранит структуру массивов (**SoA — Struct of Arrays**) — отдельный массив на каждое поле.

```Go
package main

import (
	"fmt"
	"math/rand"
	"testing"
)

// go test -bench=. comparison_test.go

// OOD-стиль: массив структур (Array of Structs)
type OODStyle struct {
	Field1 int
	Field2 string
	Field3 int
	Field4 string
	// ... остальные поля
}

// DOD-стиль: структура массивов (Struct of Arrays)
type DODStyle struct {
	Field1 []int
	Field2 []string
	Field3 []int
	Field4 []string
	// ... остальные поля тоже срезами
}

var Sink int

func BenchmarkDOD(b *testing.B) {
	data := generateDOD(rand.New(rand.NewSource(42)), 1_000_000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// проходим только по нужному срезу — Field1
		for j, f1 := range data.Field1 {
			if f1 == 500000 {
				Sink = j
			}
		}
	}
}

func BenchmarkOOD(b *testing.B) {
	data := generateOOD(rand.New(rand.NewSource(42)), 1_000_000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// а тут каждый раз тащим ВСЮ структуру ради одного поля
		for j, item := range data {
			if item.Field1 == 500000 {
				Sink = j
			}
		}
	}
}
```

**Почему DOD быстрее:** (этот же принцип используется в [[Maps#What Is a Bucket?|бакетах map]]) в OOD-варианте, чтобы прочитать `Field1` одного элемента, процессор всё равно затягивает в cache line **всю структуру целиком** (включая ненужные сейчас строки Field2, Field4 и т.д.) — это трата cache-пропускной способности. В DOD-варианте `Field1` лежит отдельным непрерывным массивом — только нужные данные, ничего лишнего между ними, максимальная locality.

**Плюсы:** реально ощутимый выигрыш по производительности и по cache-эффективности в hot path. **Минусы:** читаемость и удобство работы с данными заметно хуже — вместо `item.Field1` ты жонглируешь параллельными срезами. Используй, только когда профайлер реально показал, что тут узкое место.

Реальный боевой пример того же принципа — layout групп в новой реализации `map` через [[Swiss Tables#Swiss Table = открытая адресация + группы по 8 слотов|Swiss Table]], где ключи/control word и значения тоже разнесены по отдельным плотным массивам.

### [[Functions#Closure|Closure]] pattern — единая точка очистки ресурсов

Собирать логику закрытия ресурсов (соединения, БД, воркеры) в одном месте не всегда удобно архитектурно. Паттерн: каждый компонент, которому нужно почиститься при завершении, регистрирует свой обработчик в едином "closer'е".

```Go
package main

type Closer struct {
	actions []func()
}

func (c *Closer) Add(action func()) {
	if action == nil { // защита от nil — просто ничего не делаем
		return
	}
	c.actions = append(c.actions, action)
}

func (c *Closer) Close() {
	for _, action := range c.actions {
		action()
	}
}

func main() {
	var closer Closer

	closer.Add(func() {
		// close connections
	})
	closer.Add(func() {
		// close database
	})
	closer.Add(func() {
		// close worker
	})

	// ... код приложения ...

	closer.Close() // закрываем всё разом
}
```

### ⚠️ Подводный камень

Такой самописный Closer — это упрощённая учебная версия. В реальном проекте нужно учитывать порядок закрытия (обычно обратный — LIFO), таймауты, обработку ошибок каждого action-а, конкурентный доступ. Проще взять готовую библиотеку под эту задачу, чем писать это руками с нуля.

### Functional Options Pattern

Классика для конструкторов с кучей опциональных параметров — вместо "телескопических" конструкторов с десятком аргументов. Использует [[Functions#Variadic Parameters|variadic parameters]].

```Go
package main

type User struct {
	Name    string
	Surname string
	Email   string
	Phone   string
	Address string
}

type Option func(*User)

func WithEmail(email string) Option {
	return func(user *User) {
		user.Email = email
	}
}

func WithPhone(phone string) Option {
	return func(user *User) {
		user.Phone = phone
	}
}

func NewUser(name string, surname string, options ...Option) User {
	user := User{Name: name, Surname: surname}

	for _, option := range options {
		option(&user) // применяем каждую опцию по очереди
	}

	return user
}

func main() {
	user1 := NewUser("Ivan", "Ivanov", WithEmail("ivanov@yandex.ru"))
	user2 := NewUser("Petr", "Petrov", WithEmail("petrov@yandex.ru"), WithPhone("+67453"))

	_ = user1
	_ = user2
}
```

**Плюсы:** легко расширять новыми опциями без breaking change в сигнатуре конструктора; вызов читается декларативно. **Минусы:** больше boilerplate-кода (по функции на каждую опцию); поведение чуть менее очевидно, чем прямое присваивание полей.

### Configurable Object Pattern (мини-Builder)

Похож на Functional Options, но через chainable-методы у самого объекта:

```Go
package main

type Logger struct {
	name  string
	level string
}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) WithName(name string) *Logger {
	l.name = name
	return l // возвращаем себя же — чтобы можно было чейнить
}

func (l *Logger) WithLevel(level string) *Logger {
	l.level = level
	return l
}

func main() {
	logger1 := NewLogger()
	logger2 := NewLogger().WithLevel("INFO")
	logger3 := NewLogger().WithName("storage").WithLevel("DEBUG")

	_ = logger1
	_ = logger2
	_ = logger3
}
```

**Плюсы:** приятный, читаемый fluent-синтаксис. **Минусы:** работаем с указателем на мутируемый объект — не потокобезопасно, если параллельно из разных горутин чейнить один и тот же объект.

---

## Type Alias vs Type Definition

В Go можно определить свой тип двумя разными способами, и это два принципиально разных инструмента.

**Type Definition** — создаёт **новый, отдельный тип**:

```Go
type NewTypeName int

type (
	NewTypeName1 string
	NewTypeName2 NewTypeName
)
```

**Type Alias** (обрати внимание на `=`) — просто **псевдоним**, никакого нового типа не создаётся:

```Go
type (
	Name = string
	Age  = int
)

type Table = map[Name]Age
```

Легко перепутать — разница всего в одном символе `=`, а семантика совсем разная:

|                                  | Type Alias (`=`)               | Type Definition (без `=`) |
| -------------------------------- | ------------------------------ | ------------------------- |
| Создаёт новый тип                | Нет, это тот же самый тип      | Да, отдельный тип         |
| Нужен явный cast к базовому типу | Нет                            | Да                        |
| Можно объявлять свои методы      | Нет (это чужой/встроенный тип) | Да                        |

Новый тип из Type Definition имеет тот же **underlying type**, что и исходный, но компилятор считает их разными типами:

```Go
package main

type (
	A = int // alias
	B int   // definition
)

func main() {
	var a A = 1
	var b B = 2

	var ia int = a // ok, A — это и есть int
	_ = ia

	var ib int = int(b) // а тут нужен явный int(b), иначе compile error
	_ = ib
}
```

### ⚠️ Подводный камень: методы и alias не дружат

```Go
package main

import "fmt"

type Data struct {
	values int
}

func (d Data) Print() {
	fmt.Println("data")
}

type (
	DataType  Data // definition — НОВЫЙ тип
	DataAlias = Data // alias — это буквально Data
)

func main() {
	data1 := DataType{}
	_ = data1.values // ok, поля скопировались (тот же underlying type)
	data1.Print()    // ❌ compile error — методы Data к DataType не переходят!

	data2 := DataAlias{}
	data2.Print() // ok, DataAlias — это и есть Data, со всеми её методами
}
```

Важный нюанс: методы к базовым/чужим/указательным/interface-типам добавить нельзя в принципе — только к своим named-типам, определённым в твоём пакете через Type Definition:

```Go
package main

import "io"

type (
	IntAlias = int
	IntType  int
)

func (i int) Do()      {} // ❌ compile error: basic type
func (i IntAlias) Do() {} // ❌ compile error: тот же самый basic type
func (i IntType) Do()  {} // ✅ ok — это уже твой собственный тип

type (
	IntPtrAlias = *int
	IntPtrType  *int
)

func (i IntPtrAlias) Do() {} // ❌ compile error: pointer type
func (i IntPtrType) Do()  {} // ❌ compile error: тоже pointer, тут не спасает

type (
	CloserAlias = io.Closer
	CloserType  io.Closer
)

func (i io.Closer) Do()   {} // ❌ compile error: interface
func (i CloserAlias) Do() {} // ❌ compile error: interface
func (i CloserType) Do()  {} // ❌ compile error: interface (даже через definition нельзя!)
```

Зато Type Definition отлично работает, чтобы навесить своё поведение на примитивы/map/func:

```Go
package main

type Age int

func (age Age) LargerThan(other Age) bool {
	return age > other
}

type FilterFunc func(int) bool

func (ff FilterFunc) Filter(value int) bool {
	return ff(value)
}

type StringSet map[string]struct{}

// Использует [[Maps|map]] под капотом
func (ss StringSet) Has(key string) bool {
	_, found := ss[key]
	return found
}

func (ss StringSet) Add(key string) {
	ss[key] = struct{}{}
}

func (ss StringSet) Remove(key string) {
	delete(ss, key)
}
```

### Ускоряем cast через unsafe

Если Type Definition имеет тот же underlying type и тот же memory layout, что и базовый тип — можно обойти цикл явного преобразования через `unsafe`, реинтерпретировав память напрямую:

```Go
package main

import (
	"testing"
	"unsafe"
)

// go test -bench=. -benchmem performance_test.go

type Int int

var convertedData []int

func BenchmarkCast(b *testing.B) {
	data := make([]Int, 1024)
	for i := 0; i < b.N; i++ {
		convertedData = make([]int, 1024)
		for idx, value := range data {
			convertedData[idx] = int(value) // конвертируем каждый элемент вручную
		}
	}
}

func BenchmarkUnsafeCast(b *testing.B) {
	data := make([]Int, 1024)
	for i := 0; i < b.N; i++ {
		// просто переинтерпретируем память []Int как []int — без копирования!
		convertedData = *(*[]int)(unsafe.Pointer(&data))
	}
}
```

Разница в бенчмарке будет ощутимая — `unsafe`-версия не копирует и не проходит циклом по элементам.

### ⚠️ Подводный камень

Это работает **только** потому, что `Int` и `int` полностью идентичны по размеру и представлению в памяти (underlying type совпадает один в один). Если попробовать так же реинтерпретировать типы с разным layout-ом — сразу получишь memory corruption. Используй этот трюк только в горячем пути, когда точно уверен в идентичности layout-ов, и желательно с комментарием почему это безопасно именно тут.